#include "seed/infra/serialport/cc/serial_port.h"

#include <errno.h>
#include <fcntl.h>
#include <poll.h>
#include <sys/ioctl.h>
#include <termios.h>
#include <unistd.h>

#include <chrono>
#include <cstring>
#include <iostream>
#include <thread>

#include "absl/status/status.h"
#include "absl/status/statusor.h"
#include "absl/strings/str_cat.h"
#include "absl/strings/str_format.h"

namespace seed {
namespace infra {
namespace serialport {
namespace cc {

using ::std::string;
using ::std::vector;

::absl::Status SerialPort::open() {
  if (fd_ >= 0) {
    return ::absl::OkStatus();
  }

  ::std::cerr << ::absl::StrFormat("[SerialPort] Opening port: %s\n", port_);
  fd_ = ::open(port_.c_str(), O_RDWR | O_NOCTTY | O_NONBLOCK);
  if (fd_ < 0) {
    return ::absl::InternalError(::absl::StrCat("Failed to open port ", port_,
                                                ": ", ::std::strerror(errno)));
  }
  ::std::cerr << "[SerialPort] Port opened successfully. fd:" << fd_ << "\n";

  termios tty;
  errno = 0;
  if (::tcgetattr(fd_, &tty) != 0) {
    string error_message =
        ::absl::StrCat("tcgetattr failed: ", ::std::strerror(errno));
    ::close(fd_);
    fd_ = -1;
    return ::absl::InternalError(error_message);
  }

  // Configure: 8N1
  tty.c_cflag &= ~PARENB;         // No parity
  tty.c_cflag &= ~CSTOPB;         // 1 stop bit
  tty.c_cflag &= ~CSIZE;          // Clear size bits
  tty.c_cflag |= CS8;             // 8 data bits
  tty.c_cflag |= CREAD | CLOCAL;  // Enable reading, local mode

  tty.c_lflag &= ~ICANON;  // Non-canonical mode
  tty.c_lflag &= ~ECHO;    // No echo
  tty.c_lflag &= ~ISIG;    // No signals

  tty.c_iflag &= ~(IXON | IXOFF | IXANY);  // No software flow control
  tty.c_iflag &= ~(ISTRIP | IGNCR | INLCR | ICRNL);

  tty.c_oflag &= ~OPOST;  // Raw output

  tty.c_cc[VMIN] = 0;   // Non-blocking read
  tty.c_cc[VTIME] = 0;  // No timeout

  ::cfsetispeed(&tty, baud_rate_);
  ::cfsetospeed(&tty, baud_rate_);

  errno = 0;
  if (::tcsetattr(fd_, TCSANOW, &tty) != 0) {
    string error_message =
        ::absl::StrCat("tcsetattr failed: ", ::std::strerror(errno));
    ::close(fd_);
    fd_ = -1;
    return ::absl::InternalError(error_message);
  }

  errno = 0;
  if (::tcflush(fd_, TCIOFLUSH) != 0) {
    string error_message =
        ::absl::StrCat("tcflush failed: ", ::std::strerror(errno));
    ::close(fd_);
    fd_ = -1;
    return ::absl::InternalError(error_message);
  }
  return ::absl::OkStatus();
}

::absl::Status SerialPort::close() {
  if (fd_ >= 0) {
    errno = 0;
    if (::close(fd_) != 0) {
      return ::absl::InternalError(
          ::absl::StrCat("Failed to close port: ", ::std::strerror(errno)));
    }
    fd_ = -1;
  }
  return ::absl::OkStatus();
}

::absl::Status SerialPort::check_open() const {
  if (fd_ < 0) {
    return ::absl::FailedPreconditionError("Port not open");
  }
  return ::absl::OkStatus();
}

::absl::Status SerialPort::write(const vector<uint8_t>& data) {
  if (data.empty()) {
    return ::absl::OkStatus();
  }
  if (fd_ < 0) {
    return ::absl::FailedPreconditionError("Port not open");
  }
  size_t data_size = data.size();
  string message = ::absl::StrFormat("[SerialPort] Writing %d bytes: ",
                                     static_cast<int>(data_size));
  for (uint8_t b : data) {
    ::absl::StrAppendFormat(&message, "%02x ", b);
  }
  ::std::cerr << ::absl::StrCat(message, "\n");
  size_t total_written = 0;
  while (total_written < data_size) {
    errno = 0;
    ssize_t n =
        ::write(fd_, data.data() + total_written, data_size - total_written);
    if (n < 0) {
      return ::absl::InternalError(
          ::absl::StrCat("Write failed: ", ::std::strerror(errno)));
    }
    total_written += n;
  }
  ::std::cerr << ::absl::StrFormat(
      "[SerialPort] Write complete (%d bytes total)\n",
      static_cast<int>(total_written));
  return ::absl::OkStatus();
}

::absl::StatusOr<vector<uint8_t>> SerialPort::read_full(size_t full_size,
                                                        int timeout_ms) {
  if (fd_ < 0) {
    return ::absl::FailedPreconditionError("Port not open");
  }
  auto start = ::std::chrono::steady_clock::now();
  vector<uint8_t> buffer(full_size);
  size_t total_read = 0;
  while (total_read < full_size) {
    auto elapsed = ::std::chrono::duration_cast<std::chrono::milliseconds>(
                       ::std::chrono::steady_clock::now() - start)
                       .count();
    if (elapsed > timeout_ms) {
      return ::absl::DeadlineExceededError("Timeout waiting for data");
    }
    errno = 0;
    ssize_t n = ::read(fd_, buffer.data() + total_read, full_size - total_read);
    if (n < 0) {
      if (errno != EAGAIN && errno != EWOULDBLOCK) {
        return ::absl::InternalError(
            ::absl::StrCat("Read failed: ", ::std::strerror(errno)));
      }
    }
    if (n > 0) {
      total_read += n;
    }
    ::std::this_thread::sleep_for(::std::chrono::milliseconds(10));
  }

  string message =
      ::absl::StrFormat("[SerialPort] Read %d bytes: ", buffer.size());
  for (size_t i = 0; i < buffer.size(); ++i) {
    ::absl::StrAppendFormat(&message, "%02x ", buffer[i]);
  }
  ::std::cerr << ::absl::StrCat(message, "\n");
  return buffer;
}

::absl::StatusOr<int> SerialPort::available() {
  if (fd_ < 0) {
    return ::absl::FailedPreconditionError("Port not open");
  }
  int bytes_available = 0;
  errno = 0;
  if (::ioctl(fd_, FIONREAD, &bytes_available) < 0) {
    return ::absl::InternalError(
        ::absl::StrCat("ioctl(FIONREAD) failed: ", ::std::strerror(errno)));
  }
  return bytes_available;
}

::absl::Status SerialPort::flush() {
  if (fd_ < 0) {
    return ::absl::FailedPreconditionError("Port not open");
  }
  errno = 0;
  if (::tcflush(fd_, TCIOFLUSH) != 0) {
    return ::absl::InternalError(
        ::absl::StrCat("tcflush failed: ", ::std::strerror(errno)));
  }
  return ::absl::OkStatus();
}

}  // namespace cc
}  // namespace serialport
}  // namespace infra
}  // namespace seed
