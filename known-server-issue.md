# Retail production: incomplete eligibility contract and opaque Add rejection

## Server resolution

The server contract gaps below are implemented. See
`docs/server-contract-repairs.md`, `docs/client-gameplay-contract-v2.md` and
`plans/checklists/server-contract-repairs.md` for current wire semantics and test
records. This delivery is **server only**; no client code or live games changed.

- [x] Complete authoritative owned-city `production_options`, including all
  eligible items on nonempty queues, with exact revision/city/hash binding and
  complete-empty versus unavailable semantics.
- [x] Shared read-only eligibility and execution checks; no per-candidate world
  serialization, RNG/events or queue identity allocation.
- [x] Precise safe rejection codes/messages/correlation and private redacted
  production diagnostics; credentials never appear in the new diagnostics or
  existing authentication-store logs.
- [x] Explicit pinned rules family and original resource-generation provenance.
- [x] Document atomic-append versus whole-edit boundaries and queued-versus-executed
  semantics without weakening validation or inventing forecast turns.

**The original live click was not reproduced, and no claim is made that changing
only the server fixes the client's catalog-as-permission or rejection UI defects.**
The investigation below is preserved as a historical baseline: its statements
about missing fields/old source behavior are superseded by the contracts above.

## Original investigation status and scope

**Confirmed source-level contract gaps at the investigated baseline; the exact
cause of the reported live “none of the units can be added” failure was not established.**

The report follows the switch from synthetic to retail rules: the client lists
retail units, but pressing Build → Add appears to resynchronize rather than add
anything. There are client defects **and** server changes needed to support a
correct, fully populated city Build manager. Do not attribute every failure to
the server, remove production validation, or disguise synthetic IDs as retail
units.

Investigated server HEAD `15ada0164180b8e6b8c280029ba55ef5cac88168` and client
HEAD `40bc68c1c30ff4bb792dbb31d4d3fcc67ffd8952`, including their working trees.
The server's existing `Taskfile.yml` retail-default edits and untracked checklist
and `server-client-schisms.md` were left untouched. The client's unfinished
selection/description work was not implemented or verified in this investigation.
Only this report was written; no server implementation/configuration, database,
game, installed asset, or client code was changed. No live server calls were made.

Related context: `server-client-schisms.md` and
`docs/client-gameplay-contract-v2.md`. Proposed contracts below are **not existing
API fields or promises of current support**.

## 1. What is established

### A. Client catalog membership is incorrectly treated as an Add candidate

Client `src/gameplay/ui_model/management/data.rs::buildables` enumerates all
pinned `rules.units` and `rules.production` records when the
`atomic_production_queue` feature is present. It also merges advertised
`queue-production` actions and existing queue content types.

Client `management/queue.rs::request` permits Add/Insert after checking the
city's generic `edit-production` action, queue size, and
`server/commands.rs::known_content`. These establish city edit access and known
content identity, **not that this city may currently build that item**.

Consequences:

- Locked, obsolete, special/non-buildable, wrong-government, or otherwise
  unavailable units can appear as ordinary candidates.
- A type already in the queue is not proof that another copy can be added now.
- A feature flag is protocol support, not authorization for every catalog ID.
- Showing all retail units after changing rules mode does not prove that any
  particular Add request is legal.

This is a client responsibility. It must stop presenting catalog membership as
current eligibility. Reimplementing private server buildability in the client is
not an acceptable repair.

### B. Existing server actions cannot populate a complete Build manager

Server `internal/gameplay/visibility/visibility.go::productionActions`
(approximately lines 1204–1279):

1. Emits a generic atomic `edit-production` action for an owned, non-stale city,
   retaining its current queue entry IDs.
2. Emits entry-specific cancellation and applicable rush actions.
3. **Skips all new-production candidates when the city queue is nonempty.**
4. For an empty queue, emits eligible non-unit production records, then walks
   unit definitions and **breaks after the first eligible unit**.

Outer `legalActions` also applies viewer/phase/planning authorization; a city
being owned does not override that gating.

Thus the existing candidate actions are not a complete per-city list. Simply
filtering the client to those actions would hide other legal units and remove
all new candidates once a queue contains an item. Conversely, the generic edit
action does not certify an arbitrary replacement payload. Its retention-shaped
payload must not be read as a blanket item authorization.

