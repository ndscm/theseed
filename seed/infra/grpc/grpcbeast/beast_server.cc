#include "seed/infra/grpc/grpcbeast/beast_server.h"

#include <openssl/ssl.h>

#include <iostream>
#include <memory>
#include <string>

#include "absl/flags/flag.h"
#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "boost/asio/detached.hpp"
#include "boost/asio/ip/tcp.hpp"
#include "boost/asio/spawn.hpp"
#include "boost/asio/ssl.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"
#include "boost/beast/version.hpp"
#include "boost/config.hpp"
#include "seed/infra/grpc/grpcbeast/cors.h"

ABSL_FLAG(::std::string, https_certificate_file, "", "");
ABSL_FLAG(::std::string, https_certificate_key_file, "", "");

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;

template <typename HttpStream>
::absl::Status BeastListener<HttpStream>::initSsl(
    const string& certificate_path, const string& certificate_key_path) {
  string certificate_path_ =
      certificate_path != "" ? certificate_path
                             : ::absl::GetFlag(FLAGS_https_certificate_file);
  string certificate_key_path_ =
      certificate_key_path != ""
          ? certificate_key_path
          : ::absl::GetFlag(FLAGS_https_certificate_key_file);
  this->ssl_context_ = ::std::make_shared<::boost::asio::ssl::context>(
      ::boost::asio::ssl::context::tls_server);
  this->ssl_context_->set_options(
      ::boost::asio::ssl::context::default_workarounds |
      ::boost::asio::ssl::context::no_sslv2 |
      ::boost::asio::ssl::context::no_sslv3 |
      ::boost::asio::ssl::context::no_tlsv1 |
      ::boost::asio::ssl::context::no_tlsv1_1 |
      ::boost::asio::ssl::context::single_dh_use);

  // Set up ALPN to advertise HTTP/1.1 support
  SSL_CTX_set_alpn_select_cb(
      this->ssl_context_->native_handle(),
      [](SSL* /*ssl*/, const unsigned char** out, unsigned char* outlen,
         const unsigned char* in, unsigned int inlen, void* /*arg*/) -> int {
        // Advertise HTTP/1.1
        static const unsigned char alpn[] = {8,   'h', 't', 't', 'p',
                                             '/', '1', '.', '1'};
        if (SSL_select_next_proto(const_cast<unsigned char**>(out), outlen,
                                  alpn, sizeof(alpn), in,
                                  inlen) != OPENSSL_NPN_NEGOTIATED) {
          return SSL_TLSEXT_ERR_NOACK;
        }
        return SSL_TLSEXT_ERR_OK;
      },
      nullptr);
  ::boost::system::error_code err;
  this->ssl_context_->use_certificate_chain_file(certificate_path_, err);
  if (err) {
    return ::absl::InternalError("Failed to load certificate: " +
                                 err.message());
  }
  this->ssl_context_->use_private_key_file(
      certificate_key_path_, ::boost::asio::ssl::context::pem, err);
  if (err) {
    return ::absl::InternalError("Failed to load certificate key: " +
                                 err.message());
  }
  return ::absl::OkStatus();
}

// Generic request handler that works with any stream type
template <typename HttpStream>
static bool asyncHandleRequest(
    HttpStream& stream, ::std::function<BeastRouter<HttpStream>> router,
    ::boost::beast::http::request<::boost::beast::http::string_body>&& request,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  if (request.method() == ::boost::beast::http::verb::options) {
    ::boost::beast::http::response<::boost::beast::http::empty_body> response(
        ::boost::beast::http::status::ok, request.version());
    SetCors(response, request);
    response.content_length(0);
    response.keep_alive(request.keep_alive());
    ::boost::beast::http::async_write(stream, std::move(response), yield[err]);
    if (err) {
      ::std::cerr << "Failed to write OPTIONS: " << err.message() << "\n";
    }
    return response.keep_alive();
  }

  if (request.method() != ::boost::beast::http::verb::get &&
      request.method() != ::boost::beast::http::verb::post) {
    ::boost::beast::http::response<::boost::beast::http::string_body> response(
        ::boost::beast::http::status::bad_request, request.version());
    response.set(::boost::beast::http::field::content_type, "text/plain");
    response.body() = "Unknown HTTP-method";
    response.content_length(response.body().length());
    response.keep_alive(request.keep_alive());
    ::boost::beast::http::async_write(stream, std::move(response), yield[err]);
    if (err) {
      ::std::cerr << "Failed to write UNKNOWN: " << err.message() << "\n";
    }
    return response.keep_alive();
  }
  ::absl::StatusOr<bool> keep_alive = router(stream, request, yield);
  if (!keep_alive.status().ok()) {
    ::std::cerr << "ERROR: " << keep_alive.status().message() << "\n";
    return false;
  }
  return keep_alive.value();
}

template <typename HttpStream>
::absl::Status BeastListener<HttpStream>::listen(
    ::boost::asio::ip::tcp::endpoint endpoint) {
  ::boost::beast::error_code err;
  tcp_acceptor_.open(endpoint.protocol(), err);
  if (err) {
    return ::absl::InternalError("open: " + err.message());
  }
  tcp_acceptor_.set_option(::boost::asio::socket_base::reuse_address(true),
                           err);
  if (err) {
    return ::absl::InternalError("set_option: " + err.message());
  }
  tcp_acceptor_.bind(endpoint, err);
  if (err) {
    return ::absl::InternalError("bind: " + err.message());
  }
  tcp_acceptor_.listen(::boost::asio::socket_base::max_listen_connections, err);
  if (err) {
    return ::absl::InternalError("listen: " + err.message());
  }
  return ::absl::OkStatus();
}

