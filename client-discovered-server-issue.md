# Client-discovered server issues: authoritative map and gameplay contracts

## Current implementation status

Server contract v2 now implements the confirmed HTTP/progress/projection defects and supported map, city, queue, trade, quote/forecast, research-goal, movement-input and hybrid/resync contracts. See [the current wire contract and explicit feature boundaries](docs/client-gameplay-contract-v2.md) and [implementation/test evidence](plans/checklists/client-report-contracts.md).

Canonical world downloads and terrain-derived ratings/pollution aggregates are retired, including for owners. Planning validates inputs before persistence, handles invalidated orders transactionally with private rejection results, and keeps completed-round history correctly labeled.

**Not implemented:** retained multi-turn movement, mayor/governor automation, the exact slider optimizer, culture/style modeling and ruin rewards. Foreign statistics remain private. These are explicit false capability flags, not claims of completed retail features. Client UI/decoders/optimistic animation/full-model reconciliation still require client integration.

The source inspection sections below are **historical evidence at the cited revisions**, not current-code descriptions. Implementation was verified with engine/mock/in-memory tests, not a running external game.

## Original context and scope

Discovered during the CTP2 client's original-game terrain/UI fidelity work. The supplied reference screenshots show connected rivers, river mouths meeting coastal surf, and river tiles retained as darkened explored terrain after a settler moves away.

The original investigation was based on **read-only source inspection**, not a failing request against a running server. At that time no server implementation or tests were changed. Subsequent server implementation is tracked in the current status above.

At the inspected revisions the engine supported river-bearing terrain, but normal generated-world and viewer-safe projection contracts did not carry the information needed by an authoritative client. This is not solvable by changing the client's texture selection alone.

### Source revisions and reading guide

