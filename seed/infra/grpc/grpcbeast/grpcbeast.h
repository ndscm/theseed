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

namespace seed {
namespace infra {
namespace grpc {
namespace grpcbeast {

using ::std::string;

namespace internal {

string wrapGrpcWebChunk(uint8_t chunkType, string content);

template <class Message>
string wrapGrpcWebMessage(Message messagePb) {
  string messageBytes = messagePb.SerializeAsString();
  return wrapGrpcWebChunk(0, messageBytes);
}

string wrapGrpcWebTrailer(::grpc::Status status);

template <class Reply>
string wrapGrpcWebReply(::grpc::Status status, Reply replyPb) {
  return wrapGrpcWebMessage(replyPb) + wrapGrpcWebTrailer(status);
}

}  // namespace internal

::absl::Status ParseGrpcWebRequestPath(
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    string& outService, string& outMethod);

template <class Request>
::absl::Status ParseGrpcWebRequest(const string requestBody,
                                   Request& outRequestPb) {
  bool ok = outRequestPb.ParseFromString(requestBody.substr(5));
  if (!ok) {
    return ::absl::InvalidArgumentError("Failed to parse Request");
  }
  return ::absl::OkStatus();
}

template <class Reply>
bool WriteGrpcWebResponse(
    ::boost::beast::tcp_stream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::grpc::Status status, Reply replyPb, ::boost::asio::yield_context yield) {
  ::boost::beast::error_code err;
  ::boost::beast::http::response<::boost::beast::http::string_body> response(
      ::boost::beast::http::status::ok, request.version());
  response.set(::boost::beast::http::field::access_control_allow_origin, "*");
  response.set(::boost::beast::http::field::content_type,
               "application/grpc-web+proto");
  if (!status.ok()) {
    response.set("Message", status.error_message());
  }
  response.body() = internal::wrapGrpcWebReply(status, replyPb);
  response.content_length(response.body().length());
  response.keep_alive(request.keep_alive());
  ::boost::beast::http::async_write(stream, std::move(response), yield[err]);
  if (err) {
    ::std::cerr << "Failed to write reply: " << err.message();
  }
  return response.keep_alive();
}

bool WriteGrpcWebStreamHeader(
    ::boost::beast::tcp_stream& stream,
    ::boost::beast::http::request<::boost::beast::http::string_body>& request,
    ::boost::asio::yield_context yield);

template <class Reply>
void WriteGrpcWebStreamReply(::boost::beast::tcp_stream& stream, Reply& replyPb,
                             ::boost::asio::yield_context yield) {
  string chunk = internal::wrapGrpcWebMessage(replyPb);
  ::boost::asio::async_write(stream,
                             ::boost::beast::http::make_chunk(
                                 ::boost::asio::buffer(chunk, chunk.size())),
                             yield);
}

void WriteGrpcWebStreamTrailer(::boost::beast::tcp_stream& stream,
                               ::grpc::Status replyStatus,
                               ::boost::asio::yield_context yield);

}  // namespace grpcbeast
}  // namespace grpc
}  // namespace infra
}  // namespace seed

#endif
