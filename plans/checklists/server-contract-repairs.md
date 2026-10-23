# Server contract repairs

Scope: `known-server-issue.md` and `server-client-schisms.md`; server only. No live services or user games. Preserve authoritative mechanics and fog privacy.

- [x] Read reports and previous plans; commit user-authorized baseline changes with unchanged hook.
- [x] Source-check original production, river artwork, naming and terrain resources.
- [x] RED → GREEN → refactor: complete, shared authoritative append production options, revision/city/hash/phase bound, owned only; full edit remains authoritative.
- [x] RED → GREEN → refactor: safe structured command rejections and correlated private diagnostics.
- [x] RED → GREEN → refactor: explicit pinned rules provenance.
- [x] RED → GREEN → refactor: observed river geometry with persisted fog memory and no hidden-neighbor records.
- [x] RED → GREEN → refactor: nontechnical city default names for human/AI/ruin paths, without foreign-name oracles.
- [x] RED → GREEN → refactor: original-catalog terrain goods generation; explicit version lineage.
- [ ] RED → GREEN → refactor: public turn context and bounded timing/protocol observability; no fabricated AI progress or route guarantee.
- [ ] Document new contracts and server/client delivery boundary in reports.
- [ ] Thorough source/security/maintainability review, full tests, race-focused tests, vet/build, freeze verification, diff checks; atomic commits.

## Evidence

Baseline commit `fbc9c0c` includes the supplied reports and retail Taskfile defaults. Client repository remains untouched.

Production RED: missing `production_options` on an empty city. GREEN: all eligible unit/building/conversion IDs available on empty/nonempty queues, sorted, canonical costs, complete/unavailable distinction, owner/phase gates. Refactored append validation into a source-free read-only helper and removed the divergent visibility predicates. Existing retained edits still revalidate in isolation. Engine/visibility tests pass, including 18 eligibility cases (imports, effective buildings, global wonder restrictions, government, advances, special units and duplicates), full capacity and invalid-retained-entry tests. Source: original `CityData.cpp::CanBuildUnit/CanBuildBuilding` and `bldque.cpp` existing atomic queue source notes. No prospective turn estimate is fabricated.

Rejections RED: HTTP returned only the generic error field. GREEN/refactor: allowlisted stable codes/messages and random server correlation ID/header, with typed queue identity/shape/capacity/content/prerequisite errors and phase/revision/controller distinctions. Private logs retain server-authored production reason, game ID, expected revision and SHA-256 command fingerprint; payload/session/raw IDs and arbitrary decoder/database errors are never logged. Public JSON never includes internal error text. Kept private error chains/messages for existing pending-takeover behavior. Engine/service/HTTP suites pass, including log-redaction and unknown-code tests.

Provenance RED: projected rules family absent. GREEN/refactor: synthetic manifest provenance, new retail runtime `family` included in canonical pin, and `unknown` for unlabelled legacy runtime pins. Existing hashes remain the identity; old games are not migrated or heuristically labelled retail. Rules/visibility/game tests pass.

River RED: a visible frontier river lacked its legitimate NW/SE artwork when SE continuation was hidden. GREEN/refactor: versioned 4-bit SW/NW/NE/SE art facts plus beach-mouth discriminator and observed revision, indexed once in Inspector. Visible tiles only; persisted in last-observed inspection, never recalculated for fog or legacy unknowns. Original `tiledmap.cpp:1100,1420–1520` computes art from neighboring rivers; `UnseenCell.cpp:52` snapshots TileInfo. All 16 edge patterns, beach suppression, hidden-neighbor absence, fog/reload/hidden changes/reveal tests pass; engine/visibility/service suites pass. This deliberately reveals only the legitimate center tile artwork, not neighbor entities or terrain.

Naming RED: omitted default rejected, supplied names unvalidated, global founding-name check exposed foreign names. GREEN/refactor: shared bounded/control-safe name normalization for founding/rename; duplicate display names allowed (IDs identify cities). Human action suggests own-only `City N`, AI omits a custom name, and ruin cities use the same neutral default. Original `CivilisationData.cpp:165–221` chooses per-player names; localized civilization lists are not pinned today, so neutral fallback is explicit rather than fake Roman/civilization names. Existing saved names are not overwritten. Found/foreign-name/invalid rollback/ruin/visibility and engine/visibility/AI/service tests pass.

Resources RED: v4 had no resources/lineage and handoff flattened mountain/desert to plains under synthetic rules (a concrete oil-on-plains path independent of the unobserved screenshot). GREEN/refactor: new `ctp2-original-goods-v4` defaults to an immutable original semantic resource registry, excluding nonexistent FOUR/THREE and shelf resources; keeps V1 terrain/spawns, V2 rivers, V3 ruins unchanged. Public generation metadata pins `ctp2-original-resources-v1` without exposing seeds. New synthetic pins explicitly support all generated terrain families rather than relocating resources through flattening; startup rejects unsupported terrain and invalid original initial pairs. Older generator versions/saves remain unchanged. 100-seed placement, repeatability/other-lineage, wrong/unknown/extended-ID, startup handoff and persistence tests pass; world/game/rules/engine/visibility/AI/service suites pass. Original client `ctp2_data/default/gamedata/terrain.txt` Resources and goods.txt (two variants, tundra/swamp ONE, shelf none) were compared against server released installation (GOG extended four variants). `Cell.cpp:779` preserves goods on terrain change: this initial-placement policy does NOT delete legitimate transformed deposits or remembered facts.
