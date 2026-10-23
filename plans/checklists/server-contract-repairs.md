# Server contract repairs

Scope: `known-server-issue.md` and `server-client-schisms.md`; server only. No live services or user games. Preserve authoritative mechanics and fog privacy.

- [x] Read reports and previous plans; commit user-authorized baseline changes with unchanged hook.
- [ ] Source-check original production, river artwork, naming and terrain resources.
- [x] RED → GREEN → refactor: complete, shared authoritative append production options, revision/city/hash/phase bound, owned only; full edit remains authoritative.
- [ ] RED → GREEN → refactor: safe structured command rejections and correlated private diagnostics.
- [ ] RED → GREEN → refactor: explicit pinned rules provenance.
- [ ] RED → GREEN → refactor: observed river geometry with persisted fog memory and no hidden-neighbor records.
- [ ] RED → GREEN → refactor: nontechnical city default names for human/AI/ruin paths, without foreign-name oracles.
- [ ] RED → GREEN → refactor: original-catalog terrain goods generation; explicit version lineage.
- [ ] RED → GREEN → refactor: public turn context and bounded timing/protocol observability; no fabricated AI progress or route guarantee.
- [ ] Document new contracts and server/client delivery boundary in reports.
- [ ] Thorough source/security/maintainability review, full tests, race-focused tests, vet/build, freeze verification, diff checks; atomic commits.

## Evidence

Baseline commit `fbc9c0c` includes the supplied reports and retail Taskfile defaults. Client repository remains untouched.

Production RED: missing `production_options` on an empty city. GREEN: all eligible unit/building/conversion IDs available on empty/nonempty queues, sorted, canonical costs, complete/unavailable distinction, owner/phase gates. Refactored append validation into a source-free read-only helper and removed the divergent visibility predicates. Existing retained edits still revalidate in isolation. Engine/visibility tests pass, including 18 eligibility cases (imports, effective buildings, global wonder restrictions, government, advances, special units and duplicates), full capacity and invalid-retained-entry tests. Source: original `CityData.cpp::CanBuildUnit/CanBuildBuilding` and `bldque.cpp` existing atomic queue source notes. No prospective turn estimate is fabricated.
