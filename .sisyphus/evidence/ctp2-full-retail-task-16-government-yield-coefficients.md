# CTP2 full-retail parity task 16 government yield coefficients evidence

Scope: project core retail government economy coefficients into runtime and apply them as nation effects.

Original source/data checked:
- Retail `govern.txt` contains `ProductionCoef`, `KnowledgeCoef`, `GoldCoef`, `MaxScienceRate`, and `PollutionCoef` for every government.
- Retail `GOVERNMENT_DEMOCRACY` has `EnableAdvance ADVANCE_DEMOCRACY`, `ProductionCoef 1.15`, `KnowledgeCoef 2`, and `GoldCoef 1.25`.
- Original city production code applies government production coefficient after building/wonder modifiers.

Implemented server-compatible slice:
- Runtime governments now include `production_percent` and `gold_percent` in addition to existing research/public-works fields.
- Retail projection maps `ProductionCoef`/`GoldCoef` to positive percentage bonuses and `KnowledgeCoef` to the existing flat research bonus slice.
- Government switching installs production/gold percentage effects on the nation, preserving other non-government effects and replacing previous government effects.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run GovernmentRuntimeYieldPercentEffects -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
