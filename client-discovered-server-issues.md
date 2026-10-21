# Client-discovered server issues affecting CTP2 reference parity

## Current implementation status

The confirmed HTTP, future-queue progress, city/tile metadata and authorized trade-accounting defects now have server implementations and regression coverage. Atomic entry-addressed queues and supported city/Sleep/Wake/research-goal contracts are also available. The [v2 wire contract](docs/client-gameplay-contract-v2.md) and [live checklist](plans/checklists/client-report-contracts.md) describe exact semantics and verification.

Canonical world/terrain-aggregate bypasses are also retired. Invalid planning inputs reject before persistence; previously valid orders invalidated during resolution reject privately without partial changes or stalled phases.

**Current feature scope:** retained movement destinations, villages-only ruin rewards and city-production governor priorities are implemented authoritatively; see [the current contract](docs/client-gameplay-contract-v2.md). Barbarian ruin profiles, full empire governor population/army/public-works management, the exact policy optimizer and culture/style remain unsupported. Foreign history stays private. Client bindings/rendering/optimism/resync UI are separate work. Findings below retain the original pinned inspection as historical evidence.

## Original scope and evidence

This report documents server findings from the CTP2 client's reference-image and interaction work. Server revision inspected: **`73e4f53245bc891bcda856ed3d2000768281eb1d`**.

Findings below come from **read-only source inspection**, including the engine, viewer projection, HTTP gateway and existing tests. They have not been reproduced against a running server. Example scenarios are proposed in-memory test fixtures, not instructions to mutate a live game. No server implementation, tests, migrations, configuration or hooks were changed during that original investigation; later implementation is tracked above.

The earlier, separately named report [client-discovered-server-issue.md](client-discovered-server-issue.md) contains detailed river/resource/territorial-ownership findings. That file was unchanged during the original inspection; both reports now carry the current implementation status. This plural-named report originally added interaction findings and indexed the earlier metadata issues.

**Classification matters:** a gateway/projection inconsistency is a server defect; an operation not yet supported is a feature/contract gap; an unimplemented client window is neither. Not every unfinished client feature is blocked by the server.

## Findings at a glance (original inspected revisions)

| ID | Classification | Finding | Client impact |
| --- | --- | --- | --- |
| HTTP-01 | Confirmed gateway inconsistency | Viewer legal actions advertise routes absent from the HTTP command table | An enabled legal action can receive HTTP 404 without reaching the command service |
| QUEUE-01 | Confirmed progress-loss behavior | Cancelling any queue entry clears the city's stored production, including an entry behind the current item | Removing an unrelated future item destroys current production |
| QUEUE-02 | Queue-editing contract gap | Append/cancel operations cannot safely implement atomic Insert/Move Up/Move Down | Client-side cancel/re-add is destructive and can partially fail |
| QUEUE-03 | Deliberate current restriction / retail compatibility gap | A unit type cannot appear twice in a city's queue | Retail-style repeated unit orders cannot be represented |
| CITY-01 | Confirmed projection omission | Canonical city names are not sent or retained in city fog observations | City selectors, map labels, manager rows and messages cannot reliably display authoritative names |
| ORDERS-01 | Feature/contract gaps | No registered Sentinel/Sleep, city rename, city disband or mayor-control operation | Original controls cannot be implemented by silently substituting another command |
| TRADE-01 | Projection contract gap | Projected routes omit canonical good identity and route gold | Trade displays lack per-route commodity/income facts |
| MAP-01 | Previously reported metadata omissions | River, resource-good and territorial-owner facts are missing from viewer/fog contracts; normal generation also lacks river/good handoff | Original map/fog/border artwork cannot be selected faithfully from authoritative data alone |

## HTTP-01 — advertised legal actions are not reachable through HTTP

### Evidence

- `internal/gameplay/visibility/visibility.go`, `expelActions` (around lines 1360–1377), emits an action with route **`expel`**, fields `unit_id` and `target_unit_id`, and the exact payload identifying the acting and target units.
- `internal/gameplay/engine/engine.go` defines `CommandExpel`.
- `internal/gameplay/engine/special.go`, `expelUnit`, implements the operation; `special_task11_test.go` includes retail expulsion tests.
- `internal/gameplay/http/http.go`, the `commands` map (around lines 62–111), does **not** register `expel`.
- `Handler.ServeHTTP` (around lines 154–158) returns `http.StatusNotFound` when `commands[commandName]` is absent, before dispatch to the executor.

A static comparison of literal `action(engine.Command..., "route", ...)` calls in the viewer projector with the HTTP table found these advertised-but-unregistered routes at the inspected revision:

