# CTP2 Full Retail Task 130: submarine projection detection

## Implementation
- Viewer projections now hide enemy submarine units even when the submarine tile is otherwise visible.
- Enemy submarines are projected only when the viewer has an underwater-capable detector unit within that unit's sight radius.
- Own units remain fully projected; non-submarine visible enemy units retain prior safe public projection behavior.

## Tests
- `go test ./internal/gameplay/visibility -run TestRetailSubmarineProjectionRequiresUnderwaterDetection -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full CTP2 stealth vision classes, sonar-specific fields, and active stealth reveal timers remain future grouped work.
