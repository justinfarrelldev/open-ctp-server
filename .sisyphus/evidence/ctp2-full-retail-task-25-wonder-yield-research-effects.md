# CTP2 full-retail parity task 25 wonder yield/research effects evidence

Scope: project a bounded set of retail wonder effects that map to existing city yield/research mechanics.

Original source/data checked:
- Retail `Wonder.txt` includes `IncKnowledgePercent 20` on `WONDER_ARISTOTLES_LYCEUM` and `35` on `WONDER_FIELD_DYNAMICS_LABORATORY`.
- Retail `Wonder.txt` includes `IncreaseProduction 10` on `WONDER_GENOME_PROJECT`.
- Retail `Wonder.txt` includes `IncreaseFoodAllCities 30` on `WONDER_NANITE_DEFUSER`.
- Retail `Wonder.txt` includes `BonusGold 100` on `WONDER_PYRAMIDS`.

Implemented:
- Retail runtime projection maps:
  - `BonusGold` -> `yield_gold`
  - `IncKnowledgePercent` -> `research`
  - `IncreaseFoodAllCities` -> `yield_food_percent`
  - `IncreaseProduction` -> `yield_production_percent`
- Engine-owned wonder effects already apply to all owner cities through `cityEffects`; tests now assert owner-city yield changes from wonder effects.

Limitations pending full parity:
- This does not yet implement all wonder fields (diplomacy, slavery, radar, protection, Gaia details, scientists/specialists, trade-route bonuses, etc.).
- Research effect remains bounded to current research-rate mechanics pending final CTP2 science slider parity.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run RetailRuntimeWonderEffectsDriveOwnedCityYields -count=1`
