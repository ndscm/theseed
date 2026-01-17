#ifndef SEED_INFRA_GRPC_GRPCBEAST_GRPCBEAST_H
#define SEED_INFRA_GRPC_GRPCBEAST_GRPCBEAST_H

#include <string>

#include "absl/status/status.h"
#include "boost/asio/dispatch.hpp"
#include "boost/asio/error.hpp"
#include "boost/asio/spawn.hpp"
#include "boost/asio/strand.hpp"
#include "boost/beast/core.hpp"
#include "boost/beast/http.hpp"
#include "grpcpp/grpcpp.h"
#include "seed/infra/grpc/grpcbeast/cors.h"

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;

namespace internal {

string wrapGrpcWebChunk(uint8_t chunkType, string content);

template <typename Message>
string wrapGrpcWebMessage(Message message_pb) {
  string message_bytes = message_pb.SerializeAsString();
  return wrapGrpcWebChunk(0, message_bytes);
}

string wrapGrpcWebTrailer(::grpc::Status status);

template <typename Reply>
string wrapGrpcWebReply(::grpc::Status status, Reply reply_pb) {
  return wrapGrpcWebMessage(reply_pb) + wrapGrpcWebTrailer(status);
}

}  // namespace internal

::absl::Status ParseGrpcWebRequestPath(
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    string& out_service, string& out_method);

template <typename Request>
::absl::Status ParseGrpcWebRequest(const string request_body,
                                   Request& out_request_pb) {
  bool ok = out_request_pb.ParseFromString(request_body.substr(5));
  if (!ok) {
    return ::absl::InvalidArgumentError("Failed to parse Request");
  }
  return ::absl::OkStatus();
}

template <typename HttpStream, typename Reply>
bool WriteGrpcWebResponse(
    HttpStream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::grpc::Status status, Reply reply_pb, ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::http::response<::boost::beast::http::string_body> response(
      ::boost::beast::http::status::ok, request.version());
  SetCors(response, request);
  response.set(::boost::beast::http::field::content_type,
               "application/grpc-web+proto");
  if (!status.ok()) {
    response.set("Message", status.error_message());
  }
  response.body() = internal::wrapGrpcWebReply(status, reply_pb);
  response.content_length(response.body().length());
  response.keep_alive(request.keep_alive());
  ::boost::beast::http::async_write(stream, std::move(response), yield[err]);
  if (err) {
    ::std::cerr << "Failed to write reply: " << err.message();
  }
  return response.keep_alive();
}

template <typename HttpStream>
bool WriteGrpcWebStreamHeader(
    HttpStream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::http::response<::boost::beast::http::empty_body> response(
      ::boost::beast::http::status::ok, request.version());
  SetCors(response, request);
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

template <typename HttpStream, typename Reply>
void WriteGrpcWebStreamReply(HttpStream& stream, Reply& reply_pb,
                             ::boost::asio::yield_context yield) {
  string chunk = internal::wrapGrpcWebMessage(reply_pb);
  ::boost::asio::async_write(stream,
                             ::boost::beast::http::make_chunk(
                                 ::boost::asio::buffer(chunk, chunk.size())),
                             yield);
}

template <typename HttpStream>
void WriteGrpcWebStreamTrailer(HttpStream& stream, ::grpc::Status reply_status,
                               ::boost::asio::yield_context yield) {
  string chunk = internal::wrapGrpcWebTrailer(reply_status);
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

#endif
