#ifndef SEED_INFRA_GRPC_GRPCBEAST_BEAST_SERVER_H
#define SEED_INFRA_GRPC_GRPCBEAST_BEAST_SERVER_H

#include <string>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "boost/asio/dispatch.hpp"
#include "boost/asio/error.hpp"
#include "boost/asio/spawn.hpp"
#include "boost/asio/ssl.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;

// Generic router function template that works with any stream type
template <typename HttpStream>
using BeastRouter = ::absl::StatusOr<bool>(
    HttpStream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield);

template <typename HttpStream>
class BeastListener
    : public ::std::enable_shared_from_this<BeastListener<HttpStream>> {
  ::boost::asio::io_context& asio_context_;
  ::boost::asio::ip::tcp::acceptor tcp_acceptor_;
  ::std::function<BeastRouter<HttpStream>> router_;
  ::std::shared_ptr<::boost::asio::ssl::context> ssl_context_;

 public:
  BeastListener(::boost::asio::io_context& asio_context,
                ::std::function<BeastRouter<HttpStream>> router)
      : asio_context_(asio_context),
        tcp_acceptor_(::boost::asio::make_strand(asio_context_)),
        router_(router),
        ssl_context_(nullptr) {}

  ::absl::Status initSsl(const string& certificate_path = "",
                         const string& certificate_key_path = "");

  ::absl::Status listen(::boost::asio::ip::tcp::endpoint endpoint);
  void asyncStart();
};

extern template class BeastListener<::boost::beast::tcp_stream>;
extern template class BeastListener<
    ::boost::asio::ssl::stream<::boost::beast::tcp_stream>>;

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif
