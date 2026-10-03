# Task 044: retail DoubleTelevangelists conversion gold gate

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:5860-5877`
  - Converted cities siphon gold with the cleric conversion factor by default.
  - If `m_convertedBy == CONVERTED_BY_TELEVANGELIST`, the city only uses `TelevangelistConversionFactor()` when `buildingutil_GetDoubleTelevangelism(GetEffectiveBuildings(), doubleTelevangelists)` is true; otherwise it falls back to the cleric factor.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/buildingutil.cpp:154-162`
  - `buildingutil_GetDoubleTelevangelism` detects active `DoubleTelevangelists` buildings.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Buildings.txt` / imported runtime effects
  - Retail building records include the `DoubleTelevangelists` marker imported as `double_televangelists`.

## Implementation notes

- `conversionGoldPercent` now receives canonical state so it can evaluate effective city building/wonder effects.
- Televangelist-converted cities use the enhanced conversion percentage only when the converted city has an active `EffectDoubleTelevangelists`; otherwise they use the cleric percentage exactly like retail.
- Existing cleric conversion behavior is unchanged.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-044-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailTelevangelist|TestRetailConversionSiphonsGold' -count=1`