```text
assassinate-ruler
bio-terror
bombard
convert-city
create-franchise
create-park
expel
injoin
investigate-city
nano-terror
plague
plant-nuke
reform
settler-slave-raid
slave-raid
sue
underground-railway
upgrade-unit
```

These are conditional legal actions: the problem arises when the relevant eligibility conditions cause one to be advertised. This is not a claim that every game currently offers all of them.

### Example consequence

Given a viewer projection authorizing an expulsion by `agent` against `intruder`, a client submitting the advertised payload to `/game/{game_id}/expel` encounters the missing HTTP registration. Changing client artwork or selecting a different unit cannot fix that gateway inconsistency.

### Recommended server correction

Register supported engine commands with their exact accepted payload fields, authorization and normal command envelope. Prefer a shared command manifest, or a cross-layer validation test, so the viewer cannot advertise routes the gateway cannot dispatch. Do not weaken unknown-field rejection or let clients bypass the authenticated command service.

### Suggested tests

- Use `httptest` and mocked authentication/execution; advertise a legal expel fixture and verify the matching HTTP route reaches the executor with the exact acting/target identities.
- Cover all advertised route definitions against the gateway manifest, including optional rule-dependent actions.
- Verify stale revision, wrong ownership, forged identity, illegal target, unknown fields and idempotency behavior remain enforced.
- If a command is intentionally unsupported, do not advertise it as executable; document that restriction explicitly.

## QUEUE-01 — cancelling a future entry destroys current production

### Evidence

`internal/gameplay/engine/cities.go`, `cancelProduction` (around lines 1008–1021), searches for the requested `item_id`, removes the matching entry, and then unconditionally executes:

```go
city.StoredProduction = 0 // Supported default: cancelled production has no refund.
```

The reset is not conditional on cancelling the head/current item.

### Minimal in-memory scenario

1. Create an owned city with queue `[UNIT_HOPLITE, BUILDING_GRANARY]` and `StoredProduction = 40`.
2. Cancel `BUILDING_GRANARY`, leaving the Hoplite as the current item.
3. The current implementation leaves queue `[UNIT_HOPLITE]` but resets `StoredProduction` to zero.

The source establishes this behavior; no live cancellation was performed during this investigation.

### Recommended server correction

Separate removing a future order from cancelling/changing the currently produced item. Preserve current-item progress when an unrelated future entry is removed. Define and source-check the production-transfer/loss policy when the head actually changes; this report does **not** assume that every head change must preserve production without penalty.

### Suggested tests

- Removing a middle/tail entry preserves the unchanged head and its stored production.
- Cancelling the head follows the explicitly selected retail-compatible progress policy.
- Missing/unowned entries leave both queue and production untouched.
- Existing cancellation tests in `cities_task10_test.go` intentionally assert the current no-refund default. Extend coverage for non-head cancellation rather than merely deleting those assertions.

## QUEUE-02 — no safe authoritative queue insertion/reordering contract

### Evidence

- The HTTP gateway registers `queue-production` with `city_id` and `item`, and `cancel-production` with `city_id` and `item_id`.
- `queueProduction` ends with `city.Queue = append(city.Queue, payload.Item)`.
- There is no queue index/entry identity in these payloads and no registered insert/reorder/replace-queue operation in the inspected command table.
- Cancellation has the progress-loss behavior in QUEUE-01.

### Client impact

The reference Build Manager has **Insert**, **Move Up** and **Move Down**. Reimplementing these as a series of cancel/re-add commands is not an equivalent operation: it clears production, exposes intermediate states, advances revisions separately, and can leave a partially rebuilt queue after a failure. The client must not use that sequence as an apparently harmless queue reorder.

### Recommended server correction

Provide a validated, revision-checked queue-edit operation or equivalent atomic contract, with explicit production-progress semantics. Stable queue-entry identities would support precise selection, repeated unit orders and safe edits. The server must continue to validate ownership, production prerequisites, unit/building/wonder restrictions, costs and effects; the client should not become authoritative by sending arbitrary replacement rules.

### Suggested tests

- Insert before/after a selected entry; move an entry up/down; preserve all unaffected entries and order.
- Apply the documented progress policy exactly once when the head changes.
- An invalid edit is atomic: no partial queue, resource or production changes.
- Concurrent/stale edits conflict safely; replaying an accepted idempotent edit does not reorder twice.
- Command receipts and resyncs return the same resulting queue and selected-entry identities.

## QUEUE-03 — repeated unit types are rejected

### Evidence and classification

`queueProduction` (around lines 765–770) rejects any queued item whose `ID` matches an existing entry, returning `production item already queued`. `ProductionItem` uses the content/type `ID`; there is no separate queue-entry ID. Existing engine tests explicitly exercise duplicate rejection.

