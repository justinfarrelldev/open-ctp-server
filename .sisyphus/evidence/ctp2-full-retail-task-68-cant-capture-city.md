# CTP2 Full Retail Task 68: non-capturing units

## Source references
- Retail `Units.txt` marks special/civilian units such as corporate branches with `CantCaptureCity`.
- `unit.cdb` defines `CantCaptureCity` as a gameplay-affecting unit flag.

## Implementation
- Retail unit projection maps `CantCaptureCity` to a runtime `cant_capture_city` effect.
- Combat resolution now prevents an attacking army made only of non-capturing surviving units from moving into and capturing a defeated city.
- Mixed armies can still capture if at least one surviving attacker can capture cities.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCantCaptureCityPreventsCityCapture -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules -count=1`

## Remaining gaps
- Full retail post-combat handling for empty-but-uncaptured cities, retreat, and special units still needs deeper parity review.
