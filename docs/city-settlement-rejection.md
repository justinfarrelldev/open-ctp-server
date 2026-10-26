# Game97: advertised settlement is rejected

**Repair status:** implemented in [city-settlement-parity.md](city-settlement-parity.md). The report below records the pre-fix investigation and its historical results; it is not a description of current enforcement.

## Conclusion

This is a **server settlement-rule bug plus an advertised-action mismatch**, not a malformed client payload or a regression introduced by the turn-performance changes.

The engine treats **foreign national ownership** as a blanket settlement prohibition. Original released CTP2 instead rejects a tile already assigned to a **city's influence area**, including your own city's area. These are different fields and different rules.

Game97's human settler is on a foreign nationally owned tile with **no city influence assignment**. The server advertises Found City there, then rejects its own advertised payload. The public error/log category hides the underlying reason as `command_invalid`.

No live commands were submitted or records modified. No runtime code or client files were changed for this investigation.

## Actual read-only evidence

At inspection, game97 remains revision17/turn3, with the human player active. The human owns a settler and no city; three AI cities exist server-side. A viewer log saying `0 cities` is not evidence that the entire game has no cities: foreign cities may be outside the viewer's sight.

Current human unit/site:

- Unit: `settler-184-1`, type `UNIT_SETTLER`, owner `account-184`.
- Position: `tile-20-6` (`20,6`).
- Movement:100; no army or transport; the actual `settler` capability is present.
- National owner: `account-187`, **not** the human player.
- `influence_city_id`: empty; no city occupies this tile.
- Advertised command: `found-city`, payload:

```json
{"unit_id":"settler-184-1","city_id":"client-city-settler-184-1","name":"Athens"}
```

The stored human revision10 receipt already advertises that exact payload at this site despite the foreign national owner. Earlier human receipts2/3 place the settler on `tile-20-7`, before AI city founding. That earlier tile is also foreign national territory with empty city influence in the inspected state. AI city-founded events occur at revisions4/6/8, followed by turn resolution at9; the next round resolves at17.

Rejected commands do not produce accepted-command receipts. The supplied fingerprints/correlation IDs do not retain the raw rejected bodies, so this investigation does not claim to recover every individual request or an archived complete canonical state at revision9. The revision17 failure is reproduced directly, and the receipt history/source explain the earlier same-rule failures.

The client's UTC `17:54` timestamps and the server's local `13:54` timestamps describe the same period, not a four-hour processing delay.

## Offline reproduction and controls

Canonical state was obtained by a read-only SELECT and passed through stdin to an external diagnostic. It calls the real projector to obtain the advertised action, then the real `engine.Apply` with injected clock/ID/RNG sources. Results are discarded in memory; no HTTP mutation or persistence gateway is called.

| Copied-state scenario | Result |
| --- | --- |
| Actual state + exact advertised Found City payload | Rejects: **`city site is not owned by founding player`** |
| Same tile/payload; only national owner changed to neutral | Accepts; computes revision18 and creates the human city |
| Same tile/payload; only national owner changed to the human | Accepts; computes revision18 and creates the human city |
| Neutral-owner control, but settler capability removed | Rejects: `unit cannot found a city` |

Both successful controls leave the actual game at revision17. They establish that the name, city/unit identities and the settler's ordinary eligibility are not the cause of the reproduced failure.

All four cases were run against **both** current server `7085e58` and original pre-performance `b8fbda9` implementations using compiler overlays. Results match in every case. `cities.go` and `city_territory.go` also have no changes between those commits. This is not caused by the new caches, faster AI, snapshot streaming or client font changes.

Local diagnostic artifacts: `/tmp/ctp-settlement-diagnosis.go`, `/tmp/ctp-settlement-diagnosis`, `/tmp/ctp-settlement-baseline`, `/tmp/ctp-settlement-baseline-overlay.json`. No canonical-state dump or session credentials were saved with this report.

## Source trace

### Server enforcement

`internal/gameplay/engine/cities.go::foundCity`:

```go
tile, ok := tileByID(*state, unit.TileID)
if !ok || tile.OwnerID != "" && tile.OwnerID != actor {
    return errors.New("city site is not owned by founding player")
}
```

This rejects the actual empty-influence site solely because its national owner is foreign. It also does not express the released rule against settling inside an existing city's influence, including one's own city influence.

`internal/gameplay/engine/city_territory.go` explicitly distinguishes these concepts: `claimCityNationalBorders` can assign national ownership beyond city influence, while `refreshCityInfluence` assigns `InfluenceCityID`. The rejection is not evidence of corrupted or fictitious tile ownership; it is the wrong settlement predicate for a legitimate national border.

