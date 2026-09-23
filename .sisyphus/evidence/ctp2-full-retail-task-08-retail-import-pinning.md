# CTP2 full-retail parity task 08 retail import pinning evidence

Scope: make user-owned released-game gamedata importable as a pinned runtime projection instead of only as an in-memory source audit artifact.

Original data/source checked:
- `ctp2_released_game/ctp2_data/default/gamedata/gamefile.txt` points city improvements at `buildings.txt` and terrain improvements at `tileimp.txt`.
- Retail import uses the already-audited released-game catalogs: `Units.txt`, `Advance.txt`, `Wonder.txt`, `buildings.txt`, `tileimp.txt`, `terrain.txt`, `govern.txt`, and `pollution.txt`.
- `ADVANCE_SUBNEURAL_ADS` in retail `Advance.txt` self-prerequisites and is `GLHidden`; direct retail runtime validation therefore preserves retail projection data without applying the stricter synthetic DAG gate.

Implemented server-compatible slice:
- `rules.CanonicalRuntimeDefinitions` canonicalizes any `RuntimeRules` projection and computes a content hash suitable for `RulesPin` use.
- `rules.CompileRuntime` now accepts direct runtime-definition JSON as well as legacy synthetic record-envelope JSON.
- Direct runtime compile validates duplicate/malformed projection records, movement terrain completeness, unit basics, advance references, government unlock references, and starter availability without requiring `UNIT_SYNTHETIC`.
- `cmd/import-rules --retail --data-dir <gamedata>` emits a deterministic retail runtime identity and counts.
- `cmd/import-rules --retail --data-dir <gamedata> --format definitions` emits canonical runtime definitions that compile with the identity hash.

Validation commands:
- `go test ./cmd/import-rules -count=1`
- `go test ./internal/gameplay/rules -count=1`
