#ifndef SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H
#define SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H

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
  SerialPort(const string& port = "/dev/ttyUSB0") : port_(port), fd_(-1) {}
  ~SerialPort() { (void)close(); }

  ::absl::Status open();
  ::absl::Status close();
  ::absl::Status check_open() const;

  ::absl::Status write(const vector<uint8_t>& data);
  ::absl::StatusOr<vector<uint8_t>> read_full(size_t full_size,
                                              int timeout_ms = 1000);
  ::absl::StatusOr<int> available();

  ::absl::Status flush();

 private:
  string port_;
  int fd_;
};

}  // namespace cc
}  // namespace serialport
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_SERIALPORT_CC_SERIAL_PORT_H
