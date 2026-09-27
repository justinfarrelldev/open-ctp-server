# CTP2 Full Retail Tasks 312-316: rebel ownership for revolts/uprisings

## Source notes
- Retail `UnitData::InciteRevolution` calls `CityData::Revolt(..., TRUE)` on success instead of transferring the city to the acting spy owner.
- Retail `CityData::Revolt` sends the city to a `RevoltingCitiesJoinPlayer` wonder owner when applicable, otherwise creates/uses a new civilisation or vandals.
- Retail `CityData::DoUprising` creates an uprising civilisation and `CleanupUprising` transfers the city to the uprising owner on successful slave uprising.

## Implementation
- Added deterministic rebel-player creation (`rebels-<old-owner>`, `rebel-nation-<old-owner>`) with AI controller and active turn-order inclusion.
- Incite-revolution now clears the city queue, removes attached trade routes, updates city tile ownership, and transfers the city to the retail-style join-player wonder owner or rebel owner instead of the acting player.
- Slave-uprising now removes slave specialists and transfers the city through the same rebel ownership path while keeping the safe aggregate event payload.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(SlaveUprisingTransfersCityToRebels|InciteRevolutionUsesChanceAndTransfersCityToRebels|InciteRevolutionHonorsRevoltingCitiesJoinPlayerWonder)' -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
- `go test ./internal/gameplay/ai -count=1`
