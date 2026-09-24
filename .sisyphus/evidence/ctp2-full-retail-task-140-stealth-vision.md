# CTP2 Full Retail Task 140: stealth vision-class import and projection

## Implementation
- Retail unit `VisionClass: Stealth` imports to a `stealth` runtime effect.
- Retail unit `CanSee: Stealth` imports to a `can_see_stealth` runtime effect.
- Viewer projections now hide visible-tile enemy stealth units unless the viewer has a stealth-capable detector in sight range.
- Existing submarine detection remains intact and still uses underwater-capable detectors.

## Tests
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/visibility -run 'TestRetail(StealthProjectionRequiresStealthDetection|SubmarineProjectionRequiresUnderwaterDetection|UnitTargetActionsUsePublicTargetMarkers)' -count=1`
- `go test ./internal/gameplay/rules -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Full CTP2 vision classes and globe-satellite exceptions remain future visibility work.
