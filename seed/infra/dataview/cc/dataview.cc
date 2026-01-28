#include "seed/infra/dataview/cc/dataview.h"

namespace seed {
namespace infra {
namespace dataview {
namespace cc {

uint32_t DataView::GetUint32(const ::std::vector<uint8_t>& data, size_t offset,
                             bool little_endian) {
  if (little_endian) {
    return (static_cast<uint32_t>(data[offset + 3]) << 24) |
           (static_cast<uint32_t>(data[offset + 2]) << 16) |
           (static_cast<uint32_t>(data[offset + 1]) << 8) |
           static_cast<uint32_t>(data[offset]);
  }
  return (static_cast<uint32_t>(data[offset]) << 24) |
         (static_cast<uint32_t>(data[offset + 1]) << 16) |
         (static_cast<uint32_t>(data[offset + 2]) << 8) |
         static_cast<uint32_t>(data[offset + 3]);
}

uint16_t DataView::GetUint16(const ::std::vector<uint8_t>& data, size_t offset,
                             bool little_endian) {
  if (little_endian) {
    return (static_cast<uint16_t>(data[offset + 1]) << 8) |
           static_cast<uint16_t>(data[offset]);
  }
  return (static_cast<uint16_t>(data[offset]) << 8) |
         static_cast<uint16_t>(data[offset + 1]);
}

}  // namespace cc
}  // namespace dataview
}  // namespace infra
}  // namespace seed
