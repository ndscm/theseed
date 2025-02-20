var fs = require("node:fs")
var http = require("node:http")
var process = require("node:process")

var server = http.createServer(function (request, response) {
  if (request.url != "/wpad.dat" && request.url != "/proxy.pac") {
    console.error("Invalid url: ", request.url)
  }
  console.info("Request")
  fs.readFile("./proxy.pac", { encoding: "utf-8" }, function (error, content) {
    if (error) {
      if (error.code == "ENOENT") {
        response.writeHead(404)
        response.end("404", "utf-8")
      } else {
        response.writeHead(500)
        response.end("Error: " + error.code, "utf-8")
      }
    } else {
      content = content.replace(
        /^var[ ]proxy[ ][=].*/,
        'var proxy = "' + process.env["WPAD_TARGET"] + '";',
      )
      response.writeHead(200, {
        "Content-Type": "application/x-ns-proxy-autoconfig",
      })
      response.end(content, "utf-8")
    }
    console.info("Response")
  })
})

server.listen(9723, "0.0.0.0")
console.info("Wpad server started")
