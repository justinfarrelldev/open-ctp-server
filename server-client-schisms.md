# Server/client schisms reported from current client screenshots

User-requested diagnostic report only. Server inspected read-only at `15ada0164180b8e6b8c280029ba55ef5cac88168`; no server implementation/configuration/assets/hooks changed. Client worktree: `/home/justi/Projects/ctp-client`. Implementation ownership below is a proposal, not proof that every symptom is a server bug. These follow-ups are lower priority than the client's current original-source UI/reference-screen wave.

## 1. Fog-boundary rivers look like enormous pools/lakes

### Evidence and report

The user reports that river lake/end shapes look massive while the rest of the river has not yet been revealed. Current-client evidence (do not copy into server assets):

- `images/current-client-game-screen-oct-22-2026.png`
- `images/current-client-game-screen-2-oct-22-2026.png`
- `images/current-client-ruin-oct-22-2026.png`

The latter shows several broad water/pool shapes around the revealed frontier; further river segments are absent behind fog. Compare with the local **original release** `ctp2_source` river rendering/`TryRiver` and original-game references, not the modified civctp/civctp2/Apolyton code. No live snapshot/latency trace was captured for this report, so generation versus observation versus compositor contributions are not yet fully isolated.

### Concrete source mismatch candidate

- Server `internal/gameplay/visibility/visibility.go::TileProjection` projects `River *bool` only (`river,omitempty`). Visible tiles copy `tile.River`; remembered tiles copy `observation.River`. `engine.FogObservation` stores the observed river presence, not an explicit observed river-art connectivity/shape fact.
- `docs/client-gameplay-contract-v2.md` correctly distinguishes absent/unknown versus known `false`, and explicitly says the renderer must not request hidden neighbors merely to finish a river mouth.
- Client `src/gameplay/server.rs::ServerTileProjection` retains optional river presence; `model/server_projection.rs` only marks known received river tiles.
- Client `src/gameplay/runtime/textures.rs::terrain` currently computes a9-cell `river_neighbors` array with `is_some_and(|tile| tile.river && tile.fog != Hidden)`. Thus unknown/unprojected/hidden neighbors become the SAME `false` used for known no-river.
- Client `src/gameplay/assets/tile/rivers.rs::river_piece` exact-matches SW/NW/NE/SE boolean connectivity against installed original overlays. A known center whose connections are hidden may therefore select an isolated/pool/end overlay, not the shape observed on the actual known tile. This is a concrete client information-loss path, not yet a verified defect in map generation.

### Responsibility / privacy-safe resolution options

Do NOT reveal hidden tile IDs, terrain, deposits, ownership, units or full geography to patch artwork. The client must not treat unknown neighbors as proven absence or procedurally invent continuations.

A potentially needed server fact is the **viewer-safe observed river geometry of the already-visible tile**: versioned shape/art discriminator, or an edge-connectivity value with explicit known/unknown semantics and observation revision/turn. Names here are proposals, not an implemented API. It must represent information legitimately visible in that tile's original artwork, not an unrestricted hidden-neighbor map query. An alternative is a documented client-side partial-information art policy if it can faithfully reproduce the original without inventing geometry.

For remembered fog, retain the last legally observed geometry. Do not recompute remembered shapes from unseen current-world changes or re-read private river neighbors during rendering. Preserve actual isolated source/lake/mouth cases separately from incomplete knowledge.

### Acceptance / next owner actions

- [ ] Reproduce a deterministic straight/joining river with center visible and one/both connected neighbors unprojected; compare to the same center fully observed and to a genuinely isolated river/lake.
- [ ] Confirm which shape details the original release permits a viewer to see on a known tile; decide projection versus partial-information renderer responsibility.
- [ ] If needed, add a versioned observed shape/connectivity contract with unknown distinct from absent/false and safe bounds; client report does not implement server changes.
- [ ] Remembered tile shape must stay fixed across hidden upstream/downstream changes until it is observed again.
- [ ] Client may then select installed original art from received legal geometry; no hidden-cell request, local world generation, copied art or guessed continuation.
- [ ] Tests cover visible/partial/remembered/hidden/reveal, source/mouth/junction/isolated, actual lake versus unknown river, both projection and in-memory SDL negative controls. Confirm false broad pools disappear without leaking hidden geography.

## 2. City names expose technical identities (related screenshot finding)

The same captures show names such as `City unit:...settler...` / AI account-unit identifiers in map captions and selectors. This is not merely a visual clipping issue:

