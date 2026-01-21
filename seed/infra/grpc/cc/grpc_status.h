#ifndef SEED_INFRA_GRPC_CC_GRPC_STATUS_H
#define SEED_INFRA_GRPC_CC_GRPC_STATUS_H

#include "absl/status/status.h"
#include "grpcpp/support/status.h"

namespace seed {
namespace infra {
namespace grpc {
namespace cc {

::grpc::Status ToGrpcStatus(::absl::Status status);

::absl::Status ToAbslStatus(::grpc::Status status);

}  // namespace cc
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_GRPC_CC_GRPC_STATUS_H
