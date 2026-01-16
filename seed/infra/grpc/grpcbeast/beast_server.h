#ifndef SEED_INFRA_GRPC_GRPCBEAST_BEAST_SERVER_H
#define SEED_INFRA_GRPC_GRPCBEAST_BEAST_SERVER_H

#include <string>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "boost/asio/dispatch.hpp"
#include "boost/asio/error.hpp"
#include "boost/asio/spawn.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;

using BeastRouter = ::absl::StatusOr<bool>(
    ::boost::beast::tcp_stream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield);

class BeastListener : public ::std::enable_shared_from_this<BeastListener> {
  ::boost::asio::io_context& asio_context_;
  ::boost::asio::ip::tcp::acceptor tcp_acceptor_;
  ::std::function<BeastRouter> router_;

 public:
  BeastListener(::boost::asio::io_context& asio_context,
                ::std::function<BeastRouter> router)
      : asio_context_(asio_context),
        tcp_acceptor_(::boost::asio::make_strand(asio_context)),
        router_(router) {}
  ::absl::Status listen(::boost::asio::ip::tcp::endpoint endpoint);
  void asyncStart();
};

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif
