#include "seed/infra/grpc/grpcbeast/cors.h"

#include <string>
#include <vector>

#include "absl/flags/flag.h"
#include "absl/strings/str_split.h"

ABSL_FLAG(::std::string, cors_origins, "",
          "allowed CORS origins, comma separated");
ABSL_FLAG(::std::string, cors_methods, "GET,POST,PUT,DELETE,OPTIONS",
          "allowed CORS methods, comma separated");

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;
using ::std::vector;

::absl::Status CheckCorsOrigin(const string& origin) {
  if (::absl::GetFlag(FLAGS_cors_origins) == "*") {
    return ::absl::OkStatus();
  }
  if (origin.empty()) {
    return ::absl::InvalidArgumentError("Origin is empty");
  }
  vector<string> allowed_origins =
      ::absl::StrSplit(::absl::GetFlag(FLAGS_cors_origins), ',');
  for (const auto& allowed : allowed_origins) {
    if (origin == allowed) {
      return ::absl::OkStatus();
    }
  }
  return ::absl::InvalidArgumentError("Origin not allowed");
}

string GetCorsMethods() { return ::absl::GetFlag(FLAGS_cors_methods); }

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed
