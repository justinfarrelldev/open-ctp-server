# CTP2 Full Retail Task 114: city-removing wonder recalculation safety

## Implementation
- Hardened round recalculation so production completion is driven by stable city IDs instead of stale slice indexes.
- If a production-completed wonder removes other cities (for example polluters-to-parks), later city processing skips removed IDs instead of panicking.
- The per-city production/yield sequencing remains in deterministic city-ID order.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailRecalculateRoundSurvivesCityRemovingWonder -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact late-turn ordering between multiple simultaneous city-removing and empire-wide wonder effects still needs a deeper CTP2 audit.
