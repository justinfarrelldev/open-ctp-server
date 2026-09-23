# CTP2 full-retail parity task 36 population-limited worked tiles evidence

Scope: bound city worked-tile economy by population instead of allowing every listed worked tile to contribute regardless of city size.

Implemented:
- City center yield remains automatic.
- Additional worked tiles are capped to `Population` entries from `WorkedTileIDs`.
- Population `0` remains legacy-compatible/unbounded for older synthetic fixtures that predate population state.
- Duplicate worked-tile validation still applies across the complete submitted `WorkedTileIDs` list, including ignored overflow entries.

Validation commands:
- `go test ./internal/gameplay/engine -run TestTask10ProductionAndEffectFailureVectors -count=1`
- `GOMAXPROCS=4 go test ./... -count=1`
- `git diff --check`
