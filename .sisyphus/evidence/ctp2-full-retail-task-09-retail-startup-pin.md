# CTP2 full-retail parity task 09 retail startup pin evidence

Scope: allow server startup to use a pinned retail runtime projection from user-owned released-game data.

Original data/source checked:
- Released-game gamedata catalogs remain the source for the retail projection (`Units.txt`, `Advance.txt`, `Wonder.txt`, `buildings.txt`, `tileimp.txt`, `terrain.txt`, `govern.txt`, `pollution.txt`).
- Startup still validates the resulting pin through `rules.CompileRuntime` before deriving the initial state.

Implemented server-compatible slice:
- Default startup remains server-owned synthetic rules for deterministic existing tests and deployments.
- Setting `OPEN_CTP_RULES_MODE=retail` and `OPEN_CTP_RETAIL_GAMEDATA=<released-game-gamedata>` switches startup to a canonical retail runtime projection pin.
- `initialGameplayState` now works with the direct retail runtime pin; starter units inherit `UNIT_SETTLER`-projected runtime capabilities.
- Missing retail data directory fails closed before startup state derivation.

Validation commands:
- `go test ./internal/game -run 'ServerRuntimePinCanUseRetailProjectionFromEnvironment|RetailRuntimeModeRequiresDataDirectory' -count=1`
- `go test ./internal/game -count=1`
