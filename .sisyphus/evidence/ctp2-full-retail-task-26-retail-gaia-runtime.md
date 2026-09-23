# CTP2 full-retail parity task 26 retail Gaia runtime evidence

Scope: wire retail Gaia-controller prerequisites from projected retail runtime into the existing Gaia countdown/victory path.

Original source/data checked:
- Retail `Wonder.txt` marks `WONDER_THE_SOLARIS_PROJECT` with `StartGaiaController`.
- Retail `buildings.txt` defines `IMPROVE_GAIA_COMPUTER` gated by `ADVANCE_GAIA_CONTROLLER`.
- Retail `EndGameObjects.txt` defines `ENDGAME_GAIA_CONTROLLER_CORE` as `Wonder WONDER_THE_SOLARIS_PROJECT` with `TurnsToActivate 10`, and `ENDGAME_GAIA_COMPUTER` as `Building IMPROVE_GAIA_COMPUTER`.

Implemented:
- `RuntimeGaia` now carries optional `building_id` and `wonder_id` alongside activation turns.
- Retail runtime projection sets Gaia to 10 turns, `IMPROVE_GAIA_COMPUTER`, and the `StartGaiaController` wonder (`WONDER_THE_SOLARIS_PROJECT`).
- Engine Gaia prerequisite selection uses runtime Gaia IDs when present, with the synthetic fallback preserved.

Limitations pending full parity:
- Other retail endgame objects (`POWER_SATELLITE`, `PROCESSING_TOWER`, world-union coverage) are not yet fully modeled.
- EndGameObjects.txt is not yet a first-class typed retail catalog; the current retail projection uses the released-game Gaia constants verified above.

Validation commands:
- `go test ./internal/gameplay/rules -run ProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -run RuntimeRulesGaiaPrerequisitesCanComeFromRetailRuntime -count=1`