The expanded frontend-linked requirements below were checked against client revision [`9cf86953d4e288d93719f733b7050844f9a6a5d6`](https://github.com/justinfarrelldev/ctp-client/tree/9cf86953d4e288d93719f733b7050844f9a6a5d6) and server revision [`5ed6237f6cd79d446fc7b804b3c554d2e963a00f`](https://github.com/justinfarrelldev/open-ctp-server/tree/5ed6237f6cd79d446fc7b804b3c554d2e963a00f). Frontend links are commit-pinned GitHub permalinks with actual file/line anchors, not local filesystem paths or moving branch links.

The original river/resource/ownership findings are retained below. [Frontend-linked implementation requirements](#frontend-linked-implementation-requirements) adds the required changes, their client consumers, reasons and acceptance cases. Related findings from [the separate interaction report](client-discovered-server-issues.md) are consolidated here without changing that report; its older revision and statements about this file being unchanged describe the earlier investigation, not this update.

**Classification:** confirmed projection/gateway omissions and production-loss behavior are distinguished from new feature contracts and client defects. Proposed field/operation names are recommendations, not claims that an endpoint already exists. End Turn input dispatch, cursor alignment, asset decoding, missing client windows and rendering are not automatically server defects. The original inspection was not a live-game reproduction; current implementation scope and verification are tracked above.

## Confirmed data-loss points (original inspected revisions)

### 1. The canonical gameplay engine has an independent river property

`internal/gameplay/engine/engine.go`, `type Tile`, includes:

```go
Terrain string `json:"terrain"`
River   bool   `json:"river,omitempty"`
```

River is not encoded by the terrain ID. Plains with a river and plains without a river can have the same `Terrain` value. Engine consumers use the separate flag:

- `internal/gameplay/engine/units.go` applies `terrain.RiverCost` when `tile.River` is true.
- `internal/gameplay/engine/cities.go` adds river food/production/gold yields.
- `internal/gameplay/engine/combat.go` consults river defensive modifiers.
- `internal/gameplay/engine/terrain_river_task118_test.go` exercises retail river terrain effects.

Consequently, a client cannot reliably reconstruct the flag from movement costs, terrain names, nearby units, or an image of a river. Those are not an authoritative river topology contract.

### 2. Generated-world data cannot currently represent rivers

`internal/world/contracts.go`, `type WorldTile`, contains only `X`, `Y`, and `Terrain`.

`internal/game/start_gameplay.go`, `initialGameplayState`, constructs each `engine.Tile` from a world snapshot using ID, coordinates, terrain and `Open: true`, without setting `River`.

Thus, although the engine model supports rivers, **this inspection does not establish that ordinary newly generated games contain any rivers at all**. The normal world-snapshot-to-gameplay path currently has no river value to propagate. Adding a projection field alone would not make rivers appear in newly generated maps.

### 3. Visible tiles omit river information

`internal/gameplay/visibility/visibility.go`, `type TileProjection`, exposes ID, X, Y, Terrain, Pollution, Stale, Observed, Planning and Overlays, but no River field.

The visible-tile branch of `visibility.Project` constructs `TileProjection` from canonical tile terrain/pollution plus planning and overlay information, and drops `tile.River`.

`tileOverlayProjections` iterates `state.Overlays`; it does not synthesize river overlays from `engine.Tile.River`. Therefore the current `overlays` field is not an alternative source of river topology.

Even a canonical fixture or loaded game that already contains `River: true` cannot currently expose that fact to the rendering client through this tile projection.

### 4. Fog memory cannot preserve last-observed rivers

`internal/gameplay/engine/engine.go`, `type FogObservation`, remembers Terrain and city metadata, but no River field.

In `internal/gameplay/visibility/visibility.go`:

- `Update` constructs `engine.FogObservation` from a currently visible tile but remembers only its terrain ID and city data.
- The remembered-tile branch of `Project` uses `observation.Terrain`, `Stale: true`, and the observation turn, without river metadata.

Only adding `River` to visible `TileProjection` would leave a second bug: once a settler moves away, the client would lose the river for the stale tile, producing a discontinuity or disappearing river precisely where the reference game shows dimmed remembered terrain.

## Why a client workaround would be wrong

The client must render from the viewer's projection, not from canonical game state or its own world generator. It should not:

- Invent rivers using a client-side seed or heuristic.
- Assume every plains tile or coastal tile has a river.
- Fetch hidden canonical world data to choose river graphics.
- Read the current hidden `engine.Tile.River` value for an explored tile rather than its last observation.
- Treat an omitted legacy JSON field as evidence that the entire map is authoritatively river-free.

The original CTP2 river artwork uses edge-connectivity masks and distinct river-mouth pieces. Drawing the correct piece requires a tile's river flag and its legally known adjacent connections. No client renderer can recover information the server never sends.

## Recommended fix

### A. Define a versioned generated-world river contract

1. Add river metadata to the generated `WorldTile` contract, or another canonical, documented representation with equivalent information.
2. Generate river topology on the server deterministically, with the server owning generation parameters, seed and generator version.
3. Persist river metadata in the world snapshot.
4. Copy that metadata into `engine.Tile.River` in `initialGameplayState`.
5. Keep river identity distinct from base terrain IDs; a river can traverse multiple terrain families.
6. Update generation/snapshot compatibility and replay/version contracts as appropriate. Do not silently change the meaning of an already persisted generator version or mutate existing authoritative games just to improve graphics.

If river generation is deliberately deferred, make that explicit: the projection fix can still land independently so existing/loaded/test canonical river tiles are represented correctly.

### B. Expose the river flag for visible tiles

Add `River bool` to `visibility.TileProjection`, using the JSON name `river`, and populate it from the canonical tile **only when that tile is visible**.

An explicit `"river": false` is preferable for a newly documented wire contract because a client can distinguish an explicit negative from an old server omitting the property. If `omitempty` is retained, document its semantics and consider a protocol capability/version field. A boolean is enough for the current engine representation; use edge metadata only if the canonical engine later models directed/partial rivers.

### C. Persist and project last-known river information under fog

1. Add river metadata to `engine.FogObservation`.
2. Capture `tile.River` alongside Terrain in `visibility.Update`.
3. For stale tiles, populate the projected field from **the viewer's stored observation**, not the current canonical tile.
4. Keep the existing visibility boundary: unexplored tiles and enemy-only observations must not be newly exposed.
5. Decide compatibility behavior for old persisted fog observations without the field. Do not backfill them from current hidden world state; that can reveal changes the viewer never observed. A tri-state/known marker, versioned observation, or conservative legacy handling can preserve that distinction.
6. Verify canonical JSON serialization, aggregate persistence, command receipts, replay and idempotent readback still preserve the new observation metadata.

### D. Keep art selection in the client

The server should send authoritative terrain/river facts, not copyrighted pixel data or SDL texture IDs. The client can use the player's installed CTP2 tileset to select connected river pieces and river mouths. Neighbor information beyond the projection must remain unknown; graphics should not require an omniscient world read.

## Suggested regression tests

Use in-memory engine states and existing mocked service stores. No live server/database or asset copying is necessary.

1. **Visible river/non-river pair:** same Terrain ID, different River values; both remain distinguishable after `Update`, `Project`, JSON encode and decode.
2. **Movement into fog:** reveal a river, move all viewer vision sources away, refresh/project, and assert the stale tile retains the observed river value and original observation turn.
3. **No hidden-state leak:** after observation, change the canonical hidden river flag; the viewer must still receive the old remembered flag until re-observation.
4. **Re-observation:** regain vision and assert both the current river value and observation timestamp update.
5. **Player isolation:** a second player with no sight or memory must not learn river tiles or topology from the first player's observations.
6. **Never-explored neighbor:** an unseen adjacent river tile must not be appended solely to make the client's river graphics connect.
7. **Durable command response:** extend the mocked aggregate/receipt test pattern in `internal/gameplay/service/fog_task17_test.go`; assert the accepted move persists river memory and the returned viewer projection contains only legal river metadata.
8. **Persistence/replay/idempotency:** serialize/restore the aggregate and compare projections; replaying the same receipt must not change river observations or leak fresher hidden data.
9. **World start handoff:** a generated snapshot containing a river should produce a canonical tile with `River: true`, with identical topology for identical version/seed/config.
10. **Legacy observations/projections:** cover omitted river fields explicitly and verify the chosen compatibility behavior without consulting unseen current tiles.
11. **Coastal river mouth:** a legal visible shore/river neighborhood should deliver sufficient facts for selecting a mouth piece, without requiring hidden neighbors.

## Client integration status

The client accepts optional `river` metadata and uses original tileset river connectivity/overlay data, with read-only retail-art tests for all land connections and coastal mouths. Older servers remain compatible; an omitted field does not cause guessed rivers. Full river parity in live authoritative games remains blocked until the generation/projection/fog-memory chain above supplies the facts.

The broader client work also preserves distinct retail terrain artwork (beach, deep water, snow mountains, jungle, etc.) rather than flattening everything to a few generic tile images. That client-side correction does not replace the missing independent river flag.

---

# Additional confirmed tile-metadata omissions

## Resource goods cannot be located from the viewer projection

The later water/fog reference screenshot includes an elephant resource deposit on remembered terrain. Its location is an independent server fact, not part of the terrain texture.

**Evidence:**

- `engine.Tile` in `internal/gameplay/engine/engine.go` has `GoodID string` with JSON name `good_id`.
- `internal/gameplay/engine/cities.go` uses `tile.GoodID` in city-available-goods calculations. A city's aggregate list of available goods does not identify the exact tile containing each deposit and is not a substitute for a map projection.
- `visibility.TileProjection` has no `GoodID`, and neither the visible nor remembered tile construction in `visibility.Project` copies it.
- `engine.FogObservation` and `visibility.Update` do not remember a last-observed good ID.
- `tileOverlayProjections` copies independent `state.Overlays`; it does not automatically convert `engine.Tile.GoodID` into a resource overlay.
- `world.WorldTile` and `initialGameplayState` also currently lack the generated resource-good handoff, just as they lack rivers. Engine support for a good is not proof that normal generation already places it.

**Effect:** an authoritative client cannot faithfully place an elephant, spices, or another resource deposit on the map, or retain the observed deposit under fog. Guessing goods from the terrain family or reproducing the original game's probability table on the client would invent game state. Inferring deposit coordinates from a city report or trade route would be unreliable and could reveal information not observed by that viewer.

**Recommended fix:** add a viewer-safe `good_id` tile property, populate it from currently visible canonical tiles, persist it in the viewer's fog observation, and use that observation for stale tiles. Handle explicit no-deposit versus legacy unknown metadata deliberately. Preserve the canonical good identifier/content namespace; do not send sprite pixels or infer the identifier from a texture index. If generation is intended to include goods, generate/persist them server-side and carry them into `engine.Tile.GoodID` as a separately versioned handoff change.

The original client installation can map a known canonical resource identifier to artwork. For example, the retail `goodsID.txt`/`goods.txt` mapping identifies `TERRAIN_PLAINS_GOOD_TWO` as sprite 8 (elephant). This is an artwork lookup only; it must not determine whether that deposit exists. The client now accepts optional `good_id` metadata, resolves the installed resource catalog, decodes original good SPR groups/hotspots, and renders visible or dimmed remembered resource actors. Tests retain all framebuffer data in memory and never invent deposits when the field is absent. This client support is not a server implementation change and cannot make missing authoritative deposits appear.

**Tests:** use equal-terrain tiles with and without different good IDs; verify JSON round trips, visible-to-stale retention, re-observation, persisted command receipts and no cross-player/never-explored leak. Change a canonical good while its tile is hidden and ensure the viewer still sees its last observation until sight returns. Test that aggregate city/trade goods do not cause previously unseen deposit tiles to be exposed.

## Independent tile ownership is also omitted

**Evidence:** `engine.Tile.OwnerID` exists, but `visibility.TileProjection` has no tile-owner field, the tile construction branches do not project one, and fog observations do not remember one. `CityProjection.OwnerID` and `FogObservation.CityOwnerID` describe city ownership, not ownership of every surrounding tile.

**Effect:** original-style political borders and ownership coloring cannot be exactly reconstructed for known non-city cells from these tile projections alone. The client should not rerun border/claim rules or assume that ownership of a city specifies all current surrounding claims.

**Recommended fix:** define the visibility policy for territorial ownership, expose the permitted `owner_id` for visible tiles, and persist/project last-known ownership for stale tiles if borders are intended to be remembered. Never expose per-tile ownership or borders on never-discovered cells, including through a separate supposedly public contract. Public aggregate diplomacy metadata must not identify hidden territorial cells. Do not mix current hidden ownership with last-observed terrain; remembered ownership changes only on permitted re-observation. Treat known neutral/unowned tiles and old observations with no ownership information distinctly when compatibility requires it.

**Tests:** use two visible non-city tiles with different owners, a neutral tile, a hidden ownership change, re-observation and different viewers. Assert that no territory/owner is invented by city proximity, and that the chosen public-versus-remembered policy is explicit and consistent after serialization/replay.

## Frontend-linked implementation requirements

### MAP-01 — complete the generation → visible projection → remembered-observation chain

**Frontend evidence:** [tile wire fields and lossless extra facts](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/server.rs#L124-L157), [projection-to-map conversion](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/model/server_projection.rs#L200-L229), and [resource actors drawn only from known tile goods](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/actors.rs#L14-L47).

**Required server changes and why:** implement the river, `good_id` and territorial `owner_id` contracts described above throughout generation, canonical state, visibility and persistence. The [generated tile schema](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/world/contracts.go#L52-L56) and [game-start handoff](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/game/start_gameplay.go#L123-L144) cannot currently carry river/good placement. The [projected tile schema](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/visibility/visibility.go#L28-L38) and [fog observation schema](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/engine.go#L462-L470) omit these independent facts. Artwork cannot restore data lost at those boundaries. Ownership must follow actual territorial rules, not be invented by the generator or inferred from the nearest city.

**Compatibility:** distinguish known false/no-good/neutral from an old server or observation with unknown metadata. The frontend currently defaults an omitted river to `false` and an omitted good to an empty string for compatibility; it does **not** yet preserve that distinction. A new capability/version or nullable/known-value contract requires matching client decoder changes. Do not describe the current defaults as sufficient tri-state support.

**Acceptance:** all original river tests above also apply to goods/ownership as appropriate: deterministic versioned generation, world-start handoff, observation, loss of vision, hidden canonical change, re-observation, player isolation and persistence/receipt/replay consistency. Never expose unseen neighbors just to improve river connectivity.

### MAP-02 — project effective, observation-safe tile information, not just base rules

**Frontend evidence:** [Map Info obtains values from base terrain definitions](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/context.rs#L68-L145), [goods read from the cached tile](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/context.rs#L146-L161), and [retained rules, reports and planning state](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/server/interface.rs#L120-L147).

**Confirmed gap:** `rules.movement.terrains` contains reusable rule definitions, not the effective facts for each tile. River bonuses, completed improvements, effective terrain and city-center rules can change yields. The server already has [terrain/city tile yield calculations](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/cities.go#L3417-L3482); the tile projection has no corresponding effective-yield contract. A city total is not a per-tile breakdown.

**Required server changes:** define the meaning of inspectable food/production/gold and movement cost, including whether the display is general tile yield, city-center yield or viewer-specific worked yield. Derive permitted values from the engine and pinned rules, attach them to visible tile snapshots and remember last-observed values under fog. Separate displayable terrain movement cost from unit/domain/access/ZOC planning information; current planning facts must not be made available for hidden terrain merely to fill a tooltip. Include observation turn/staleness and known/unavailable semantics. Do not guess resource yield contributions that the actual engine does not apply.

**Why:** opening Map Info should be a read of cached snapshot/discovery facts, not a request to a tile-lookup endpoint or a second economy implementation. Hidden improvements or policy-dependent changes must not leak through recomputation using fresh canonical state.

**Rules-mode prerequisite, not another confirmed defect:** [game start chooses synthetic rules unless retail mode is explicitly enabled](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/game/start_gameplay.go#L50-L94). Retail expectations require `OPEN_CTP_RULES_MODE=retail` and a valid `OPEN_CTP_RETAIL_GAMEDATA`; the actual pinned/projected rule content is authoritative. Verify the game's rules/content hash before diagnosing a numerical mismatch. Do not silently substitute the client's installed retail values for a synthetic or differently configured server game, or change existing games' rules to repair a display.

**Acceptance:** compare projected values with engine calculations for river/non-river, completed/in-progress improvements and city-center cases; test explicit zero versus unknown, hidden changes and re-observation. Receipts, initial/current resync and persisted observations must agree. Client follow-up must replace the base-definition lookup and its `unwrap_or(0)` fallback; that fallback is a client bug, not something the server should accommodate with invented zeroes.

### CITY-01 / CITY-02 — preserve real names and define permitted city-art facts

**Frontend evidence:** [city conversion reads `name` and queue head rather than inventing captions](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/model/server_projection.rs#L147-L173), [foreign remembered-city safety gate](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/cities.rs#L18-L46), [building sprite selection](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/cities.rs#L49-L96), and [name/population/owned production labels](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/cities.rs#L99-L178).

**Confirmed omission:** canonical [City has `Name` and `TypeID`](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/engine.go#L292-L318), but `CityProjection` and remembered city construction do not return them. Project names for owned/visible cities and retain the permitted last-observed name for foreign stale cities. Hidden rename, capture or destruction must not update stale facts until the viewer is allowed to know.

**Additional contract needed for exact artwork:** define authoritative city type, culture/style, age and capital status, including their relationship to owner versus founder and capture/era changes. Project semantic IDs or facts that the client can map through installed catalogs, not GC sprite filenames, SDL IDs or copied artwork. Do not assume canonical city type is automatically the same thing as culture/style; some of these facts need an engine/rules design, not merely a new JSON field. The current ancient-style fallback is presentation only and cannot certify the actual city's culture or age.

**Visibility:** owned queue/economy facts may be fresh; remembered foreign art/name facts must come from observations. The existing memory does not store population/style/age either. If those labels or sprite recipes are to persist under fog, explicitly remember the permitted values or leave them unknown; never send fresh private queue data for a remembered enemy city. Production-turn estimates are covered under ECON-01.

**Acceptance:** a non-ID city name survives founding receipts, reconnect and JSON round trips; foreign names/population/art facts survive only as last observed; capture, era/style/capital changes obey the declared policy; legacy observations remain unknown. Test two viewers with different observations of the same city.

### HTTP-01 — make advertised legal actions reachable without weakening validation

**Frontend evidence:** [explicit HTTP-supported route filter](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/server/interface.rs#L50-L106), [unit context actions require an exact unambiguous candidate](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/context.rs#L245-L291), and [manager dispatch wraps only supported advertised actions](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/bindings.rs#L241-L250).

**Required change and why:** reconcile `visibility` action definitions with the [HTTP command table and exact payload allowlists](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/http/http.go#L62-L111). At the inspected revision, a static comparison still finds 18 advertised-but-unregistered literal routes: `assassinate-ruler`, `bio-terror`, `bombard`, `convert-city`, `create-franchise`, `create-park`, `expel`, `injoin`, `investigate-city`, `nano-terror`, `plague`, `plant-nuke`, `reform`, `settler-slave-raid`, `slave-raid`, `sue`, `underground-railway`, `upgrade-unit`. When eligible, an advertised action must reach the normal authenticated command service rather than 404. Intentionally unsupported commands must not be advertised as executable.

Use a shared manifest or cross-layer tests to prevent drift; preserve ownership, target, prerequisites, unknown-field rejection, revision checks and idempotency. Add mocked HTTP dispatch tests for every route/field set, not just one expel example. The frontend must subsequently update its whitelist and target choosers; registering `reform` does not repair a client looking for the different string `reform-city`.

### QUEUE-01 / QUEUE-02 / QUEUE-03 — fix progress loss and provide atomic entry-addressed edits

**Frontend evidence:** [Remove is deliberately blocked while production is stored](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/bindings.rs#L116-L130), [Insert/Up/Down/Clear/custom queue controls explain the unsafe contract](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/bindings.rs#L135-L145), and [current queue rows use content IDs as selection IDs](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/data.rs#L102-L110).

**Required server changes:**

1. Fix [unconditional stored-production reset on cancellation](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/cities.go#L1008-L1021). Removing a future entry must preserve the unchanged head's progress. Source-check and document transfer/loss rules for actual head changes rather than declaring all cancellation loss a bug.
2. Add a revision-checked atomic insert/reorder/remove/clear/replace contract or equivalent operations with explicit progress semantics. Cancel/re-add is not equivalent: it exposes partial queues and separate revisions and can destroy production.
3. Distinguish stable queue-entry identity from content/type identity. The [current type-ID duplicate rejection](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/cities.go#L759-L771) prevents repeated unit orders. Permit them where source rules allow, retaining building/wonder restrictions. Keep authoritative item costs/effects/prerequisites server-validated; an arbitrary replacement list cannot redefine rules.

**Acceptance:** cancel a tail behind 40 stored production without loss; head changes apply the documented policy exactly once; insert/move/clear are atomic; invalid or stale edits change nothing; idempotent replay does not repeat an edit; identical unit types remain individually selectable and only the completed entry is removed. Persistence, receipts and resync retain stable entry IDs. Client follow-up must migrate queue row selection and payloads before enabling these controls.

### ORDERS-01 — define distinct city and unit operations, not lookalike substitutions

**Frontend evidence:** [Rename City, Disband City and Mayor are explicitly unavailable](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/context.rs#L186-L203), and [Sentinel does not secretly submit Fortify](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime.rs#L598-L601).

**Feature contracts needed:** source-verified rename-city, disband-city and mayor/governor enablement/priority operations, with legal-action advertisement and projected resulting state. Sentinel/Sleep needs a documented order/lifecycle or an explicit decision that it is session-local unit-cycling state; it is not equivalent to defensive fortification. Rename/disband/governor controls cannot use unit `disband`, `fortify` or local fake mutations. Operations already implemented by the engine but missing HTTP routes belong under HTTP-01, not under a claim that their rules do not exist.

**Acceptance:** exact city/unit identity, ownership, eligibility and applicable last-city/size restrictions; invalid names/targets leave state unchanged; revision/idempotency and reconnect behavior; governor decisions remain authoritative; foreign remembered facts obey CITY-01. Existing Fortify, unit Disband, policy and research routes remain unblocked client integration work.

### TRADE-01 — expose authorized route commodity and accounting facts

**Frontend evidence:** [trade lists currently retain endpoint identities without invented goods/income](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/data.rs#L189-L224).

**Required change and why:** extend `TradeRouteProjection`/`tradeRouteProjections` with permitted canonical `GoodID` and `Gold` (and define whether income is gross/net/per-turn). Endpoint IDs, city goods and national totals cannot identify the commodity or income of each route. Establish owner/endpoint/embassy/public visibility policies rather than exposing all routes globally.

**Acceptance:** known zero/no-good differs from unavailable; each endpoint/viewer receives only allowed facts; piracy and route removal update accounting consistently; command receipts, persistence and resync agree. Creating a route already has an HTTP operation; missing client selection/filter/advisor rendering is not another server defect.

### ECON-01 — define display quotes and forecasts using authoritative calculations

**Frontend evidence:** [City panel looks for a rush-buy cost and leaves missing values unavailable](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/panel.rs#L278-L312), [Science already consumes projected progress/cost but has no turns estimate](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/data.rs#L302-L324), and [Empire values intentionally do not invent forecasts](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/panel.rs#L315-L356).

**Contract needed:** publish viewer-authorized effective rates/remaining-turn estimates and rush-buy quotes, or document complete sufficient inputs and exact display formulas where a pure client calculation is genuinely equivalent. Define stalled/indefinite/unknown estimates separately from zero. Production estimates must account for the actual effective engine economy, not divide by a terrain base value; research progress alone does not establish the research rate.

**Important payload boundary:** the current client searches `legal_actions[*].payload.cost`, but the server's rush-buy payload accepts only `city_id` plus the command envelope. Do **not** insert an extra `cost` into that executable payload just to satisfy the client: it would conflict with strict HTTP fields. Add a separate display quote/preview field or owned-city accounting fact and update the client to read it. Compute quotes with the same [rush-buy calculation used by execution](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/cities.go#L1023-L1060), tied to the displayed revision; clients cannot select their own price.

**Acceptance:** quotes match execution at the same revision; zero/completed, insufficient-gold and non-rushable cases are explicit; stale quotes do not authorize a stale purchase; policy/government/specialist changes refresh permitted forecasts. Existing projected city yields/reports, reserves, policies and research facts should be consumed first, not re-reported as universally absent.

### POLICY-01 / RESEARCH-01 — support optimizer and persistent research goals deliberately

**Frontend evidence:** [the Optimize Sliders control is created in the source-backed layout](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/fidelity.rs#L89-L99), [policy commands already submit typed values without live trial mutations](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/policy.rs#L152-L181), [Set Goal is explicitly disabled for the missing persistent contract](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/bindings.rs#L232-L234), and [Change Research already handles currently available advances](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/bindings.rs#L155-L175).

**Feature contracts needed:**

- For Optimize Sliders, follow the community-source callback/evaluator evidence in [the interaction report](client-discovered-server-issues.md#optimize-sliders--clarified-source-behavior-not-a-claimed-engine-bug). Provide server-owned optimization or read-only recommendations using actual counterfactual city/specialist/happiness/economy calculations. Specify constraints, deterministic tie-breaking, specialist side effects and apply/revision semantics. Repeated `set-social-policy` mutations to search a live game are unacceptable; the existing mutation endpoint is not a promised optimizer.
- If Set Goal is to be persistent and authoritative, add owned-player goal state, set/change/clear semantics, prerequisite progression and projection/reconnect support. Selecting a future goal is not the same as immediately researching an available advance. Keep local article navigation separate from game-state changes.

**Acceptance:** optimizer evaluation does not mutate persisted state; empty/no-solution cases are explicit; deterministic recommendations and subsequent apply respect revisions. Goals survive receipts/reload, progress only through legal prerequisites, can be replaced/cancelled and reveal no other player's research. These are feature requirements, not claims that existing endpoints are broken.

### MOVE-01 — establish retained multi-turn movement semantics separately from path artwork

**Frontend evidence:** [path input finalizes a whole good-path vector](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/movement.rs#L133-L175), and [numbered turn boundaries are presentation derived from planning facts](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/movement.rs#L116-L147).

**Current contract:** [engine `moveUnit`](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/engine/units.go#L431-L496) validates the full submitted path against current movement/fuel/domain/access/occupancy/ZOC before changing unit positions. A path exceeding current movement is rejected; this operation does not persist its remainder as a future order. That is not evidence the atomic current-turn command is defective.

**Contract needed for retail retained orders:** define server-owned destination/path intent, its cancellation/replacement, permitted replanning and how later turns resume, or explicitly document another supported authoritative design. Preserve exact unit/army identity; hidden obstacles must be handled without omniscient preview disclosure. Do not pretend that a numbered line means an order has been accepted, or blindly replay stale client path segments each round. Current-turn prefix submission remains a client option but is not equivalent to retained orders across reconnects/multiple sessions.

**Acceptance:** insufficient movement leaves a current-turn command unchanged; retained orders resume only when authorized, stop/replan safely for new obstacles/ZOC and ownership/army changes, survive persistence and do not execute twice on idempotent replay. Resolved end-turn outcomes must use accepted viewer-safe transitions. This does not prohibit optimistic client movement: for ambiguous moves, including movement toward obscured terrain, the client should begin provisional animation before server confirmation without inventing hidden terrain. The server still validates the command; rejection or divergence requires rollback, a visible “resynchronizing” message and replacement of the client's full authoritative model from the current viewer-safe snapshot. Speculation is presentation-only and must never be persisted as accepted game state.

### PRESENTATION-01 — define world/calendar and permitted history/message contracts

**Frontend evidence:** [world dimensions are currently inferred from the maximum projected coordinates](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/model/server_projection.rs#L102-L125), [the panel currently displays a hardcoded initial date then `Turn N`](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/render/panel.rs#L48-L60), [message-panel source selection](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/panel.rs#L24-L44), and [Statistics rows are not yet populated by the client](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model/management/data.rs#L151-L159).

**Contract requirements, not a claim all unfinished UI is server-blocked:**

- Publish public map dimensions and supported axis-wrap/topology semantics independently of visible cells, consistent with actual generation/movement rules. Fog-filtered maxima are not map bounds. Do not expose hidden terrain just to size the camera, or advertise wrap the engine does not implement.
- Publish an authoritative calendar value or versioned turn-to-calendar rules sufficient for the actual game configuration; client date formatting/localization remains local.
- For completed-build/ruin notifications and Statistics history/ranking, define which semantic results, stable IDs, event turns, permitted locations and historical samples each viewer may receive. The [current event projector emits type/sequence/revision/ordinal only](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/visibility/visibility.go#L2429-L2436); that is an invalidation cue, not complete text/target/history for original popups. Audit already available snapshot facts before adding data. Do not derive authoritative ruin rewards, past scores or hidden foreign history from screenshot fixtures. Decide separately whether read/unread state is local or synchronized.

**Acceptance:** bounds are stable as exploration changes; camera/wrap agree with legal engine movement; calendar changes match the pinned game rules; notification/result IDs allow deduplication/resume; historical samples obey viewer permissions. Reconnect does not duplicate rewards or expose never-observed locations. Rendering, popup interaction, eye navigation and graphs still require client implementation.

### TURN-01 / SYNC-01 — preserve the existing phase/resync contract; repair client consumption

**Frontend evidence:** [Return maps to End Turn but keypad Enter is not mapped here](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/ui_model.rs#L490-L499), [End Turn currently hardcodes `round-complete`](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime.rs#L561-L570), [receipt application rebuilds state](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime.rs#L663-L714), [one-shot resync request](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/server.rs#L421-L445), and [runtime loop currently processes SDL events without a gameplay event-stream consumer](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/runtime/session.rs#L163-L211).

**Do not misclassify this as a confirmed missing server endpoint:** the server already registers both `round-complete` and `pass`, projects `planning_status`, conditionally [advertises the applicable action](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/visibility/visibility.go#L548-L558), and provides [authenticated streaming plus `X-Gameplay-Resync: current`](https://github.com/justinfarrelldev/open-ctp-server/blob/5ed6237f6cd79d446fc7b804b3c554d2e963a00f/internal/gameplay/events/http.go#L32-L119). A queued/waiting projection does not justify unconditionally sending another round-complete or advancing the local turn. An already supported stream must be integrated before blaming the server for missing updates.

**Server acceptance work:** document/test open → queued/resolving → next-open/terminal transitions for local-AI and multiplayer play, explicitly including hybrid planning → conflict components → next planning round, inactive/component-external players, action availability while waiting, monotonic revisions, stale conflicts, lost-response retries, reconnect/event resume and idempotent End Turn. Hybrid is the primary multiplayer mode, not an optional sequential-only afterthought. Preserve viewer-safe receipts/current snapshots; all new metadata above must appear consistently through both paths. If a phase remains stuck in an in-memory engine/service test, report that distinct failure; no such live/server reproduction is claimed here.

**Client work remains:** resolve the exact current authorized action instead of guessing route/phase; support Return/keypad Enter and the original semantic button through one dispatch path; respect modal/focus handling and pending/error states; consume invalidations and refresh the viewer snapshot as needed. On disagreement, show “resynchronizing”, discard speculative state and replace the full authoritative client model with the current viewer-safe snapshot; incremental patching alone is not reconciliation. Resume input only after reconciliation. Optimistic animation may precede confirmation but must be reverted on rejection. [The current decoder already validates the resync envelope revision](https://github.com/justinfarrelldev/ctp-client/blob/9cf86953d4e288d93719f733b7050844f9a6a5d6/src/gameplay/server/projection.rs#L52-L71); do not replace it with an empty-world fallback or read canonical events.

## Server implementation acceptance checklist

This original handoff checklist remains a record of requested scope, not a claim that all retail features are complete. The live implementation checklist and wire contract linked above distinguish delivered server work, explicit unsupported feature decisions and remaining client work. Server implementation should follow its own `AGENTS.md`, existing plans and up-to-date `plans/checklists` entries, using red → green → refactor and original-source verification.

- [ ] MAP-01: versioned generated river/good handoff, permitted territorial ownership and viewer-specific remembered facts, including legacy-unknown behavior.
- [ ] MAP-02: effective inspectable tile yields/movement/goods cached in viewer projections/observations, with no hidden-state recomputation.
- [ ] CITY-01/02: real names/type and explicitly supported culture/age/capital/population memory semantics.
- [ ] HTTP-01: route/field manifest consistency and authenticated dispatch tests for advertised actions.
- [ ] QUEUE-01/02/03: preserve unrelated head progress, support atomic stable-entry edits and source-allowed repeated unit orders.
- [ ] ORDERS-01: explicitly supported city/governor/Sentinel contracts without substituting different operations.
- [ ] TRADE-01 / ECON-01: permitted route goods/income and separate authoritative display quotes/forecasts.
- [ ] POLICY-01 / RESEARCH-01: verified optimizer/evaluation and persistent goal contracts if those features are supported.
- [ ] MOVE-01: documented and tested retained-order semantics, distinct from current-turn atomic path execution.
- [ ] PRESENTATION-01: public world/calendar metadata and permitted notification/history contracts needed for the reference workflows.
- [ ] TURN-01 / SYNC-01: phase/action/receipt/event-resume regressions for the existing APIs; isolate any demonstrated server failure from the known client consumption gaps.
- [ ] Cross-cutting: every added fact survives canonical serialization, schema compatibility, mocked persistence, replay and idempotent readback without bypassing visibility. Use in-memory fixtures/injected stores/isolated HTTP stubs, not real server calls or file-mutating tests. Run normal server checks and unmodified commit hooks before implementation commits.

**Documentation verification:** frontend/server permalinks and line ranges checked against the pinned repository objects; the advertised-route comparison was repeated against the current inspected server revision. These checks establish source evidence only, not successful gameplay, full screenshot parity or completed server changes.

## Client defects are not server findings

During this work the client also needed fixes for stacked-unit owner mapping, exact unit identity in command payloads, retaining selection across projection updates, hidden panel hit boxes, empty-projection fallback inventing local armies/geography, unknown terrain being exposed by the local fog preference, and original asset decoding/placement. Those are client defects and are being corrected in the client repository; they are **not** reported here as server inaccuracies. Likewise, the `terrain-third.png` scene is an explicitly hand-built visual fixture, not evidence about the server's settler sightline calculations.