### C. Visibility eligibility duplicates, and diverges from, engine validation

Authoritative path:

`http.Handler.ServeHTTP → service.ExecuteCommand → engine.Apply →`
`editProduction (queue.go) → canonicalProductionItem → queueProduction (cities.go)`.

`editProduction` rebuilds the proposed queue in an isolated candidate. It
resolves new content from pinned rules and calls `queueProduction` for **each
entry, including retained entries**, restoring retained identities afterwards.
One invalid entry rejects the entire edit. Therefore a failed Add can also be
caused by an existing entry failing revalidation, not necessarily the newly
selected item.

`queueProduction` and `requireProductionUnlock` enforce more than the projection
filters. Concrete source discrepancies include:

| Case | Authoritative engine | Current projection candidate filtering |
| --- | --- | --- |
| Unit with `EffectHasPopAndCanBuild` | Rejects normal city production | No corresponding unit-loop exclusion |
| Create-park unit before an enabling wonder | Rejects | No corresponding park-ranger unlock check |
| Slave-raid/uprising production after applicable FreeSlaves/ProhibitSlavers wonder effects | Rejects | Exclusive-slavery helper checks owned unit effects, not this global wonder condition |
| Required city good supplied by an incoming trade route | `productionCityGoodSatisfied` uses `cityAvailableGoods`, including incoming route goods | `unitProductionBlockedByCityGood` uses `CityProjection.Goods`; `projectCity` copies raw `city.Goods`, not route-augmented goods |
| Building prerequisite supplied by a BuildingEverywhere wonder | `cityHasEffectiveBuilding` accepts the effective building | `cityBuildingIDs`/`allBuildingsSatisfied` use actual city building records |
| Already existing world wonder / nuclear plant after applicable elimination effect | Explicit rejection in `queueProduction` | Non-unit candidate loop does not perform those corresponding checks |

These are branch-level mismatches, not a reproduction of the user's game. Some
produce false positives; others hide legal options. Earlier eligible units and
the first-unit `break` can mask a later discrepancy in a particular fixture.

Existing engine tests explicitly cover the HasPopAndCanBuild rejection and the
park-ranger unlock rule (see verification below). Do not “fix” these mismatches
by weakening authoritative mechanics merely to match the projection.

### D. The server computes a rejection reason, then discards it at HTTP

`service/service.go::ExecuteCommand` wraps engine failure as status 422 with
`"gameplay command is invalid: " + err.Error()`. Stale revisions and other
failures have their own status paths.

However, `http/http.go::Handler.ServeHTTP` passes only
`service.StatusCode(err)` to `writeError`. `writeError` returns:

```json
{"error":"gameplay request rejected"}
```

It does not return the computed engine reason or a stable machine-readable
reason code. This is a **confirmed diagnostic limitation**, not merely missing
client log verbosity. Current client logs may show only HTTP 422 and this generic
body; asking for those logs cannot recover the discarded engine reason.

Do not solve this by exposing raw internal errors wholesale: some prerequisite
checks depend on information the viewer must not learn.

### E. Client recovery makes a rejected command look like a resync problem

Client `server.rs::post_json` includes the response status/body in its error.
`runtime/network.rs` retries certain ambiguous transport failures using the same
immutable command, then reports recovery and fetches authoritative state for
remaining failures. `runtime/sync.rs` maps failure/recovering outcomes to
`NetworkStatus::Resynchronizing` and a `Resynchronizing: ...` status message.

A valid rejection can therefore trigger visible recovery without indicating a
broken queue engine. Separately, finite event reads can be ordinary polling;
seeing a read/resync-like request after a click does not prove an Add POST was
sent or rejected. Correlate command ID, route, status and revision.

## 2. What the live symptom does NOT prove

The following remain unobserved for the failing click:

- Whether an `edit-production` POST was actually sent, and its exact entries.
- Its HTTP status, command ID, expected revision and authoritative game pin.
- The engine rejection branch, if any.
- Whether the receipt was accepted as a planning intent rather than immediately
  executed, or an unchanged queue was just observed after recovery.
- Whether this city legitimately has zero buildable units at its current
  population, advances, government, location and effects.

Population-one settler restrictions, missing advances, current phase/controller,
stale revisions, a full queue, malformed/cross-city identities, retained-entry
revalidation, rate limits, or an incompatible game pin are possible categories,
**not established explanations** for all units failing. The catalog alone cannot
be used as a positive-control test.

## 3. Server changes needed to set the client up for success

### Priority 1: one authoritative, complete owned-city production query

Provide a documented viewer-safe production-options contract in the matching
viewer snapshot (or another explicitly revision-bound authenticated interface):

- Complete eligible content IDs and kinds for each controllable city, including
  **all** eligible units, buildings, wonders and supported conversion types.
- Available even when the queue is nonempty; enforce capacity and duplicate
  non-unit constraints for the actual proposed operation.
- Exact revision, city identity, rules content hash and phase/planning context.
- Explicit completeness/availability semantics. Complete-empty must be distinct
  from unavailable, unsupported or partial; absence must not mean unlocked.
- Server-owned cost/forecast information only where actually supported. Unknown
  prospective turns must remain unknown, not be reconstructed from private rules.
- Optional stable disabled reason codes only when safe to disclose to this
  viewer. A generic unavailable reason is preferable to leaking hidden causes.

An illustrative name such as `production_options` would be a **new proposal**,
not a field the client may begin assuming exists today. Extending legal actions
is also possible if completeness and operation semantics are documented.

Share the authoritative eligibility implementation with command validation,
rather than maintaining another set of visibility predicates. The existing
`engine.Inspector` pattern is a suitable architectural precedent; it does not
currently expose this complete production-options query. Compile rules once and
keep inspection deterministic/non-mutating; avoid full-state serialization or
random/event/identity allocation for every candidate.

Importantly, **append eligibility is not authorization for every atomic edit**:
Insert, moving a funded head, retained-entry validation, non-unit duplicates and
whole saved templates depend on the complete ordered replacement. Specify this
boundary. Either expose revision-bound full-edit validation/preview or document
how clients submit an edit and receive a precise safe rejection. A preflight must
never replace final authoritative validation or become an execution guarantee.

### Priority 2: preserve safe rejection identity and diagnostics

Add a documented structured rejection envelope with stable code, safe message,
and correlation identity. Distinguish at least revision conflict, phase/controller
restriction, invalid queue identity/shape, capacity, content unavailable and
production prerequisite failure. Do not disclose protected hidden state through
overly specific codes or text.

Preserve the detailed internal reason in a private, redacted server diagnostic
path correlated with game/command ID; **never log session credentials**. Return a
safe correlation ID to the client. Update HTTP tests and contract documentation;
the generic existing `error` field can remain for compatibility.

Do not add blind retries, rebase a rejected command automatically, or change
idempotency identity. An old rejected action must not silently target a newer
city/queue snapshot.

### Priority 3: expose unambiguous rules provenance

`internal/game/start_gameplay.go::serverRuntimePinForSetupRules` uses synthetic
rules unless `OPEN_CTP_RULES_MODE=retail`. Retail additionally requires
`OPEN_CTP_RETAIL_GAMEDATA`. Compatibility version alone is not sufficient to
identify which rules content a game uses; keep/display the pinned content hash
and, preferably, a safe explicit rules-family/provenance marker.

A working-tree change already present in `Taskfile.yml` sets retail environment
values for `task run`; this investigation did not change it or verify a running
process. Restart the process with the intended configuration and start a **new
game** to verify a new pin. Do not expect an existing game's pin to migrate.
Retail loading already exists: server code changes are not necessary merely to
replace synthetic content in a newly started, correctly configured game.

## 4. Client changes needed alongside the server work

- Separate public catalog/Library and local Custom-template content from
  city-buildable production options. Custom authoring is not permission to
  execute a template against a city.
- Render/enable city Add and Insert only using the new authoritative eligibility
  semantics; retain exact revision, city and entry identities.
- Until that contract exists, a conservative advertised-action subset may offer
  limited functionality, but must be labelled incomplete. Do not claim the full
  retail Build workflow is restored by hiding all other rows.
- Keep Remove/reorder/Clear authorization separate from new-item eligibility;
  preserve server head-progress rules and explicit destructive confirmation.
- Display a persistent command rejection separately from synchronization state.
  Refresh when necessary, but do not erase the rejection on successful recovery
  or misrepresent all 4xx responses as transient network trouble.
- Preserve queued-versus-executed semantics. Simultaneous/hybrid planning may
  accept an intent without changing the canonical queue immediately. The server
  preflights planning commands in isolation (`engine.go::preflightPlanningCommand`
  and `planningCommandTrial`); receipt revision/acceptance is not execution proof.
