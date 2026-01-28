#include "seed/infra/rfid/epc/cc/epc_reader.h"

#include <cstdint>
#include <vector>

#include "absl/status/status.h"
#include "absl/status/statusor.h"

namespace seed {
namespace infra {
namespace rfid {
namespace epc {
namespace cc {

using ::std::vector;

::absl::Status EpcReader::CheckHealthy(void) {
  return ::absl::UnimplementedError("CheckHealthy not implemented");
}

::absl::StatusOr<ReaderInfo> EpcReader::GetReaderInfo(void) {
  return ::absl::UnimplementedError("GetReaderInfo not implemented");
}

::absl::StatusOr<vector<vector<uint8_t>>> EpcReader::SynchronousInventory(
    void) {
  return ::absl::UnimplementedError("SynchronousInventory not implemented");
}

::absl::Status EpcReader::WriteTagEpc(const vector<uint8_t>& new_epc,
                                      const vector<uint8_t>& original_epc) {
  return ::absl::UnimplementedError("WriteTagEpc not implemented");
}

}  // namespace cc
}  // namespace epc
}  // namespace rfid
}  // namespace infra
}  // namespace seed