static void setStreamTimeout(::boost::beast::tcp_stream& stream) {
  stream.expires_after(::std::chrono::seconds(305));
}

static void setStreamTimeout(
    ::boost::asio::ssl::stream<::boost::beast::tcp_stream>& stream) {
  ::boost::beast::get_lowest_layer(stream).expires_after(
      ::std::chrono::seconds(305));
}

static void shutdownStream(::boost::beast::tcp_stream& stream,
                           ::boost::asio::yield_context /*yield*/) {
  ::boost::beast::error_code err;
  stream.socket().shutdown(::boost::asio::ip::tcp::socket::shutdown_send, err);
}

static void shutdownStream(
    ::boost::asio::ssl::stream<::boost::beast::tcp_stream>& stream,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  stream.async_shutdown(yield[err]);
  if (err && err != ::boost::asio::error::eof &&
      err != ::boost::beast::error::timeout) {
    ::std::cerr << "SSL shutdown failed: " << err.message() << "\n";
  }
}

// Generic session handler that works with any stream type
template <typename HttpStream>
static void doSession(HttpStream& stream,
                      ::std::function<BeastRouter<HttpStream>> router,
                      ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::flat_buffer buffer;
  while (true) {
    setStreamTimeout(stream);
    ::boost::beast::http::request<::boost::beast::http::string_body> request;
    ::boost::beast::http::async_read(stream, buffer, request, yield[err]);
    if (err == ::boost::beast::http::error::end_of_stream) {
      break;
    }
    if (err) {
      ::std::cerr << "Failed to read: " << err.message() << "\n";
      return;
    }
    bool keep_alive =
        asyncHandleRequest(stream, router, ::std::move(request), yield);
    if (!keep_alive) {
      // This means we should close the connection, usually because
      // the response indicated the "Connection: close" semantic.
      break;
    }
  }
}

// doAccept for tcp_stream (HTTP)
static void doAccept(
    ::boost::asio::io_context& asio_context,
    ::boost::asio::ip::tcp::acceptor& tcp_acceptor,
    ::std::shared_ptr<::boost::asio::ssl::context> /*ssl_context*/,
    ::std::function<BeastRouter<::boost::beast::tcp_stream>> router,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  while (true) {
    ::boost::asio::ip::tcp::socket socket(asio_context);
    tcp_acceptor.async_accept(socket, yield[err]);
    if (err) {
      ::std::cerr << "accept: " << err.message() << "\n";
      continue;
    }
    ::boost::asio::spawn(
        tcp_acceptor.get_executor(),
        [socket = std::move(socket),
         router](::boost::asio::yield_context yield) mutable {
          ::boost::beast::tcp_stream stream(std::move(socket));
          doSession(stream, router, yield);
          shutdownStream(stream, yield);
        },
        ::boost::asio::detached);
  }
}

// doAccept for ssl::stream<tcp_stream> (HTTPS)
static void doAccept(
    ::boost::asio::io_context& asio_context,
    ::boost::asio::ip::tcp::acceptor& tcp_acceptor,
    ::std::shared_ptr<::boost::asio::ssl::context> ssl_context,
    ::std::function<
        BeastRouter<::boost::asio::ssl::stream<::boost::beast::tcp_stream>>>
        router,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  while (true) {
    ::boost::asio::ip::tcp::socket socket(asio_context);
    tcp_acceptor.async_accept(socket, yield[err]);
    if (err) {
      ::std::cerr << "accept: " << err.message() << "\n";
      continue;
    }
    ::boost::asio::spawn(
        tcp_acceptor.get_executor(),
        [socket = std::move(socket), ssl_context,
         router](::boost::asio::yield_context yield) mutable {
          ::boost::beast::tcp_stream tcp_stream(std::move(socket));
          ::boost::asio::ssl::stream<::boost::beast::tcp_stream> stream(
              std::move(tcp_stream), *ssl_context);

          // Perform SSL handshake
          ::boost::beast::error_code err;
          stream.async_handshake(::boost::asio::ssl::stream_base::server,
                                 yield[err]);
          if (err) {
            ::std::cerr << "SSL handshake failed: " << err.message() << "\n";
            return;
          }
          doSession(stream, router, yield);
          shutdownStream(stream, yield);
        },
        ::boost::asio::detached);
  }
}

template <typename HttpStream>
void BeastListener<HttpStream>::asyncStart() {
  ::boost::asio::spawn(
      asio_context_,
      [this](::boost::asio::yield_context yield) {
        doAccept(asio_context_, tcp_acceptor_, ssl_context_, router_, yield);
      },
      [](::std::exception_ptr ex) {
        if (ex) {
          ::std::rethrow_exception(ex);
        }
      });
}

template class BeastListener<::boost::beast::tcp_stream>;
template class BeastListener<
    ::boost::asio::ssl::stream<::boost::beast::tcp_stream>>;

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed
