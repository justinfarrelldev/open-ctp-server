# AI production priority and original civilization city names

Scope: server only; original CTP2 names, pinned data, deterministic owner-local
selection; no live save/database repair or automatic renaming of existing cities.

- [x] Inspect original CTP2 AI production and civilization naming/data behavior.
- [x] RED: planner/one-action driver production must beat closeout and routine exploration; empty/illegal/foreign/stale/already-queued cities remain safe.
- [x] GREEN/refactor: fix ordinary production priority without unbounding the runtime action budget or activating pointless low-priority actions.
- [x] RED: import original civilization records, ordered CityName references and localized text; reject malformed/missing names and preserve exact names/encoding.
- [x] GREEN/refactor: versioned canonical pinned city-name catalogs, explicit selected civilization mapping and runtime validation; no mutable asset reads when loading saves.
- [ ] RED → GREEN/refactor: default founding, AI, ruins and own-only projected suggestions use the selected civilization; exhaustion/collisions, custom names, fog privacy, reload determinism and legacy fallback.
- [ ] Integration: real runtime worker with ActionBudget=1/closeout queues legal production, completes units, and names human/AI cities from their respective original lists; restart/replay preserved.
- [ ] Review against original source, production authority, legacy pins, privacy and deterministic boundaries; update contracts/report.
- [ ] Full suite, race tests, vet/build, format/diff/freeze checks; atomic commits with unchanged hooks; complete checklist.
