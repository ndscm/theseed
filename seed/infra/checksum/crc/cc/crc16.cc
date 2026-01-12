#include "seed/infra/checksum/crc/cc/crc16.h"

#include <cstdint>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace infra {
namespace checksum {
namespace crc {
namespace cc {

using ::std::vector;

namespace {

uint16_t reflect(uint16_t data, uint8_t bits) {
  uint16_t reflection = 0x0000;
  for (uint8_t bit = 0; bit < bits; ++bit) {
    if (data & 0x01) {
      reflection |= (1 << ((bits - 1) - bit));
    }
    data >>= 1;
  }
  return reflection;
}

}  // namespace

::absl::StatusOr<uint16_t> CRC16::calculate(const vector<uint8_t>& data) {
  return calculate(data.data(), data.size());
}

::absl::StatusOr<uint16_t> CRC16::calculate(const uint8_t* data,
                                            size_t length) {
  uint16_t crc = init_;
  for (size_t i = 0; i < length; ++i) {
    uint8_t byte = data[i];
    if (refin_) {
      byte = static_cast<uint8_t>(reflect(byte, 8));
    }
    crc ^= (static_cast<uint16_t>(byte) << 8);
    for (int j = 0; j < 8; ++j) {
      if (crc & 0x8000) {
        crc = (crc << 1) ^ poly_;
      } else {
        crc <<= 1;
      }
    }
  }
  if (refout_) {
    crc = reflect(crc, 16);
  }
  return crc ^ xorout_;
}

::absl::Status CRC16::sign_frame(vector<uint8_t>& data) {
  auto crc = calculate(data);
  if (!crc.ok()) {
    return crc.status();
  }
  data.push_back(static_cast<uint8_t>((*crc) & 0xFF));
  data.push_back(static_cast<uint8_t>(((*crc) >> 8) & 0xFF));
  return ::absl::OkStatus();
}

::absl::Status CRC16::verify_frame(const vector<uint8_t>& data) {
  if (data.size() < 2) {
    return ::absl::InvalidArgumentError("Data too short");
  }
  size_t data_length = data.size() - 2;
  auto calculated_crc = calculate(data.data(), data_length);
  if (!calculated_crc.ok()) {
    return calculated_crc.status();
  }
  uint16_t frame_crc = static_cast<uint16_t>(data[data_length]) |
                       (static_cast<uint16_t>(data[data_length + 1]) << 8);
  if (*calculated_crc != frame_crc) {
    return ::absl::DataLossError("CRC verification failed");
  }
  return ::absl::OkStatus();
}

}  // namespace cc
}  // namespace crc
}  // namespace checksum
}  // namespace infra
