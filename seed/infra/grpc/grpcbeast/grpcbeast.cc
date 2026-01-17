#include "seed/infra/grpc/grpcbeast/grpcbeast.h"

#include <string>
#include <vector>

#include "absl/status/status.h"
#include "absl/strings/str_split.h"
#include "boost/asio/dispatch.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"
#include "boost/beast/version.hpp"
#include "boost/config.hpp"
#include "grpcpp/grpcpp.h"
#include "seed/infra/grpc/grpcbeast/cors.h"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;
using ::std::vector;

namespace internal {

string wrapGrpcWebChunk(uint8_t chunkType, string content) {
  uint32_t contentLength = content.length();
  uint8_t chunkHead[5];
  chunkHead[0] = chunkType;
  chunkHead[1] = contentLength >> 24;
  chunkHead[2] = contentLength >> 16;
  chunkHead[3] = contentLength >> 8;
  chunkHead[4] = contentLength >> 0;
  return string((char*)chunkHead, 5) + content;
}

string wrapGrpcWebTrailer(::grpc::Status status) {
  string trailer = "grpc-status: " + ::std::to_string(status.error_code()) +
                   "\r\ngrpc-message: " + status.error_message();
  return wrapGrpcWebChunk(0x80, trailer);
}

}  // namespace internal

::absl::Status ParseGrpcWebRequestPath(
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    string& out_service, string& out_method) {
  string path = request.target();
  vector<::absl::string_view> parts = ::absl::StrSplit(path, '/');
  if (parts.size() != 3) {
    return ::absl::InvalidArgumentError("Wrong request path format: " + path);
  }
  out_service = parts[1];
  out_method = parts[2];
  return ::absl::OkStatus();
}

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed
