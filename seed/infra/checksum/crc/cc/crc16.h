#ifndef INFRA_CHECKSUM_CRC_CC_CRC16_H
#define INFRA_CHECKSUM_CRC_CC_CRC16_H

#include <cstdint>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace infra {
namespace checksum {
namespace crc {
namespace cc {

using ::std::vector;

class CRC16 {
 private:
  const uint16_t poly_;
  const uint16_t init_;
  const uint16_t xorout_;
  const bool refin_;
  const bool refout_;

  CRC16(uint16_t poly, uint16_t init, uint16_t xorout, bool refin, bool refout)
      : poly_(poly),
        init_(init),
        xorout_(xorout),
        refin_(refin),
        refout_(refout) {}

 public:
  ::absl::StatusOr<uint16_t> calculate(const vector<uint8_t>& data);
  ::absl::StatusOr<uint16_t> calculate(const uint8_t* data, size_t length);

  ::absl::Status sign_frame(vector<uint8_t>& data);
  ::absl::Status verify_frame(const vector<uint8_t>& data);

  static CRC16 MODBUS() { return CRC16(0x8005, 0xFFFF, 0x0000, true, true); }
  static CRC16 MCRF4XX() { return CRC16(0x1021, 0xFFFF, 0x0000, true, true); }
};

}  // namespace cc
}  // namespace crc
}  // namespace checksum
}  // namespace infra

#endif  // INFRA_CHECKSUM_CRC_CC_CRC16_H
