# CTP2 full-retail parity task 23 building SciencePercent evidence

Scope: project retail building `SciencePercent` fields into runtime effects and confirm authoritative research calculation consumes city building research effects.

Original source/data checked:
- Retail `buildings.txt` includes `SciencePercent` for buildings such as `IMPROVE_ACADEMY` (`0.1`), `IMPROVE_UNIVERSITY` (`0.2`), and later science buildings.
- CTP2 city science is derived from city commerce/trade after science allocation and building/city modifiers; this server currently models the bounded percent bonus as a city research effect.

Implemented:
- `retailBuildingRuntimeEffects` now maps `SciencePercent` to runtime effect kind `research` with integer percentage points.
- `IMPROVE_ACADEMY` now projects a `research` effect value of `10`.
- Engine research-rate tests verify city building research effects are consumed in addition to city trade yield.

Limitations pending full parity:
- This is a bounded flat/percentage-point projection into the current research model, not the final CTP2 slider/science-allocation formula.
- Other building effects from the field inventory (crime, happiness, pollution, population, protection, combat, trade, etc.) remain pending.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'RetailResearchRateUsesCityTradeYield|RetailResearchRateUsesCityScienceBuildingEffect' -count=1`