This is a **documented current restriction/retail compatibility gap**, not an assertion that the implementation violates its own existing tests. It prevents a client from representing a retail queue containing, for example, two separate Warrior orders.

### Recommended contract and tests

Distinguish queue-entry identity from production content identity. Permit repeated unit orders where the original rules allow them, while retaining appropriate building/wonder uniqueness rules. Test two identical unit types, independent selection/removal/reordering of each entry, completion of only one order at a time, and persistence/replay of their identities.

## CITY-01 — canonical city names are lost at the viewer boundary

### Evidence

- `internal/gameplay/engine/engine.go`, `City` (around lines 292–298), includes `Name` and `TypeID` separately from `ID`.
- `internal/gameplay/engine/cities.go`, `FoundCityPayload` and `foundCity`, require/store a name; the engine also checks for duplicate city names.
- `visibility.CityProjection` (around lines 95–117) has neither `Name` nor `TypeID`.
- `projectCity` (around lines 421–449) does not copy them, even for an owned city.
- `engine.FogObservation` (around lines 462–470) remembers city identity/ownership/effects but no city name. Remembered city projection construction likewise has no name to return.

### Client impact

After a receipt, resync, reconnect or encountering another city, the client cannot reliably label the city with its authoritative name. A client-local cache of the name submitted by the current process cannot cover other clients, existing games, enemy cities or renames. Rendering a machine ID as though it were the city name is not parity with the references.

### Recommended server correction

Expose the city name under an explicit visibility policy, including owned-city snapshots. If discovered city names are remembered under fog, store the last-observed name in viewer-specific memory and project that value for stale cities. A hidden rename must not leak through a fresh canonical read. Define legacy-observation behavior without inventing names.

Also determine which canonical city type/culture/era/capital facts the client may receive for original artwork and labels. The confirmed omission here is that `Name` and `TypeID` exist canonically but are absent from `CityProjection`; this report does not prescribe exposing unrelated private city state.

### Suggested tests

- Found a city with a non-ID name; verify the owned-city receipt, JSON round trip and resync retain it.
- Reveal a foreign city, lose sight, and retain only its last-observed name.
- Change the hidden canonical name; no update until permitted re-observation.
- Persist/replay observations and verify viewer isolation and legacy handling.

## ORDERS-01 — missing contracts for distinct original-game operations

The inspected gateway table has no registered operation for:

- Sentinel/Sleep as a distinct order from Fortify.
- Rename City.
- Disband City.
- Hiring/toggling a city mayor/governor or selecting its priority.

These are **feature/contract gaps**, not proof that all underlying rules are already implemented. Some special orders, such as Pillage/Reform, additionally have engine implementations but no corresponding HTTP registration; audit them with HTTP-01 rather than assuming there is no engine support.

The server does expose Fortify, unit Disband, research, social policy, tax/public-works rates, readiness, army composition and production operations. A missing client binding for those existing operations is a **client defect**, not another server issue.

### Required distinction

The client must not submit `fortify` when the user selects Sleep/Sentinel, or submit unit `disband` for Disband City. The context controls should remain explicitly unavailable until the appropriate authoritative semantics are supported. If original Sentinel/Sleep is intended to be purely client-side selection/cycling state in this architecture, document its lifecycle across rounds, reconnects and multiple sessions instead of conflating it with fortification.

Recommended tests for any new operation: exact target identity, ownership, eligibility/size/last-city constraints as appropriate, legal-action advertisement, revision/idempotency, persistence and visibility-safe receipts. Governor behavior must remain server-authoritative rather than a second client-side simulation changing real city results.

## TRADE-01 — route commodity and income facts are not projected

### Evidence and scope

- `engine.TradeRoute` includes `GoodID` and `Gold`.
- `visibility.TradeRouteProjection` (around lines 125–132) has route ID, endpoint IDs, water/piracy and transport-cost fields, but no commodity or route-gold field.
- `tradeRouteProjections` (around line 336) constructs the projected route without `GoodID`/`Gold`.

This is a **projection-contract gap for detailed trade UI**, not proof that every existing canonical route has a nonempty commodity or nonzero income. Current city goods and nation totals do not identify which commodity/income belongs to an individual route.

Define which per-route goods/income details are public, owner-visible, embassy-visible or intentionally undisclosed, and project only the permitted facts. Test zero/no-good states distinctly from unavailable information, both endpoints' visibility/ownership policies, piracy changes, and persistence/receipt consistency. The client must not infer hidden route details from terrain art or fabricate commodity assignments.

## MAP-01 — previously reported river/resource/territorial facts

See [the original detailed report](client-discovered-server-issue.md) for evidence, proposed contracts and regression cases.

The current source still has the important distinctions:

