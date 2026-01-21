#ifndef SEED_INFRA_HEX_CC_HEX_H
#define SEED_INFRA_HEX_CC_HEX_H

#include <cstdint>
#include <string>
#include <vector>

#include "absl/status/statusor.h"

namespace seed {
namespace infra {
namespace hex {
namespace cc {

::std::string ConvertBinaryToHex(const ::std::vector<uint8_t>& data,
                                 const ::std::string& separator = "");

::absl::StatusOr<::std::vector<uint8_t>> ConvertHexToBinary(
    const ::std::string& hex);

}  // namespace cc
}  // namespace hex
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_HEX_CC_HEX_H
