# Retail starting-advance initialization

Scope: the latest startup omission added to `known-server-issue.md`. Server only;
new games only, no live requests, database repairs, client edits or save migration.

- [x] Read the new report and inspect the actual startup/import/research paths.
- [x] Verify original released Player::InitPlayer, Advances::InitialAdvance and DiffDB::ParseAdvanceChances plus original DiffDB data (all six profiles).
- [x] RED → GREEN → refactor: import source-ordered human/AI chance pairs and caps; validate and pin an explicit versioned starting policy/difficulty/age.
- [x] RED → GREEN → refactor: deterministic per-player startup grants, guaranteed-over-cap semantics, later-age grants, legal initial research, one-time initialization and rollback safety.
- [x] RED → GREEN → refactor: wire real new-game startup; all six profiles, mixed human/AI, freshly founded population-one Warrior options/acceptance versus Settler denial.
- [x] Cover later-age startup, canonical restart/idempotency, malformed policy/config and no implicit legacy-save initialization.
- [x] Verify transactional public startup/persistence and unchanged baseline/synthetic behavior (targeted suites; full verification below).
- [ ] Review against original sources and privacy/determinism boundaries; document exact supported setup/default policy and differences (no SLIC or unrelated difficulty bonuses).
- [ ] Full suite, race, vet/build, freeze, formatting/diff verification; atomic commits through unchanged hooks.

## Source decisions

Original `Player.cpp:505–575` traverses difficulty chance rows in file order. A
chance >=100 always grants and increments the count, even when the cap is already
reached; a chance below 100 draws only while granted < controller-specific cap.
The grant does not require researching its prerequisites first. Later-age starts
instead grant all strictly earlier-age advances, then compute legal research.
`DiffDB.cpp:325–362` reads human then AI chance and appends each row (order matters).
`Advances.cpp:436` uses InitialAdvance/SetHasAdvance rather than the earned-research
reward path. Original data guarantees Toolmaking at every difficulty; Impossible
also guarantees Agriculture. Userprofile difficulty is a local preference, not a
server-authoritative setting to read implicitly.

## Implementation evidence

- RED actual-startup test: all six profiles produced an empty discovery list and no Toolmaking. GREEN: new-game startup initializes policy before visibility/persistence; all profiles now advertise/accept Warrior from a population-one first city while still rejecting Settler. The same test verifies legal initial research, age, deterministic reordered-roster replay and public selected-policy metadata.
- Import RED: no starting policy/parser existed. GREEN/refactor: strict six-block DiffDB parsing, controller-specific caps, ordered chance pairs, unknown-reference/runtime validation, deep-copy projection and hashed versioned policy. Both original client release and configured server installation have identical parsed starting grant profiles. No userprofile preference, client roll or hardcoded grant list is used.
- Engine RED: missing bootstrap. GREEN/refactor: source-order cumulative caps (guaranteed grants count AND bypass cap), exact `< chance` threshold, zero-chance draws while eligible, per-player injected rolls, later-age helper integration, legal research after grants and persisted one-time marker. Invalid RNG/policy/progress fails without mutating input; initial grants do not become earned-research feats/events.
- Public integration: authenticated startup persists human/AI discoveries and research at revision 1; first city at revision 2 has usable Warrior choices; accepted Add survives runtime restart and immutable receipt replay. Duplicate start cannot reroll. Malformed difficulty first reproduced as 500, now returns safe 400 with no persisted world/state, and a subsequent valid retry succeeds. All fixtures use injected seed/sources and an isolated in-memory SQL runtime, never a live service or user save.