- Never send client-authored costs/effects or implement private prerequisites
  locally. Keep immutable retry envelopes and idempotency keys.

## 5. Evidence to collect for the next failing click

1. Start a new verified retail-pinned game; capture viewer revision, rules hash,
   phase/planning status and the owned city's received queue/entry identities.
   Record received advances/population and advertised production/edit actions.
2. Capture exactly one Add interaction: whether a POST occurred, its route,
   redacted body, command/idempotency IDs, expected revision, HTTP status and
   response body. Remove session IDs/tokens before sharing logs.
3. Correlate the command to the engine failure on the server. With the current
   generic HTTP handler, this may require server-owner diagnostic instrumentation
   or an isolated in-memory fixture; the existing client cannot recover the
   discarded reason. No such instrumentation was added here.
4. Compare the failed replacement as a whole, including retained entries, with
   `editProduction`/`queueProduction`. Establish a genuinely eligible positive
   control from authoritative state, not the complete catalog.
5. Compare the returned receipt/recovery projection at its matching revision.
   For planning mode, inspect accepted intent and later resolution separately.
   Do not treat an unchanged immediate queue as evidence that acceptance failed.

Use in-memory fixtures and mocked HTTP gateways for regressions; do not mutate a
user's live game, write user files, or contact a live server in tests.

## 6. Server regression/acceptance requirements — completed

- [x] Empty owned city advertises multiple eligible unit IDs; each unchanged-state
  advertised atomic append passes authoritative validation.
- [x] Nonempty queue retains complete choices; duplicate units have independent
  entry identities and unchanged-head funding survives.
- [x] Projection/execution comparisons cover CantBuild, HasPopAndCanBuild,
  population, advances/obsolescence, government, imported city goods, park/global
  wonder restrictions, effective buildings, nuclear restrictions and duplicates.
- [x] Atomic edit regressions cover retained rows, reorder/head changes, batches,
  256/257 boundaries, explicit empty versus null/missing, invalid/foreign/missing
  identities and complete rollback, plus stale-revision service rejection.
- [x] Complete-empty and unavailable are distinct and revision/city/hash-bound.
- [x] Owner/phase/privacy and safe error/logging regressions protect foreign data
  and credentials; canonical world endpoints remain retired.
- [x] Existing command receipt/replay/planning/production-completion tests remain
  authoritative; no preview bypass, identity rebase or synthetic-as-retail alias.

Outside this server-only delivery: client mocked Add/rendering/rejection recovery
integration, and a server-owner live retail acceptance session. Neither was
performed or falsely marked verified; tests use in-memory/mock gateways only.

Retail UI authority remains the released source at
`$CTPCLIENT/ctp2_source/ctp2_code/`, not the Apolyton Edition:
`ui/interface/EditQueue.cpp` unit-list population (around lines 612–688) uses
`CityData::CanBuildUnit` for a city and separately handles no-city Custom mode;
`gs/gameobj/CityData.cpp::CanBuildUnit` checks advances, exclusions, CantBuild,
government, global/owned effects, script policy and movement/location constraints.
The original local script path is not permission to execute asset SLIC or infer
private eligibility client-side. Those semantics belong on the authoritative
server, with explicit supported/deferred differences.

## 7. Historical verification performed during the original investigation

Read-only source/call-chain investigation plus this focused existing test run:

```sh
go test ./internal/gameplay/visibility ./internal/gameplay/engine \
  -run '^(TestProductionActionsIncludeUnlockedRuntimeBuildings|TestProductionActionsHideUnitsNeedingAbsentCityGood|TestClientAtomicQueueReorderClearAndInvalidRollback|TestClientPinnedProductionCannotRedefineRules|TestRetailHasPopAndCanBuildUnitsAreNotCityProduction|TestRetailCreateParkProductionRequiresAnyEnableParkRangersWonder)$' \
  -count=1 -v
```

**All six selected tests passed**, using in-memory fixtures and injected engine
sources, without live requests or test file writes. They establish existing
validation/projection baseline behavior, **not** a passing cross-layer eligibility
contract or a reproduction/fix of the user's live rejection. No new regression
was implemented, no full suite/build/coverage was claimed, and the client's
uncommitted description changes remain unverified.
