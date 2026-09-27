# CTP2 Full Retail Tasks 363-368: terrain-improvement special fields

## Source notes
- Retail `tileimp.txt` `TerrainEffect` blocks include gameplay fields beyond yields/costs: `DefenseBonus`, `HealRate`, `VisionRange`, `RadarRange`, `Airport`, `Radar`, `ListeningPost`, `CanUpgrade`, and `Endgame`.
- `TILEIMP_FORTIFICATIONS` has `DefenseBonus 0.5`, `HealRate 1.0`, `VisionRange 2`, and `CanUpgrade`.
- Air bases expose airport/radar-range metadata; listening posts expose radar/listening-post metadata; processing towers are marked `Endgame`.

## Implementation
- Retail parser and runtime projection preserve these terrain-improvement fields.
- Queued/completed canonical overlays preserve these fields when improvements are built.
- Defense-bonus overlays apply a defender combat modifier equal to the configured percentage of the defender's defense.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
- `go test ./internal/gameplay/engine -run 'TestRetail(TerrainImprovementDefenseBonusModifiesDefenders|TerrainImprovementBuildPreservesDefenseMetadata)' -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