### Advertised actions and AI

`internal/gameplay/visibility/visibility.go::unitActions` advertises founding for an owned, untransported/unarmied settler on a tile without a projected city. It does not reconcile the site with the engine's national-owner rejection or the released city-influence restriction.

`internal/gameplay/ai/ai.go::projectedFoundingSiteAvailable` independently excludes foreign nationally owned sites. A future repair must reconcile engine, viewer action and AI predicates, not introduce a client bypass.

### Client request

`$CTPCLIENT/src/gameplay/runtime.rs::apply_authoritative_action` selects the exact `found-city` legal action for the selected server unit and submits its payload unchanged. The HTTP route accepts the matching `unit_id`, `city_id`, `name` fields. The existing isolated native settlement test confirms those identities and the name are preserved.

The client is following the server's advertised action. Disabling or overriding national-owner checks locally would either hide a legitimate original-game action or continue sending a rejected command; it would not repair server authority.

### Original released CTP2 authority

Local source root: `ctp2/ctp2_source_code/ctp2_code/` (not Apolyton).

- `gs/gameobj/UnitData.cpp::UDUnitTypeCanSettle` checks the settlement/city type, an existing city and supported terrain.
- `UnitData::Settle` calls `CanSettle`, rejects `GetCell(m_pos)->GetCityOwner().IsValid()` with `29IASettlingTooClose`, checks movement/first-move and terrain, then queues unit removal/city creation. It does **not** reject merely because `Cell::GetOwner()` is another nation.
- `gs/world/Cell.h` and `Cell.cpp` distinguish `GetOwner()`/`m_cellOwner` (national owner) from `GetCityOwner()`/`m_city` (city influence). `SetCityOwner` maintains the city-radius flag separately from an actual city center.
- `gs/gameobj/armyevent.cpp::ArmySettleOrderEvent` queues the army order; `ArmyData::Settle` queues `GEV_Settle`. `PlayerEvent.cpp::SettleEvent` calls `Player::Settle`, which selects the settling unit and calls its settlement method. These wrappers add no foreign-national-owner prohibition.
- `PlayerEvent.cpp::CreateCityEvent` calls `Player::CreateCity`, which rejects an existing/adjacent city (`World::IsNextToCity`), creates the unit and calls `World::InsertCity`. `gs/world/WrldCity.cpp::World::InsertCity` checks the actual city slot and assigns the new national owner; it does not require the prior national owner to match. The complete inspected creation chain therefore does not restore the server's blanket foreign-owner rule.

## Why the logs look this way

`service.engineRejection` wraps non-production/non-phase engine failures as HTTP422 `command_invalid`. `attachDiagnostic` intentionally logs a safe allowlisted reason, not arbitrary private error text. Therefore `reason=command_invalid` cannot distinguish this site error from other engine failures.

The client explicitly fetches a current snapshot after a rejected command. Repeated attempts at the unchanged illegal-under-current-server site produce repeated unchanged-revision refreshes; legacy/fault current reads remain rate-spaced. They do not mean a city was founded and lost, or that the turn failed to advance. The startup `add_ai` capacity retries completed before gameplay, and font fallback warnings affect rendering, not this rejection.

## Safe follow-up repair (not implemented here)

1. Write failing regressions for foreign-national/no-city-influence settlement and own/foreign existing-city-influence rejection, using the shared authoritative predicate and appropriate retail/legacy pin compatibility.
2. Use city influence/occupied/adjacent-city site constraints and actual unit/terrain/movement eligibility for retail settlement; retain authentication, ownership, phase, revision, atomic consumption and unique-city checks.
3. Align advertised actions and AI visible-site choices with that policy, without exposing hidden city locations or recomputing game rules in the client.
4. Add a bounded, viewer-safe settlement rejection category/message where appropriate; do not log arbitrary raw engine errors or return hidden city identities.
5. Verify current/synthetic/legacy saves explicitly. Do not edit live tile ownership, rewrite immutable rules pins, clear caches or fabricate a client-side city.

Until repaired, choosing a neutral/own-national tile outside existing city influence avoids this particular server guard. That is a workaround, not original-game parity.

Verification: the existing engine founding test and client `settle_uses_server_recommended_unique_identity_and_name_then_opens_that_city` both pass. The former currently encodes the foreign-owner prohibition, which is why existing green tests did not catch the retail mismatch. This documentation-only investigation adds no runtime changes or new executable tests; full builds/lints/coverage were not rerun for unchanged runtime code.
