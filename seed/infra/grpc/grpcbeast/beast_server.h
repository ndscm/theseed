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

typedef ::absl::StatusOr<bool> BeastRouter(
    ::boost::beast::tcp_stream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield);

class BeastListener : public ::std::enable_shared_from_this<BeastListener> {
  ::boost::asio::io_context& asioContext;
  ::boost::asio::ip::tcp::acceptor tcpAcceptor;
  ::std::function<BeastRouter> router;

 public:
  BeastListener(::boost::asio::io_context& asioContext,
                ::std::function<BeastRouter> router)
      : asioContext(asioContext),
        tcpAcceptor(::boost::asio::make_strand(asioContext)),
        router(router) {}
  ::absl::Status listen(::boost::asio::ip::tcp::endpoint endpoint);
  void asyncStart();
};

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif
