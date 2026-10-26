# Viewer-safe gameplay contract v2

The client ideal interaction topology applies to local singleplayer and multiplayer.
The server owns canonical state, rules, generation and validation. This document
supersedes architecture suggestions in the two client-discovered issue reports;
their pinned source excerpts describe the older investigation, not current code.
See [server contract repairs](server-contract-repairs.md) for complete production
options, structured rejections, provenance, observed river geometry and turn/read
observability.

## Reading and reconciling

- Command receipts contain the complete viewer projection at their revision.
- `GET /game/{game_id}/events?session_id=...` with
  `X-Gameplay-Resync: current` returns a finite SSE `resync` response containing
  `current_revision`, durable `event_cursor`, and `snapshot`. Validate that the
  snapshot revision matches the envelope; resume using the durable cursor, not
  the revision. Resync SSE `id` is the durable event sequence. Legacy event
  streaming remains the invalidation/resume path; negotiated full-model streaming
  below avoids a second current read for every AI update.
- On disagreement, show **resynchronizing**, revert tentative presentation, replace
  the full authoritative model from a current viewer snapshot, then resume input.
  Do not patch invented geography into an empty snapshot or read canonical events.
- Retry a lost command response with the same envelope/idempotency key; do not
  generate a different purchase/edit/order. Never regress an already newer model
  by blindly applying an older retried receipt.
- **`GET /world/{game_id}`, `/ratings` and `/pollution` are retired (410)** after
  authentication/membership checks. The old aggregates counted all hidden terrain;
  they were not actual viewer-safe nation statistics or observed tile pollution.
  Canonical maps, seeds and foreign spawns are not client APIs. There is no
  omniscient fallback, including for the game owner or local play.
- Command bodies are bounded at 256 KiB; oversized requests return 413 before
  command execution. Shared and command-specific field validation still applies.
- Gameplay rejections include safe `code`, `message`, and `correlation_id` alongside
  `error`. A persistent rejection and a later successful refresh are separate
  facts. Ordinary current snapshots are replication, not proof of fault recovery.

`contract.version` is `2`. `contract.features` identifies supported contracts and
explicitly unavailable features. A feature flag is not phase/entity authorization:
consume `legal_actions`, `planning_status` and current revision checks as well.

## Low-latency complete snapshot streaming

When `contract.features.snapshot_stream` is true, a client may subscribe to the
same authenticated events endpoint with `X-Gameplay-Stream: snapshots-v1` (without
`X-Gameplay-Resync: current`). The server acknowledges the header and sends a
complete viewer-projected `resync` at connection time and on committed-state
wakeups. Each envelope is `{current_revision, event_cursor, snapshot}` and its SSE
`id` equals `event_cursor`. Both numbers come from the same committed aggregate
row. Terminal delivery remains viewer-projected. No canonical data/event patch or
foreign private model is exposed, and subscribing still precedes the initial read.

Subscription connections have an independent 20/minute budget: bootstrap does not
consume their first token. Legacy subscriptions/current reads keep their original
shared budget. Healthy full-model updates on an existing connection consume no
additional HTTP connection admissions. Reconnects remain bounded; durable outbox
recovery and periodic AI discovery retain restart/missed-wakeup safety.

Clients validate envelope/cursor/revision/turn before delivering complete models.
They may coalesce intermediate full snapshots, but must never coalesce away own
receipts/rejections or treat a stream model as a command acknowledgement. Keep
monotonic revision reconciliation and replace the whole viewer model. The current
client bounds negotiated snapshot data at 16 MiB; incomplete/oversized/malformed
models fail closed into safe recovery. Missing capability/header support retains
legacy invalidation/current behavior instead of inventing locally advanced turns.

Interactive command admission is bounded at 30 requests/second per IP; requests
still pass every authentication, ownership, phase, revision and transactional
command check. Rapid ordinary moves/turns need not wait for the former two/second
limiter. Background reads must not monopolize command submission, even when a
legacy/fault-recovery read is waiting for its own admission slot.

Performance and unchanged-gameplay evidence: [near-instant turns](instant-turns.md).

## Fog, independent facts and legacy data

Never-discovered cells are absent. Public bounds are not permission to fill them
with terrain. Ownership/borders, previews, notifications and trade data must not
provide an alternative hidden-map lookup.

Visible tiles include explicit `river`, `good_id`, `ruin` and territorial `owner_id`
facts. `false`, `""`, `false` and `""` mean known no-river/no-deposit/no-ruin/neutral. Omitted
metadata in legacy observations means **unknown**, not an authoritative negative.
Each remembered tile uses its viewer's last observation, observation turn and
`stale: true`; hidden canonical changes are not copied or recomputed into it.
Rivers are independent of base terrain. When `observed_river_geometry` is true,
`river_geometry` carries the legitimate center tile's original SW/NW/NE/SE art
edges and beach-mouth discriminator, retained under fog. This does not project
neighbor cells. Missing geometry is unknown, not zero connectivity; a renderer
must not treat unknown neighbors as no-river or request hidden cells to finish art.

Inspectable tiles also carry:

- `yields: {food, production, gold}` with explicit zeroes;
- `yield_context`: `terrain` or `city-center`, before city/national policy effects;
- optional `movement_cost`: the engine's general terrain/river/city/improvement
  cost, **not** a guarantee about a particular unit, domain, access or enemy ZOC;
- legally observed pollution and public improvement kind/ID/construction facts.

These display facts survive fog. Unknown legacy yields/costs remain omitted.
Only completed improvements affect yields/costs. Goods do not invent yield effects
that the pinned engine does not apply. Unit-specific planning is currently-visible
only; undetected stealth units do not become ZOC preview oracles.

Initial vision is persisted at game start. Accepted commands preserve both
pre-transition and post-transition sight, so moving away cannot erase first sight.
A legacy current view can be recorded; a legacy hidden observation cannot be
backfilled from hidden authority.

New games use `ctp2-original-goods-v4`: V1 terrain/spawns, V2 rivers and V3 ruins,
with resource placement pinned to `ctp2-original-resources-v1`, not four fabricated
variants per terrain or installed GOG extensions. Public `world.generator_version`
and `world.resource_catalog` identify this policy without exposing seeds. V3
selects at most one land ruin per complete 6x6 box, excluding starting coordinates;
private reward rolls are persisted at game start. Older versions/saves are not
regenerated. Later terrain changes may legitimately retain deposits.
Terrain families are retained through startup; unsupported new terrain rejects
rather than silently flattening deposits onto plains. Synthetic rules are explicit
bounded fixtures (`rules.family: synthetic`), not a retail numerical/artwork
substitute. New retail pins encode `rules.family: retail`; unlabelled old runtime
pins expose `unknown`. Keep the exact pinned `rules.content_hash`.

### Retail starting discoveries

New retail pins include versioned starting-advance policy, initialized before
new-game persistence. Original DiffDB human/AI chances and caps apply in source
order; guaranteed grants bypass the cap (but count toward it). Default starting
settings are medium/first age; owner setup may use `difficulty:<id>` and
`start-age:<AGE_ID>` in `rules_json`. Public `rules.starting_difficulty` and
`rules.start_age_id` identify this pin selection, not permission to infer grants.

All six source profiles guarantee Toolmaking, so a fresh population-one city can
build Warrior. Settler remains unavailable at population one; after authoritative
growth to population two it becomes an eligible option. Consume actual complete
`production_options`, not a client population/tech approximation.

Later-age starts grant strictly earlier-age discoveries instead of first-age chance
rolls. A legal initial research target is selected after grants. This only applies
to new games: old saves/pins are not rewritten or rerolled. See
[retail startup contract](retail-startup-contract.md) for source semantics, supported
settings, deterministic RNG, restart behavior and difficulty scope boundaries.

## Optimistic movement and hybrid play

An `action_intents` entry authorizes **input shape**, not a hidden outcome. A move
intent identifies only an owned unit, route `move`, and fields
`unit_id`, `target_x`, `target_y`. It contains no hidden target ID or passability.
The client supplies an adjacent public coordinate and may begin provisional
animation immediately, without guessing its terrain. Submit through the normal
command envelope, for example:

```json
{
  "session_id": 1,
  "expected_revision": 7,
  "idempotency_key": "move-7",
  "command_id": "move-7",
  "unit_id": "scout",
  "target_x": 3,
  "target_y": 4
}
```

`move` accepts either `path: [canonical tile IDs...]` or the coordinate pair, never
both. Both coordinates are required and cannot be null. Coordinates resolve only
inside the engine; movement, adjacency, domain, access, occupancy, fuel and ZOC
checks remain authoritative. All planning actions are validated against an
isolated baseline before persistence and revalidated at resolution. Invalid inputs
change no authority. A previously valid order invalidated by earlier ordered
results is discarded atomically and produces a private `order.rejected` notice;
resolution continues instead of stalling. Corrupt/invalid legacy stored envelopes
still fail closed. Terminal results stop later orders and economic recalculation.

Hybrid is the primary multiplayer mode:

1. Open planning permits actor actions and `round-complete`.
2. An actor with queued planning sees only `round-complete` for closeout.
3. Closed/locked/inactive actors have no executable actions or movement intents.
4. Conflict permits `pass` only for the current eligible component actor, along
   with that actor's permitted actions. Component-external players wait.
5. Resolution returns to the next planning round with monotonically increasing
   revisions. Current snapshots and receipts use the same viewer projector.

A queued receipt acknowledges intent, not completed movement. Keep provisional
animation separate from the canonical model. Do not infer a resolved outcome from
path artwork or replay stale segments next round.

### Retained movement

When `retained_movement_orders` is true, `set-movement-order` takes `unit_id`,
`target_x`, `target_y` plus the shared envelope. Both coordinates are required,
non-null public map coordinates; unlike `move`, the destination need not be
adjacent or reachable this turn. It applies an affordable prefix immediately and
persists the destination for the exact unit/army membership. It can be accepted
with zero current movement. In simultaneous/hybrid planning the receipt queues
intent; execution occurs at resolution, not at preview time.

Owner-only `movement_orders` survives reconnect and includes unit/army/member IDs,
expected current tile, destination coordinates and originating `command_id`.
`cancel-movement-order` takes `unit_id` (any member cancels the stack's journey).
A new destination replaces the old one. Successful manual unit/army commands,
changed membership, ownership or displacement cancel the old intent.

Once per new round, after movement refresh and in stable unit-ID order, the server
continues remaining journeys. The route steps diagonally toward the destination,
then along its remaining axis; it does not search hidden terrain or route around
obstacles. Every step reuses authoritative movement/access/domain/fuel checks and
uses the slowest member's budget. Exhausted movement retains a feasible order;
blocked/occupied/permanently unaffordable steps stop it with a generic private
notice. Entering enemy ZOC stops the journey, except arrival is completion.
There is no automatic combat or attack authorization. To detour, submit a new
waypoint/destination. No order progresses after terminal resolution.

`movement_order.completed` / `.stopped` notifications identify the unit and original
command, without revealing hidden obstruction details. Set/cancel, movement,
rewards, receipts and notifications persist in the same command transaction;
retrying an idempotency key cannot advance again.

## Atomic production queues

Queue rows have server-assigned `entry_id`, separate from content/type `id`.
Legacy rows receive deterministic identities that the next accepted mutation
persists; per-city monotonic sequences prevent reuse after removal/completion.
Repeated unit content is allowed; building/wonder restrictions remain validated.
The existing lone-unit standing-production behavior remains unchanged.

`POST /game/{id}/edit-production` uses the ordinary revision/idempotency envelope
and a **complete ordered replacement selection**:

```json
{
  "city_id": "home",
  "entries": [
    {"entry_id": "home:queue:1"},
    {"item_id": "UNIT_WARRIOR"},
    {"entry_id": "home:queue:3"}
  ]
}
```

The example omits the shared envelope for brevity. An entry selects exactly one
existing entry ID or one canonical new content ID. It cannot redefine cost,
effects, prerequisites, kind or identity. Existing entries must belong to the
specified city and appear at most once. New content uses pinned server rules.
Explicit `entries: []` clears; omitted/null entries reject. The maximum is 256.
Insert/reorder/remove/clear/custom replacement are one validated transaction, not
cancel/re-add sequences. An invalid edit leaves the entire queue/progress unchanged.

**Progress policy:** an unchanged head retains stored production, including when
future orders are removed/reordered. Replacing/removing an existing head clears
progress once (the existing supported no-refund policy). Starting an empty queue
preserves unassigned/carryover production, just like append. This is server behavior,
not a claim that every retail editor/category-switch path has identical penalties.
Retail `CityData::ChangeCurrentlyBuildingItem` is distinct from editing future rows.

`cancel-production` accepts `city_id` and `entry_id`. Legacy `item_id` remains
accepted only when it identifies exactly one row; ambiguous duplicate content
rejects. Do not send both. Legacy `queue-production` remains available, but pinned
games resolve content metadata server-side. Unknown client-defined items reject.
Sparse synthetic catalog costs use the server's documented one-point default,
not a client-selected cost. Keep client row selection keyed by entry ID.

### Complete append eligibility

`production_options` is an owned, currently observed city's complete revision-,
city- and content-hash-bound atomic-append query. `status: complete` with `items: []`
means no append choices; `unavailable` is not complete-empty or unlocked. All
eligible units/buildings/wonders/conversions are considered, including nonempty
queues. This uses execution validation, not projection-local predicates.

Retaining the ordered queue and adding one advertised item is the documented
operation. Arbitrary Insert/reorder/head changes/templates still require whole-edit
validation and can reject atomically with a safe code/correlation. Costs are
server-owned; no prospective turns are invented. See the supplement for exact wire
fields, phase/controller boundaries, unavailable semantics and error codes.

## Territory, influence and player identity

Observed tiles expose national `owner_id` and distinct `influence_city_id`.
An explicit `""` is known unowned/no city influence; omitted older fog facts are
unknown, not empty. Fog retains last observed values, never hidden current changes.
New retail pins claim source-driven national/influence radii on founding and update
influence on growth/shrink/capture/removal. Do not calculate assignments client-side.

Use `viewer_player_id` and public `players[]` (`id`, `nation_id`, optional
`civilization_id`, optional persisted `color_index`) for identity and palette slots,
not surviving units. Client RGBA/political/dashed influence rendering remains client
work. [Exact source/topology/legacy semantics](city-development-contract.md).

## Cities and distinct orders

- Owned/visible cities expose real `name`/`type_id` and observed population.
  Foreign economy, queue, specialists and purchase quotes remain private.
- `art.age_id` is the current owner's canonical age; `art.capital` uses effective
  capital rules. These facts, name/type/population are remembered only as observed.
  They are not guessed founder culture or a guarantee of a particular sprite.
- New retail pins import original civilization city-name lists from
  `civilisation.txt` and English `civ_str.txt`. Founding with omitted/empty name,
  human suggestions, AI and ruins use the selected civilization's ordered names
  and only own-name collision checks. Legacy pins retain neutral `City N`; existing
  saved/custom/foreign names are not overwritten. See
  [city development contract](city-development-contract.md).
- `rename-city`: `city_id`, `name`; trims whitespace, rejects empty/control/overlong
  names and enforces ownership. Names are at most 128 Unicode codepoints.
  Duplicates are allowed: city IDs, not names, identify cities. Checking global
  name uniqueness would create an undiscovered-city-name oracle.
- `disband-city`: `city_id`; requires another owned city, permitted population
  under pinned `MAX_DISBAND_SIZE` (legacy default 3), and a real compatible settler.
  It creates that exhausted settler, removes the city/routes and destroys its
  wonders. Missing settlers reject instead of the original assertion/destruction.
- `sleep` / `wake`: `unit_id`; retained owner-private state distinct from Fortify.
  Sleep does not grant fortification. Explicit accepted movement wakes movers;
  Fortify also clears sleep. Cargo/in-space sleep changes are not advertised.

### City governors / mayors

When `governor` is true, `set-governor` takes `city_id`, explicit non-null `enabled`,
and an enabled priority from `rules.governor_priorities`. Supported pinned focuses
are growth, production, gold, science, happiness, defense and offense; a pin lacking
the corresponding retail building/wonder lists does not advertise that focus.
On first enable, a priority is required; subsequent calls may retain it. Disable
uses `enabled: false`. Owner-private city `governor: {enabled, priority}` survives
reconnect; nil means no governor. Foreign observations never include settings.

The governor fills an empty production queue on enable and during each city round,
including after a completed build. It selects the first legal item in the pinned
focus building list, then other focus lists, then wonders in the same focus order.
All ordinary production restrictions/costs and stable queue identities apply.
Manual nonempty queues and stored progress are never replaced by a priority change;
disabling keeps existing production. Conquest/revolution resets the former owner's
governor. `governor.production_selected` privately identifies city/tile/content.

This is autonomous **city production** management, not full original empire AI:
it does not reassign specialists, manage armies, spend public works, rush-spend gold
or optimize national sliders. Culture/style remain unavailable.

## Quotes, accounting and research

Owned cities expose revision-bound `rush_buy_quote` separately from executable
payloads. `rush-buy` still accepts only `city_id` plus the command envelope.
Never put a display `cost` into that payload or let a client choose the price.
Quote statuses are `available`, `insufficient-gold`, `completed`, `empty-queue`,
`not-rushable`, `unavailable`. An unrepresentable/nonpositive unfinished purchase
price is unavailable, never a negative price or a completed build. Here `completed`
means no production remains to fund; it does not
create a unit/building before the production pipeline actually completes it.

Owned city `production_forecast` and own research `forecast` use effective engine
calculations, including public-works rounding and the execution research minimum.
`assumption: constant-current-economy` is explicit: growth, future builds/orders,
losses and other players' decisions can change an estimate. `estimated` carries
known rate/turns; `stalled` has no finite turns; `completed` has zero remaining
production/research; `no-target`, `continuous`, `unavailable` are not invented zeroes.
These are display estimates, never authorization to spend or execute a future turn.
Missing empire forecasts remain unavailable; consume current reserves/yields first.

Trade route `good_id` and explicit `gold` are endpoint-owner accounting, available
only when both endpoint cities are currently visible. Other viewers do not get
private commodity/income. Gold is the stored **gross route contribution**, before
piracy/national allocation/other economy effects, not a guaranteed treasury payout.
No commodity fact authorizes revealing its original deposit coordinates.

`set-research-goal` takes `advance_id` (explicit `""` clears; omitted/null rejects).
The owned `research.goal_id` survives reconnect. Goals route through legal pinned
prerequisites in deterministic prerequisite-ID order, respect end age, and clear
when reached. Changing goals preserves progress if current research is unchanged;
clearing keeps current work. Manual `set-research` replaces goal routing. Cyclic,
unknown, completed or unreachable goals reject. No foreign research is exposed.

Optimize Sliders remains **unsupported** (`policy_optimizer: false`). Existing
`set-social-policy` is a mutation, not a trial evaluator. Do not search a live game
by repeatedly submitting policies or pretend an AI policy suggestion is the
community governor's exact optimizer. Implementing that feature still requires its
counterfactual specialist/happiness/economy objective and apply semantics.

## Calendar, messages and history

`world` publishes stable `min_x/min_y/width/height` independently of discoveries.
Both wrap flags are false because gameplay adjacency is currently bounded, even
though legacy spawn scoring uses wrapped distance. No terrain cells are added to
size a camera. Invalid/unrepresentable coordinate extents are omitted, not wrapped
into zero or negative camera dimensions.

When the pin contains timescale data, `calendar` includes version
`ctp2-timescale-v1`, current turn and authoritative integer year. BC/AD/localization
is client formatting. New retail pins map original round zero to server turn one
and adjust their terminal-turn threshold accordingly. Old pins are never rewritten;
a pin lacking timescale metadata yields no date, not a hardcoded 4000 BC.

`notifications` contains only the viewer's authoritative semantic results:
`production.completed` includes city/tile/content IDs and an optional produced unit
ID; `order.rejected` includes the original `command_id` without disclosing why a
hidden interaction invalidated the order. Handle rejection by dropping tentative
presentation and following the **resynchronizing**/full-replacement workflow.
Notices have stable IDs and generation turns. Deduplicate across receipts/resync/
retry. A production notice may reference a formerly owned city;
it is a historical result, not fresh knowledge about its current foreign state.
Keep read/unread state local. Retain at most 128 results **per recipient**; older
unavailable notices must not be fabricated or rerun to grant rewards.

`statistics` contains up to 256 actual completed-turn samples for the viewer's own
nation: turn, engine score-victory score, city/unit count and population. Missing
older samples/legacy history are unavailable. A sample is labeled with the round
actually completed, not the scheduler's already-advanced next-round clock. Foreign
rankings/history are private.
Do not label this bounded own history a complete omniscient whole-match archive.
### Ruin discovery and rewards

`ruin_rewards` advertises authoritative discovery on successful movement (including
intermediate atomic path tiles), disembarkation or combat entry. Sight alone does
not open a ruin. The first explorer consumes it in the same transaction; a rejected
atomic path consumes nothing, and revisiting/replaying cannot award twice. Founding
a city on an unopened ruin removes it without a reward, matching city construction.

The persisted `villages-only-v1` profile follows released `risks.txt`
`RISK_VILLAGES_ONLY`: 30% gold (150–299), 15% eligible advance, 25% land mercenary,
5% city, 5% starting land settler, 20% empty. Advance/unit eligibility uses pinned
content and the source prerequisite limits (1 / 2 unknown advances); unavailable
or unsafe rewards become empty rather than crashing. Advance age/feat/research
handling and unit creation bonuses are authoritative. Unit rewards respect stack
capacity. Barbarian-risk profiles are **not implemented**.

A private `ruin.explored` notice has a stable ID, turn, tile, `reward_kind`
(`gold`, `advance`, `unit`, `city`, `empty`) and the applicable `amount`, `content_id`,
`unit_id` or `city_id`. Hidden rolls and eligibility candidates are never projected.
Visible `ruin` presence becomes a stored fog fact; another player's hidden discovery
does not update stale memory. Screenshot fixtures must never create real rewards.

### Committed turn/read observability

`turn_context` exposes matching revision/turn, mode, committed scheduler status,
and the public active actor in sequential/long-turn modes. `live_progress` is false;
hybrid private participants/orders are not exposed. `unit_route_preview` and
`live_resolution_progress` capabilities are explicitly false. Do not turn general
cost/destination intent into a green unit-route guarantee or fabricated nation strip.

Command `Server-Timing` measures gateway execution. Event `X-Gameplay-Read-Mode`
distinguishes `current` from `stream`; initial-read timing is separate. Failed
initial reads return safe 503 before success headers. Streaming subscribes before
its durable initial read to prevent lost commit wakeups. No retry/idempotency/cursor
or event-rate-limit semantics are changed; the old resync-ID/revision confusion
is corrected to use a durable `event_cursor`. Private injected command metrics separate
lock/load, engine resolution, visibility and persistence; they never enter receipts.

## Remaining client work and feature boundaries

The server work does not implement client windows, art decoding, queue selection
migration, event consumption, optimistic animation, rollback, the resynchronizing
banner, complete-model replacement, forecasts/quotes UI, or notification graphs.
Update client decoders to retain nullable/known metadata and consume contract flags.
Use installed asset catalogs locally; no copied art, SDL handles or sprite pixels
are server facts.

Explicitly unimplemented feature requests remain backlog, not fixed defects:
the exact slider optimizer, culture/style modeling, barbarian ruin profiles and
full empire-wide governor population/army/public-works automation. Foreign statistics are a
permission decision, not missing client data to work around. Do not advertise or
fake any of these features. See `plans/checklists/client-report-contracts.md` for
red/green/refactor evidence, original-source references and verification status.
