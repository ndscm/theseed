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
    string& outService, string& outMethod) {
  string path = request.target();
  vector<::absl::string_view> parts = ::absl::StrSplit(path, '/');
  if (parts.size() != 3) {
    return ::absl::InvalidArgumentError("Wrong request path format: " + path);
  }
  outService = parts[1];
  outMethod = parts[2];
  return ::absl::OkStatus();
}

bool WriteGrpcWebStreamHeader(
    ::boost::beast::tcp_stream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::http::response<::boost::beast::http::empty_body> response(
      ::boost::beast::http::status::ok, request.version());
  response.set(::boost::beast::http::field::access_control_allow_origin, "*");
  response.set(::boost::beast::http::field::content_type,
               "application/grpc-web+proto");
  response.content_length(0);
  response.chunked(true);
  response.keep_alive(request.keep_alive());
  ::boost::beast::http::response_serializer<::boost::beast::http::empty_body>
      serializer(response);
  ::boost::beast::http::async_write_header(stream, serializer, yield[err]);
  if (err) {
    ::std::cerr << "Failed to write header: " << err.message();
  }
  return response.keep_alive();
}

void WriteGrpcWebStreamTrailer(::boost::beast::tcp_stream& stream,
                               ::grpc::Status replyStatus,
                               ::boost::asio::yield_context yield) {
  string chunk = internal::wrapGrpcWebTrailer(replyStatus);
  ::boost::asio::async_write(stream,
                             ::boost::beast::http::make_chunk(
                                 ::boost::asio::buffer(chunk, chunk.size())),
                             yield);
  ::boost::asio::async_write(stream, ::boost::beast::http::make_chunk_last(),
                             yield);
}

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed
