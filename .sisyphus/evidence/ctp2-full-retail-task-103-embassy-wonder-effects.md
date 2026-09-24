# CTP2 Full Retail Task 103: embassy wonder effects

## Source references
- `Player::HasEmbassyWith` in `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` returns false when the target owns `CloseEmbassies`, true when the actor owns `EmbassiesEverywhereEvenAtWar`, and otherwise uses explicit embassies.
- `Player.cpp` wonder completion closes foreign embassies with a new `CloseEmbassies` owner and neutralizes diplomacy for `CloseEmbassies` / `EmbassiesEverywhereEvenAtWar`.
- `UnitData::EstablishEmbassy` rejects establishment against a `CloseEmbassies` owner.

## Implementation
- Added effective-embassy evaluation that honors `close_embassies` and `embassies_everywhere` wonders.
- Embassy special actions now reject protected targets and redundant effective embassies.
- Completing `close_embassies` removes foreign embassies with the owner and clears wars involving the owner.
- Completing `embassies_everywhere` clears wars involving the owner.
- Viewer projections/AI inputs now report effective embassy player IDs without exposing hidden canonical state.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(EmbassyWonderEffects|CloseEmbassies)' -count=1`
- `go test ./internal/gameplay/visibility -run TestRetailEmbassiesEverywhereProjectionHonorsCloseEmbassies -count=1`
- `go test ./internal/gameplay/engine ./internal/gameplay/visibility -count=1`

## Remaining gaps
- Exact retail diplomatic regard shifts and proposal state cleanup for these wonders remain future work.