- `visibility.go` currently advertises `found-city` with `city_id: "client-city-" + unit.ID` and `name: "City " + unit.ID`; `name` is advertised mutable.
- `visibility.go::projectCity` transmits actual `city.Name` (including remembered `CityName`). `engine/ruins.go` also has a `Name: "City " + id` fallback.
- The client should preserve identity separately from display names, and use authorized naming/rename input or an authoritative original-style civilization naming policy. It must not silently rename all cities Rome or replace foreign/private names with invented strings.

- [ ] Agree on source-accurate default city naming / viewer-safe suggested name or explicit client naming prompt under current mutable payload fields; preserve stable entity IDs and uniqueness rules.
- [ ] Check found/renamed/ruin-created/AI/remembered cities; no technical-ID fallback masquerading as an original civilization city name. Client map/panel/manager clipping still needs independent work.

## 3. Slow end-turn / acting-nation progress observability (unmeasured)

User confirms multiple turns without crashes, but says passing turns is painfully slow. They request the original small colored strip below the turn/date counter showing which nation is currently taking its turn. No wall-clock benchmark or cause is asserted here.

The client must distinguish enqueue delay, command HTTP duration, server planning/barrier/AI resolution, finite snapshot-read admission and presentation. Its shared SSE/current-read budget is3100ms for the server's20/min event limit; normal command HTTP is not directly gated. However current reads and commands share a worker, so a read-budget wait can potentially cause head-of-line delay for a newly queued command. This is a **client-side candidate**, not evidence the server itself is slow.

- [ ] Collect bounded/injected timings for each stage without secrets/private game-state logging; isolate client scheduling from actual server resolution.
- [ ] Determine whether existing viewer phase/events expose a legally displayable acting/resolving nation and progress. Simultaneous planning is not automatically a sequential nation turn, and a resolver holding a world lock may not publish mid-resolution progress.
- [ ] If missing, propose safe revision/turn-bound public progress facts; do not expose AI orders, hidden positions, private plans or fabricate a nation-color sequence in the client.
- [ ] Implement original strip geometry/color/update behavior client-side only once supported; explicit unavailable/waiting state otherwise. Keep command receipt versus queued intent versus executed turn distinct.

## 4. Impossible-looking settler crossing / unverified route drawn as legal (client candidate)

Newest evidence: client `images/current-client-impossible-move-client-bug-oct-22-2026.png` shows a land settler with a white/green turn1 path through water-looking cells toward a city. User reports an impossible ocean crossing. Without the corresponding raw viewer snapshot, distinguish actual ocean terrain/permission from section1's river-pool artwork; neither possibility makes the false-looking route acceptable.

Concrete client paths:
- `runtime/movement.rs::path_preview_between` checks the target, but `computed_path_between` early-returns `journeys::line(start,target)` when retained_movement_orders is supported. It bypasses normal intermediate passability/route search in that branch.
- `runtime/render/movement.rs::draw_planned_path` uses green if a general cost is known, not if that unit may legally traverse every proposed step. Destination intent/general tile cost is not an authoritative route or permission proof.
- Thus a land endpoint may acquire a green straight preview crossing known forbidden water/other blockers, even if the server would reject, stop, defer or choose another actual route. No server illegal-move execution was observed/proven in this report.

- [ ] Client reproduces using exact received domain/capabilities, target/city owner and intermediate planning accessibility/domains/cost; isolate actual water versus river art.
- [ ] Validate known blockers/authorized exceptions and distinguish unverified destination feedback from a legally confirmed route. Do not show fabricated green route or turn count from geometry/cost alone.
- [ ] Determine whether a current revision-bound viewer-safe unit-specific route/quote is already available. If missing, agree on a safe route-display contract or explicit unavailable preview; no hidden map queries, copied movement engine or local route execution.
- [ ] Test land/ocean/foreign-city denial, legal detours/transport exceptions, fog/remembered/unknown cells, roads/river/bridge effects, depleted movement and stale receipt identity, without contacting a live service.

This section primarily identifies a client preview defect; it is not a request to weaken server movement rules or proof that the authoritative engine permits ocean-walking settlers.

## 5. Oil-looking good on plains-looking terrain / catalog lineage

Newest evidence: client `images/current-impossible-good-oil-on-plains-oct-22-2026.png`. User reports an impossible oil spawn on plains. The raw tile/good projection was not captured, so renderer and authority contributions must be distinguished.

