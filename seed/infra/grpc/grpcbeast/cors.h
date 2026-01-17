#ifndef SEED_INFRA_GRPC_GRPCBEAST_CORS_H_
#define SEED_INFRA_GRPC_GRPCBEAST_CORS_H_

#include <string>

#include "absl/status/status.h"
#include "boost/beast/http.hpp"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

::absl::Status CheckCorsOrigin(const ::std::string& origin);
::std::string GetCorsMethods();

template <typename ResponseBody, typename RequestBody>
void SetCors(::boost::beast::http::response<ResponseBody>& response,
             const ::boost::beast::http::request<RequestBody>& request) {
  auto origin = request[::boost::beast::http::field::origin];
  auto status = CheckCorsOrigin(::std::string(origin));
  if (!status.ok()) {
    return;
  }
  response.set(::boost::beast::http::field::access_control_allow_origin,
               origin);
  response.set(::boost::beast::http::field::access_control_allow_methods,
               GetCorsMethods());
  response.set(::boost::beast::http::field::access_control_allow_headers,
               "Origin, Content-Type, X-Grpc-Web, X-User-Agent");
  response.set(::boost::beast::http::field::access_control_allow_credentials,
               "true");
  response.set(::boost::beast::http::field::access_control_max_age, "600");
}

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_GRPC_GRPCBEAST_CORS_H_