- `engine.Tile` has independent `River`, `GoodID` and `OwnerID` fields.
- `visibility.TileProjection` and `engine.FogObservation` do not carry those tile facts.
- `world.WorldTile` has coordinates and terrain only; the ordinary generated-world contract cannot carry river/resource placement into gameplay.
- City ownership and city-available goods do not substitute for per-tile ownership/resource coordinates.

Any correction must preserve viewer-specific last observations, never expose current hidden canonical changes, and keep server generation/versioning authoritative. Per-tile ownership/borders on never-discovered cells are forbidden even through a separate public endpoint; only permitted observed ownership may be projected. Adding visible projection fields alone does not create rivers/resources in the normal world-generation handoff.

## Optimize Sliders — clarified source behavior, not a claimed engine bug

The local original-release client source lacks the reference button's callback. The public **Apolyton Edition** source does provide it:

- [CauseAndEffectTab.cpp](https://github.com/civctp2/civctp2/blob/master/ctp2_code/ui/interface/CauseAndEffectTab.cpp), `OptimizeSlidersButtonActionCallback`: calls `Governor::OptimizeSliders`, then `Governor::SetSliders`, then updates city/resource displays. The file records the button addition in April 2005.
- [governor.cpp](https://github.com/civctp2/civctp2/blob/master/ctp2_code/ai/CityManagement/governor.cpp), `OptimizeSliders`, `TestSliderSettings`, `SetSliders` and `SliderTests::CalcTests`: evaluate trial policy settings using specialist reassignment, full city-resource calculations, happiness, food, production/readiness, gold/wages, government expectations and strategy thresholds.

The earlier statement that the button's behavior could not be deduced is therefore superseded: **the callback and evaluator are now located in the community source**. The remaining question is how to support equivalent behavior through this server's authoritative contract.

`set-social-policy` already applies policy values, but a mutation endpoint alone is not a read-only counterfactual optimizer. No matching optimizer or verified complete trial-evaluation contract has been established in this investigation. For example, canonical worked-tile IDs are not part of `CityProjection`. Do not repeatedly mutate a live game to probe candidate policies, or assume the AI's currently projected policy payload is this exact optimizer result.

A server-owned optimization command, or a visibility-safe read-only evaluation/recommendation contract using the actual engine calculations, is a recommended design direction—not a claim that an already-promised endpoint is broken. Specify specialist side effects, deterministic tie-breaking, prerequisites and revision behavior before implementing. Tests should use in-memory states/mocked stores, cover economic/happiness constraints and no-solution/empty-city cases, and verify that evaluation alone never mutates persisted gameplay.

## Interaction topology requirements

`$CTPCLIENT/IDEAL_INTERACTION_TOPOLOGY.md` governs every contract in this report:

- Server authority applies to singleplayer and multiplayer. On disagreement the client must show “resynchronizing”, revert speculation and replace its full authoritative model from the current viewer-safe snapshot.
- Ambiguous movement should animate optimistically before server confirmation, without inventing hidden terrain. The server validates every action; optimistic presentation never changes canonical authority.
- Never-discovered cells remain completely obscured. Terrain, goods, rivers, ownership, planning previews, notifications and route details must not bypass that boundary. Remembered cells use only the viewer's last observations, never fresh hidden facts.
- Hybrid is the primary multiplayer mode. Test planning/queued resolution, conflict components, inactive/component-external players, phase-appropriate pass/round-complete actions, revisions, retries and reconnects explicitly; sequential-only coverage is insufficient.

## Client responsibility and remediation checklist

Unimplemented manager/library/message windows, source-art parsing, cursor/animation rendering, stack selection, correct use of projected legal actions and consumption of metadata already sent by the server remain client responsibilities. These findings must not be used to label all unfinished client work as server-blocked.

The original handoff list below records reported scope. Use the current implementation checklist linked above for delivered work and explicit unsupported feature decisions:

- [ ] Reconcile advertised legal routes with the HTTP gateway and add cross-layer tests.
- [ ] Fix non-head queue cancellation so unrelated current production is not destroyed.
- [ ] Define/source-check atomic indexed queue editing and production-transfer semantics.
- [ ] Decide repeated-unit queue support and stable entry identities.
- [ ] Project authoritative city names and visibility-safe remembered names; define permitted city-art metadata.
- [x] Define distinct Sleep, city rename/disband and city-production mayor contracts, without claiming full empire governor automation.
- [ ] Define viewer-safe per-route goods/income projection.
- [ ] Address the earlier river/resource/ownership generation/projection/fog-memory chain.
- [ ] Establish a source-verified server-authoritative Optimize Sliders contract if that feature is supported.

Verification of this report: source symbols and line references checked against the revision above; no live HTTP/database requests and no server test execution were necessary for this documentation-only change.
