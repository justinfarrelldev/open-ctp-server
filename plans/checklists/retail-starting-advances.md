# Retail starting-advance initialization

Scope: the latest startup omission added to `known-server-issue.md`. Server only;
new games only, no live requests, database repairs, client edits or save migration.

- [x] Read the new report and inspect the actual startup/import/research paths.
- [x] Verify original released Player::InitPlayer, Advances::InitialAdvance and DiffDB::ParseAdvanceChances plus original DiffDB data (all six profiles).
- [x] RED → GREEN → refactor: import source-ordered human/AI chance pairs and caps; validate and pin an explicit versioned starting policy/difficulty/age.
- [x] RED → GREEN → refactor: deterministic per-player startup grants, guaranteed-over-cap semantics, later-age grants, legal initial research, one-time initialization and rollback safety.
- [x] RED → GREEN → refactor: wire real new-game startup; all six profiles, mixed human/AI, freshly founded population-one Warrior options/acceptance versus Settler denial.
- [x] Cover later-age startup, canonical restart/idempotency, malformed policy/config and no implicit legacy-save initialization.
- [x] Verify transactional public startup/persistence and unchanged baseline/synthetic behavior (full suite and race verification passed).
- [x] Review against original sources and privacy/determinism boundaries; document exact supported setup/default policy and differences (no SLIC or unrelated difficulty bonuses).
- [x] Requested additional regression: actual economic growth to population two exposes and accepts Settler across all six profiles, including a nonempty queue, without assigning population/food or relaxing validation.
- [x] Full suite, race, vet/build, freeze, formatting/diff verification; atomic commits through unchanged hooks.

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
- Requested population-two follow-up: the all-six-profile actual-startup test now advances real human/AI turns using the pinned city economy until the city grows from one to two, retaining a Warrior queue. No population/food-store override is used. Settler is then present in complete options and accepted by the full atomic editor. All six transitions passed (`TestRetailActualStartupEnablesFirstWarriorAtEveryDifficulty`).

## Review boundaries

Initialization is called only from new-game construction; no decode/read/normal
command path grants discoveries. A nil old policy remains nil, and the stored
marker prevents repeat initialization after restart. Validation and copy-on-success
keep malformed policy/RNG/progress from mutating canonical input. Dedicated streams
are seeded from the server-generated world and isolated per player; no roll/seed or
foreign discovery is projected. Policy data, difficulty and start age are hashed;
only the public selected settings are exposed. Original guaranteed/cap/source-order
semantics are preserved; research uses the engine's deterministic ID-order tie-break
rather than promising byte-for-byte original database/RNG replay. Other difficulty
bonuses, SLIC, later-start calendar epoch and existing saves remain out of scope.

The source comparison also checked `Player.cpp:334–349`'s treatAsRobot distinction:
server persisted human/AI controller kind is the authority, not display name,
account identity or game mode. The original and configured DiffDB starting profiles
are byte-for-fact equivalent under the typed parser; no asset/userprofile is written.
`Player::SetHasAdvance` (9929 onward) also guards feat emission during network launch
or before network readiness; this server bootstrap does not fabricate earned feats
or execute original script notifications.

## Final verification

- Actual all-six population-two transition passed against both configured GOG data
  and the original client release. The checked-in game integration prefers the
  original release via `CTPCLIENT`, falling back to the local installed catalog;
  the separate public HTTP/restart test exercises the configured released catalog.
- Fresh full suite: `GOMAXPROCS=4 go test -timeout=16m -p=2 ./... -count=1 -coverprofile=/tmp/retail-startup-coverage.out` — passed, including fullgame/onboarding. Log: `/tmp/retail-startup-full.log` (`EXIT=0`).
- Fresh race: `GOMAXPROCS=4 go test -race -timeout=12m -p=2 ./internal/game ./internal/gameplay/{rules,engine,visibility,service,http,events} ./internal/e2e/onboarding -count=1` — passed. Log: `/tmp/retail-startup-race.log` (`EXIT=0`).
- `go vet ./...`, `go build ./...`, 81-file production freeze, changed-Go-file gofmt
  check and `git diff --check` — passed. New bootstrap production files are locked;
  no temporary lock-refresh command remains.
- Total coverage 81.1%; engine 86.5%. The pre-existing 90% engine coverage deficit
  remains documented and was not masked by lowering any threshold or test gate.
- Import/pin, engine bootstrap, actual startup/public persistence, original-asset
  population transition and reviewed freeze changes committed separately. Normal
  unchanged `hooks` Swagger pre-commit generation ran on every commit. No client
  source, installed assets, live service, user save or database repair was touched.
