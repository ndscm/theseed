#include "seed/infra/grpc/grpcbeast/beast_server.h"

#include <algorithm>
#include <cstdio>
#include <cstdlib>
#include <functional>
#include <iostream>
#include <memory>
#include <string>
#include <thread>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "boost/asio/detached.hpp"
#include "boost/asio/dispatch.hpp"
#include "boost/asio/ip/tcp.hpp"
#include "boost/asio/spawn.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"
#include "boost/beast/version.hpp"
#include "boost/config.hpp"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::boost::string_view;
using ::std::size_t;
using ::std::string;

::boost::beast::http::status httpStatus(::absl::Status status) {
  if (status.ok()) {
    return ::boost::beast::http::status::ok;
  } else if (::absl::IsInternal(status)) {
    return ::boost::beast::http::status::internal_server_error;
  } else if (::absl::IsInvalidArgument(status)) {
    return ::boost::beast::http::status::bad_request;
  } else if (::absl::IsNotFound(status)) {
    return ::boost::beast::http::status::not_found;
  } else if (::absl::IsPermissionDenied(status)) {
    return ::boost::beast::http::status::forbidden;
  } else if (::absl::IsUnauthenticated(status)) {
    return ::boost::beast::http::status::unauthorized;
  } else if (::absl::IsUnavailable(status)) {
    return ::boost::beast::http::status::service_unavailable;
  }
  return ::boost::beast::http::status::unknown;
}

static bool asyncHandleRequest(
    ::boost::beast::tcp_stream& stream, ::std::function<BeastRouter> router,
    ::boost::beast::http::request<::boost::beast::http::string_body>&& request,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  if (request.method() == ::boost::beast::http::verb::options) {
    ::boost::beast::http::response<::boost::beast::http::empty_body> response(
        ::boost::beast::http::status::ok, request.version());
    response.set(::boost::beast::http::field::access_control_allow_origin, "*");
    response.set(::boost::beast::http::field::access_control_allow_headers,
                 "Origin, Content-Type, X-Grpc-Web, X-User-Agent");
    response.set(::boost::beast::http::field::access_control_max_age, "600");
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
  ::absl::StatusOr<bool> keepAlive = router(stream, request, yield);
  if (!keepAlive.status().ok()) {
    ::std::cerr << "ERROR: " << keepAlive.status().message() << "\n";
    return false;
  }
  return keepAlive.value();
}

::absl::Status BeastListener::listen(
    ::boost::asio::ip::tcp::endpoint endpoint) {
  ::boost::beast::error_code err;
  this->tcpAcceptor.open(endpoint.protocol(), err);
  if (err) {
    return ::absl::InternalError("open: " + err.message());
  }
  this->tcpAcceptor.set_option(::boost::asio::socket_base::reuse_address(true),
                               err);
  if (err) {
    return ::absl::InternalError("set_option: " + err.message());
  }
  this->tcpAcceptor.bind(endpoint, err);
  if (err) {
    return ::absl::InternalError("bind: " + err.message());
  }
  this->tcpAcceptor.listen(::boost::asio::socket_base::max_listen_connections,
                           err);
  if (err) {
    return ::absl::InternalError("listen: " + err.message());
  }
  return ::absl::OkStatus();
}

static void doSession(::boost::beast::tcp_stream& stream,
                      ::std::function<BeastRouter> router,
                      ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::flat_buffer buffer;
  while (true) {
    stream.expires_after(::std::chrono::seconds(305));
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
  stream.socket().shutdown(::boost::asio::ip::tcp::socket::shutdown_send, err);
  // At this point the connection is closed gracefully
}

static void doAccept(::boost::asio::io_context& asioContext,
                     ::boost::asio::ip::tcp::acceptor& tcpAcceptor,
                     ::std::function<BeastRouter> router,
                     ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  while (true) {
    ::boost::asio::ip::tcp::socket socket(asioContext);
    tcpAcceptor.async_accept(socket, yield[err]);
    if (err) {
      ::std::cerr << "accept: " + err.message() << "\n";
      continue;
    }
    ::boost::asio::spawn(  //
        tcpAcceptor.get_executor(),
        ::std::bind(&doSession, ::boost::beast::tcp_stream(std::move(socket)),
                    router, std::placeholders::_1),
        // Ignore the result of the session
        ::boost::asio::detached);
  }
}

void BeastListener::asyncStart() {
  ::boost::asio::spawn(  //
      this->asioContext,
      ::std::bind(&doAccept, ::std::ref(this->asioContext),
                  ::std::ref(this->tcpAcceptor), this->router,
                  ::std::placeholders::_1),
      [](::std::exception_ptr ex) {
        // this will cause `ioc.run()` to throw
        if (ex) {
          std::rethrow_exception(ex);
        }
      });
}

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed
