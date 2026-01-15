#ifndef SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H
#define SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H

#include <termios.h>

#include <cstdint>
#include <string>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace seed {
namespace infra {
namespace serialport {
namespace cc {

using ::std::string;
using ::std::vector;

class SerialPort {
 public:
  SerialPort(const string& port = "/dev/ttyUSB0", speed_t baud_rate = B57600)
      : port_(port), baud_rate_(baud_rate), fd_(-1) {}
  ~SerialPort() { (void)close(); }

  ::absl::Status open();
  ::absl::Status close();
  ::absl::Status check_open() const;

  ::absl::Status write(const vector<uint8_t>& data);
  ::absl::StatusOr<vector<uint8_t>> read_full(size_t full_size,
                                              int timeout_ms = 1000);
  ::absl::StatusOr<int> available();

  ::absl::Status flush();

  const string& port() const { return port_; }
  speed_t baud_rate() const { return baud_rate_; }

 private:
  string port_;
  speed_t baud_rate_;
  int fd_;
};

}  // namespace cc
}  // namespace serialport
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H
