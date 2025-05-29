require("http-proxy")
  .createServer({
    target: "http://127.0.0.1:9222",
    localAddress: "0.0.0.0",
    ws: true,
  })
  .listen(9221)
