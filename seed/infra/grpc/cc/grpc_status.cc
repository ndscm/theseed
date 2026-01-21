#include "seed/infra/grpc/cc/grpc_status.h"

#include <string>

#include "absl/status/status.h"
#include "grpcpp/support/status.h"

namespace seed {
namespace infra {
namespace grpc {
namespace cc {

using ::std::string;

::grpc::Status ToGrpcStatus(::absl::Status status) {
  return ::grpc::Status(::grpc::StatusCode(status.code()),
                        string(status.message()));
}

::absl::Status ToAbslStatus(::grpc::Status status) {
  return ::absl::Status(static_cast<::absl::StatusCode>(status.error_code()),
                        status.error_message());
}

}  // namespace cc
}  // namespace grpc
}  // namespace infra
}  // namespace seed
