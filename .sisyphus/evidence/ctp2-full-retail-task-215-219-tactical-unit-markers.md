# CTP2 Full Retail Tasks 215-219: tactical unit marker field preservation

## Source evidence
- `unit.cdb` declares tactical/unit-order fields `IsFlanker`, `CanBeachAssault`, `Paratrooper`, `ParatrooperTransport`, and `SpaceLaunch`.
- Retail `Units.txt` uses these flags/values on units such as cavalry, fighters, cargo helicopters, paratroopers, and the space plane.

## Implementation
- Runtime unit effects now preserve these retail tactical markers:
  - `is_flanker`
  - `can_beach_assault`
  - `paratrooper`
  - `paratrooper_transport`
  - `space_launch_turns` with the retail turn value
- Engine state validation accepts the marker effects so produced/runtime units remain restart-safe.
- No full execution claim is made for beach assault/paradrop/space-launch orders yet; this closes the source-data preservation gap for subsequent mechanics.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`