Concrete original/cross-installation facts (read-only):
- Client bundled original `ctp2_data/default/gamedata/goods.txt`/`goodsID.txt` map `TERRAIN_DESERT_GOOD_TWO` to oil/GG016; original plains ONE/TWO are spices/elephant. Original `terrain.txt` Resources lists those corresponding terrain goods.
- The configured GOG tree has extra THIRD/FOURTH resources, different yields and even an extra tundra oil mapping. It is not a safe substitute for original gameplay behavior merely because the installed sprite exists. The original local released source remains behavior authority.
- Server `internal/world/metadata.go::generateRiverGoods` currently builds `TERRAIN_<terrain>_GOOD_<ONE|TWO|THREE|FOUR>` with four equally selectable suffixes, without consulting the selected original terrain resource catalog. Some original terrain variants are missing or placeholder0. This is a concrete catalog/version compatibility risk, not proof of the exact screenshot's oil/terrain pair.
- Client `assets/goods.rs` maps received symbolic good IDs through selected installation goodsID.txt; map placement is not generated locally. Client terrain-art mapping/effective terrain and the GG sprite resolve also require inspection.

- [ ] Capture an injected/read-only reproduction with exact terrain, good_id, observed/stale turn, selected installation/catalog version, art base and sprite; determine whether the world really says plains+oil, or one visual is mis-mapped.
- [ ] Check original Resources eligibility and server generation/terraform/city-conversion/remembered-observation lineage, including whether a good survives an authoritative terrain change legitimately.
- [ ] Validate generated goods against the pinned canonical original catalog/version rather than assuming four resources exist for every terrain; do not use community/GOG extensions as original behavior authority.
- [ ] If client mapping is wrong, fix renderer/catalog lookup; if authoritative state violates intended terrain policy, server owner fixes state/generation contract. Client must not silently delete/move/replace goods to disguise an engine error.
- [ ] Regression every valid terrain-good pair, missing/undefined/unknown IDs, modified installation versus original rules, remembered fog and actual terrain changes; no hidden geography, copied assets, local resource generation or live-service tests.

## 6. Repeated local early-turn INFO resyncs / refresh-vs-recovery distinction

User-provided game55 logs (2026-10-22):19:07:02Z revision10/turn2/22tiles/2units/0cities;19:07:52Z rev14/turn2/22tiles;19:07:55Z rev18/turn3/28tiles;19:07:58Z rev21/turn3/28tiles;19:08:01Z rev23/turn4/32tiles;19:08:04Z rev28/turn5/36tiles. Counts remain2units/0cities. First gap50s then approximately3s. User expects healthy local early turns not to repeatedly fault/recover; no cause is claimed proven by log wording alone.

Client `server.rs::fetch_resync_projection` emits this INFO string for EVERY finite `X-Gameplay-Resync:current` snapshot, even ordinary SSE-triggered refresh. Thus these lines alone do not prove actual Resynchronizing status/rollback or a server rejection. Current client watcher also requests a mandatory current read after disconnect and shares3100ms event admission. The historical-notice state-ownership bug was removed in3df0793, but running build/restart and actual wire events must be checked before declaring that deployed/resolved this symptom.

- [ ] Distinguish snapshot replication from real error recovery via bounded cause-tagged metrics/UI transitions; don't simply suppress or relabel a genuine loop.
- [ ] Verify actual SSE revision/durable ID envelopes, normal close/idle/timeout/status/fragment behavior and protocol expectations. Accurate isolated stream tests, no live service in tests.
- [ ] Check own receipt-covered invalidations, stale/missing hints, mandatory request retention, rate-limit scheduling and whether current-read waits block new user commands in the shared worker.
- [ ] Separate measured client scheduling from actual server planning/AI-resolution duration; request server changes only for demonstrated protocol/observability/engine gaps. Local still uses the authoritative server, not a bypassed client engine.
- [ ] Live acceptance: early local turns stay healthy without unexplained recovery churn, responsive measured commands, genuine disconnects still recover. Normal full-state replication is allowed but must not masquerade as continual fault resync.

Detailed client checklist: `plans/local-game-resync-investigation.md`. This adds tracking/research only; server implementation remains untouched and UI screens remain current priority.

## Delivery boundary

This file is the only requested server-repo write. No server code/configuration/data changes or server implementation commit were made. Client single entry point: `PLAN.md` (execution order, all33 acceptances, complete active-detail index and retired implementation ledger). Detailed requests: `plans/oct-22-2026-ui-followups.md`, `plans/original-ruin-exploration-text.md`; the former separate UI-priority tracker has been consolidated into PLAN.md. All33 original complete-frame acceptances remain separate and unaccepted until honestly verified.
