var proxy = "DIRECT"

var whitelist = {

































  go: 1,
}

function FindProxyForURL(url, host) {
  var dot
  do {
    if (whitelist.hasOwnProperty(host)) {
      return "DIRECT"
    }
    dot = host.indexOf(".")
    host = host.slice(dot + 1)
  } while (dot >= 0)
  return proxy
}
