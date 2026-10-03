# Task 083 source notes — cargo size/class compatibility

Retail source/data evidence:

- `ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:2531-2591` implements `UnitCanCarry(dest, src)` by first requiring `dest.GetCanCarry()`, then checking the carried unit's `SizeSmall`/`SizeMedium`/`SizeLarge` and matching one of the carrier's `CanCarrySmall|Med|Large` bits for the carried unit's movement domain (`Land`, `Air`, `Sea`, `Space`). This means broad domain matching alone is insufficient.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:749-755` delegates `UnitData::CanCarry(src)` to `UnitCanCarry(m_type, src)`.
- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:4362-4376` defines the retail `CanCarry` tokens: `SmallLand`, `MedLand`, `LargeLand`, `SmallAir`, `MedAir`, `LargeAir`, `SmallWater`, `MedWater`, `LargeWater`, `SmallSpace`, `MedSpace`, `LargeSpace`.
- `ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:663-668` enforces `CargoData.MaxCargo` as the capacity count when inserting cargo.
- Retail data example: `ctp2_released_game/ctp2_data/default/gamedata/Units.txt:94-144` defines `UNIT_AIRCRAFT_CARRIER` with `CanCarry: MedAir`, `CanCarry: SmallAir`, and `Size: Large`.

Implementation notes:

- Runtime unit projection now imports `Size` and exact `CargoClasses` from retail `CanCarry` tokens while retaining legacy broad `CargoDomains` for compatibility.
- Runtime-created and production-created units carry size/class metadata into canonical state and visibility projections.
- Engine embark validation, visibility legal actions, and projection-only AI embark candidates require exact size/domain class matching when `CargoClasses` are present, including amphibious units matching either land or water carry classes to reflect retail's independent movement-type checks. Legacy synthetic data falls back to broad-domain matching.

Validation:

- RED evidence: `.sisyphus/evidence/ctp2-retail-progress-task-083-red.txt` and `.sisyphus/evidence/ctp2-retail-progress-task-083-red-amphibious.txt`.
- GREEN targeted package evidence: `.sisyphus/evidence/ctp2-retail-progress-task-083-green-packages.txt`.
- GREEN full suite evidence: `.sisyphus/evidence/ctp2-retail-progress-task-083-green-full.txt`.
- Freeze lock evidence: `.sisyphus/evidence/ctp2-retail-progress-task-083-freeze.txt`.
