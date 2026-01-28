#ifndef SEED_INFRA_DATAVIEW_CC_DATAVIEW_H
#define SEED_INFRA_DATAVIEW_CC_DATAVIEW_H

#include <cstdint>
#include <vector>

namespace seed {
namespace infra {
namespace dataview {
namespace cc {

class DataView {
 public:
  static uint32_t GetUint32(const ::std::vector<uint8_t>& data, size_t offset,
                            bool little_endian = false);
  static uint16_t GetUint16(const ::std::vector<uint8_t>& data, size_t offset,
                            bool little_endian = false);
};

}  // namespace cc
}  // namespace dataview
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_DATAVIEW_CC_DATAVIEW_H
