# Original-compatible settlement sites

The game97 settlement rejection is repaired in server execution, advertised actions and AI planning. No client bypass, database migration, save edits or rules-pin rewrites are needed. A running server must load the updated code before the fix takes effect.

## Site rule

- National ownership alone does not prevent founding, including foreign ownership.
- Any recorded city influence prevents founding, including your own city's influence.
- An existing or adjacent city prevents founding, regardless of its owner. Adjacency uses the server's existing bounded eight-neighbor tile geometry, not a new wrapped/skew-map implementation.
- Existing unit ownership/capability, army/transport, identity/name, phase, revision and transactional admission remain unchanged. This repair does not introduce new settler movement or terrain-capability rules.

`engine.CitySiteAvailable` is the shared, read-only site predicate. Execution supplies every canonical city center. `visibility.FoundingSiteAvailable` supplies only current projected tiles/cities and is shared with the AI; unit-action construction reuses its existing tile index. Missing/stale site data does not advertise founding. Stale city memory is not treated as a current obstacle. Filtering never consults hidden canonical cities; execution still rejects a hidden obstacle or a site that became unavailable after projection. Site errors remain generic and contain no city identities.

The rule applies to both legacy/synthetic and territory-configured pins. Existing legacy states are not backfilled with guessed influence radii. Recorded influence and known city-center adjacency are enforced; successful founding initializes the center and any territory supported by the existing pin. Existing cities are not retroactively moved or deleted.

## Source review

Authority under `ctp2/ctp2_source_code/ctp2_code/`:

- `gs/gameobj/UnitData.cpp::UnitData::Settle`: rejects `Cell::GetCityOwner().IsValid()`; this is influence, not the national `GetOwner()` field.
- `gs/world/Cell.cpp::GetCityOwner`, `SetCityOwner`, and `Cell.h::GetOwner`: separate city-radius assignment from national ownership.
- `gs/gameobj/armyevent.cpp::ArmySettleOrderEvent`, `ArmyData.cpp::Settle`, `PlayerEvent.cpp::SettleEvent`, `Player.cpp::Settle`: no intervening foreign-national-owner prohibition.
- `PlayerEvent.cpp::CreateCityEvent`, `Player.cpp::CreateCity`, `gs/world/WrlEnv.cpp::IsNextToCity`, `WrldCity.cpp::InsertCity`: occupied/adjacent-city admission followed by creation and new national ownership assignment.

Unlike the original event chain, rejection occurs before consuming the settler; a failed site check cannot destroy it or leave a partial city.

## Regression coverage and fixture changes

New tests cover foreign-national settlement, own/foreign influence, occupied/adjacent sites, legacy/configured pins, atomic consumption and territory initialization, exact deterministic replay, canonical restart, unit/identity/turn safeguards, executable advertised payloads, AI policy and viewer noninterference. RED failures were observed before implementation; focused suites passed after implementation and tile-index refactoring.

Older naming/victory/action-metadata fixtures placed cities adjacently or omitted the settler's projected tile. They now supply legal sites while retaining their original assertions. The three-city continuity harness now uses visible, affordable paths to non-adjacent sites, allows visible transit to reveal further sites and waits for genuine movement refresh with the phase-appropriate closeout. It does not mutate canonical state or invent hidden routes.

The old cold-map catalog seed reaches a score tie after legal spacing, rather than the terminal previously obtained using illegal adjacent founding. The bounded terminal-opener unit test uses a deterministic non-tied seed and retains all terminal assertions. Historical catalog hashes are **not** evidence of equivalence after this deliberate gameplay-rule correction; the complete opt-in catalog acceptance campaign has not been regenerated.

Verification commands and final outcomes are recorded in `plans/checklists/city-settlement-parity.md`. No live gameplay commands were submitted; live game97 acceptance remains unmeasured.
