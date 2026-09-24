# CTP2 Full Retail Task 73: IgnoreZOC movement

## Source references
- Retail `Units.txt` marks aircraft/special units with `IgnoreZOC`.
- `unit.cdb` defines `IgnoreZOC` as a gameplay-affecting movement flag.

## Implementation
- Retail unit projection maps `IgnoreZOC` to a runtime `ignore_zoc` effect.
- Movement path validation skips enemy ZOC stopping when every moving army member has `ignore_zoc`.
- Fortify action filtering treats IgnoreZOC as retail capability metadata when deciding whether a unit also needs `CanEntrench`.

## Tests
- `go test ./internal/gameplay/engine -run TestTask8MoveTerrainZOCAndVisibilityVectors -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/rules ./internal/gameplay/ai ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full CTP2 ZOC still needs additional review for diplomacy, visibility, stack composition, and special unit exceptions.
