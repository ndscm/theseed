var proxy = "DIRECT"

var whitelist = {










































  go: 1,
}

function FindProxyForURL(url, host) {
  var hostip = dnsResolve(host)
  if (
    isInNet(hostip, "127.0.0.0", "255.255.255.0") ||
    isInNet(hostip, "192.168.0.0", "255.255.0.0") ||
    isInNet(hostip, "172.16.0.0", "255.240.0.0") ||
    isInNet(hostip, "10.0.0.0", "255.0.0.0")
  ) {
    return "DIRECT"
  }
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
