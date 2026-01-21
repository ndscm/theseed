#include "seed/infra/hex/cc/hex.h"

#include <cstdint>
#include <string>
#include <vector>

#include "absl/status/statusor.h"

namespace seed {
namespace infra {
namespace hex {
namespace cc {

using ::std::string;
using ::std::vector;

static const char kHexChars[] = "0123456789abcdef";

string ConvertBinaryToHex(const vector<uint8_t>& data,
                          const string& separator) {
  string result;
  result.reserve(data.size() * (2 + separator.size()));
  for (size_t i = 0; i < data.size(); ++i) {
    uint8_t byte = data[i];
    result += kHexChars[(byte >> 4) & 0x0F];
    result += kHexChars[byte & 0x0F];
    if (i < data.size() - 1 || separator.empty()) {
      result += separator;
    }
  }
  return result;
}

::absl::StatusOr<vector<uint8_t>> ConvertHexToBinary(const string& hex) {
  vector<uint8_t> result;
  string hex_chars;

  // Extract only hex characters, skipping separators (space, tab, comma)
  for (char c : hex) {
    if ((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') ||
        (c >= 'A' && c <= 'F')) {
      hex_chars += c;
    } else if (c == ' ' || c == '\t' || c == ',') {
      // Skip separators
      continue;
    } else {
      return ::absl::InvalidArgumentError(
          string("Invalid character in hex string: '") + c + "'");
    }
  }

  // Must have even number of hex characters
  if (hex_chars.size() % 2 != 0) {
    return ::absl::InvalidArgumentError(
        "Hex string must have even number of hex digits");
  }

  // Convert pairs of hex characters to bytes
  for (size_t i = 0; i < hex_chars.size(); i += 2) {
    uint8_t high = 0, low = 0;

    char h = hex_chars[i];
    if (h >= '0' && h <= '9') {
      high = h - '0';
    } else if (h >= 'a' && h <= 'f') {
      high = 10 + (h - 'a');
    } else if (h >= 'A' && h <= 'F') {
      high = 10 + (h - 'A');
    }

    char l = hex_chars[i + 1];
    if (l >= '0' && l <= '9') {
      low = l - '0';
    } else if (l >= 'a' && l <= 'f') {
      low = 10 + (l - 'a');
    } else if (l >= 'A' && l <= 'F') {
      low = 10 + (l - 'A');
    }

    result.push_back((high << 4) | low);
  }

  return result;
}

}  // namespace cc
}  // namespace hex
}  // namespace infra
}  // namespace seed
