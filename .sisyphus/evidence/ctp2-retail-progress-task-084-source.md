# Task 084 source notes — remaining cargo metadata/fate closure

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:3979-4039` parses nested `CargoData` fields as `MaxCargo`, `Load`, and `Unload`; `Load`/`Unload` are looked up from the Sound DB.
- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:5116-5125` exposes `UnitRecord::CargoData::GetLoad()` and `GetUnload()` as `SoundRecord` accessors.
- No gameplay call sites reference `GetCargoDataPtr()->GetLoad`, `CargoData::GetLoad`, `GetCargoDataPtr()->GetUnload`, or `CargoData::GetUnload`; load/unload audio used by orders comes from top-level `SoundLoad`/`SoundUnload` instead.
- `ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:6188-6200` plays `transports[j].GetLoadSoundID()` during boarding, and `ArmyData.cpp:6425-6435` plays `m_array[0].GetUnloadSoundID()` during unload.
- `ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:220-272` removes a dying unit from its transport if needed and then calls `AccessData()->KillTransportedUnits()`.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:5369-5376` recursively kills cargo with `CAUSE_REMOVE_ARMY_TRANSPORT_DIED`.

Server closure:

- Nested `CargoData.Load`/`CargoData.Unload` are classified as presentation-only/unreferenced sound metadata for server gameplay. Top-level `Sound*` fields were already classified in Task 082.
- Transport destruction cargo fate is already implemented by `removeUnitAndCargo`, which recursively removes transported units; `internal/gameplay/engine/units_task8_test.go:136-156` covers atomic cargo removal when a transport is disbanded.

Validation:

- `go test ./internal/gameplay/engine -run TestTask8CapacityPhaseAndTransportLossVectors -count=1`
- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
