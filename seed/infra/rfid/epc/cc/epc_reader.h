#ifndef SEED_INFRA_RFID_EPC_CC_EPC_READER_H
#define SEED_INFRA_RFID_EPC_CC_EPC_READER_H

#include <cstdint>
#include <string>
#include <unordered_map>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace seed {
namespace infra {
namespace rfid {
namespace epc {
namespace cc {

class AntennaInfo {
 public:
  float power_dbm;
};

class ReaderInfo {
 public:
  ::std::string driver;
  ::std::unordered_map<uint8_t, AntennaInfo> antennas;
};

class EpcTag {
 public:
  ::std::vector<uint8_t> epc;
  uint8_t antenna = 0;
  int8_t rssi = 0;
  int8_t confidence = 0;
};

class EpcReader {
 protected:
  EpcReader(void) = default;

 public:
  virtual ~EpcReader(void) = default;

  // Reader
  virtual ::absl::Status CheckHealthy(void);
  virtual ::absl::StatusOr<ReaderInfo> GetReaderInfo(void);

  // Tag Inventory
  virtual ::absl::StatusOr<::std::vector<EpcTag>> SynchronousInventory(
      uint8_t antenna = 0);

  // Tag Access
  virtual ::absl::Status WriteTagEpc(
      const ::std::vector<uint8_t>& new_epc,
      const ::std::vector<uint8_t>& original_epc = {});
};

}  // namespace cc
}  // namespace epc
}  // namespace rfid
}  // namespace infra
}  // namespace seed

#endif  // SEED_INFRA_RFID_EPC_CC_EPC_READER_H
