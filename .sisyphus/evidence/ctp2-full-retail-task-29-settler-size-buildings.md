# CTP2 full-retail parity task 29 settler size/building effects evidence

Scope: preserve and execute retail unit `SettleSize` and `SettleBuilding` fields when a settler founds a city.

Retail evidence:
- `Units.txt` `UNIT_SETTLER` has `SettleSize 1`.
- `Units.txt` `UNIT_SEA_ENGINEER` has `SettleSize 3` and several `SettleBuilding` entries including `IMPROVE_GRANARY`.

Implemented:
- Retail runtime unit projection emits safe starter effects:
  - `SettleSize` -> `settle_size` effect value.
  - `SettleBuilding` -> `settle_building` effect IDs.
- Founded cities now store `Population` and seed it from the founding unit's `settle_size` effect, defaulting to 1.
- Founded cities now seed building IDs from `settle_building` effects.

Validation commands:
- `go test ./internal/gameplay/rules -run RetailRuntime -count=1`
- `go test ./internal/gameplay/engine -run RetailSettlerEffectsSeedFoundedCitySizeAndBuildings -count=1`
