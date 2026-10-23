# AI production priority and original civilization city names

Scope: server only; original CTP2 names, pinned data, deterministic owner-local
selection; no live save/database repair or automatic renaming of existing cities.

- [x] Inspect original CTP2 AI production and civilization naming/data behavior.
- [x] RED: planner/one-action driver production must beat closeout and routine exploration; empty/illegal/foreign/stale/already-queued cities remain safe.
- [x] GREEN/refactor: fix ordinary production priority without unbounding the runtime action budget or activating pointless low-priority actions.
- [x] RED: import original civilization records, ordered CityName references and localized text; reject malformed/missing names and preserve exact names/encoding.
- [x] GREEN/refactor: versioned canonical pinned city-name catalogs, explicit selected civilization mapping and runtime validation; no mutable asset reads when loading saves.
- [x] RED → GREEN/refactor: default founding, AI, ruins and own-only projected suggestions use the selected civilization; exhaustion/collisions, custom names, fog privacy, reload determinism and legacy fallback.
- [x] Integration: real runtime worker with ActionBudget=1/closeout queues legal production, completes units, and names human/AI cities from their respective original lists; restart/replay preserved.
- [x] Review against original source, production authority, legacy pins, privacy and deterministic boundaries; update contracts/report.
- [x] Full suite, race tests, vet/build, format/diff/freeze checks; atomic commits with unchanged hooks; complete checklist.

## Evidence and review

- RED reproduced the exact 90-versus-100 failure for both pass and open-planning
  closeout. The configured one-action/closeout driver now submits queue production
  then closeout. Routine exploration and formation cannot starve an idle city;
  urgent policy/combat priorities remain unchanged. The fallback was not lowered,
  which would activate repeated low-value fortification.
- Complete-empty/unavailable/foreign/stale/nonempty queues are excluded. A stronger
  illegal unit does not displace a legal weaker defender; the exact advertised
  item metadata is submitted. No authority or action budget was weakened.
- Original Governor::FillEmptyBuildQueues (governor.cpp:2371+) actually queues next
  units/buildings/wonders. Original CivilisationData.cpp:168–218 selects city names
  in source order with per-player usage and numeric suffixes. Server selection
  fixes counter/custom-name collisions with owner-only current-name checks.
- Import verifies original legacy CIV_* fields plus CITY_IS_CAPITAL and modern
  fields, actual English strings and Windows-1252 decoding. Shipped trailing spaces
  are normalized. Original source IDs/spellings are retained (AMERICANS and
  Washington D.C., not invented AMERICAN/Washington identifiers). Lists and aliases
  are deep-copied/pinned; all reads/reloads use canonical definitions.
- Both original release and selected extended catalog are covered. The old public
  `america` slug explicitly maps to AMERICANS even when the latter's localized
  CountryName is United States; unknown countries fail safely rather than borrowing
  names. The original release has no AZTEC civilization record, while the extended
  catalog does; no original record/list is fabricated.
- Founding, own-only suggestions and ruins use the same chooser; tests cover Roman
  versus Greek names, foreign-name independence, custom names, exhaustion and
  canonical restart. Existing cities and old pins are not rewritten.
- Authenticated public runtime test proves Rome/Athens, actual city-produced AI units
  (production ID tied to its city, not a ruin reward), same one-action worker,
  runtime reconstruction and immutable founding receipt replay. No live endpoint,
  client source, installed text file or user save was changed.
- New territory makes foreign settlement sites genuinely unavailable: a RED planner
  regression now verifies the AI skips them rather than looping on rejected founding.
- Full-suite follow-up found an E2E puppet's missing public rush-buy route once
  production worked. RED route regression → normal route mapping; mixed-controller
  campaign passed without changing strategy/goals/budget. Row 7's canonical state /
  event pins changed with persisted identities/center ownership: two independent
  observed runs produced the same winner and hashes before repinning. No forced
  winner, fabricated hash or increased campaign horizon was used.
- Enlarged legitimate viewer frames exposed two fullgame HTTP scanner defaults
  (64KiB), unlike onboarding's existing 16MiB allowance. RED large-SSE regression
  reproduced the transport error; both harness scanners now use the same bounded
  16MiB limit. No server event payload or game state was trimmed to hide the issue.
  Row 1 subsequently completed its unchanged conquest objective with identical
  observed artifacts across two clean runs and a durable-round restart; those
  measured state/event pins were updated together, without changing budgets.

## Final verification

- Fresh complete default suite with coverage: `GOMAXPROCS=4 go test -timeout=20m -p=2 ./... -count=1 -coverprofile=/tmp/city-development-coverage.out` — passed. `/tmp/city-development-full.log`, `EXIT=0`.
- Fresh core/game/public-runtime race suite: `GOMAXPROCS=4 go test -race -timeout=15m -p=2 ./internal/game ./internal/gameplay/{ai,rules,engine,visibility,service,http,events} ./internal/e2e/onboarding -count=1` — passed. `/tmp/city-development-race.log`, `EXIT=0`.
- Additional fullgame race batch exercised the large-frame test, mixed-controller
  campaign and pinned conquest. Its triple-run test exceeded its unchanged 2-minute
  per-scenario budget under concurrent full-suite load (no data race); isolated
  `GOMAXPROCS=4 go test -race -timeout=10m ./internal/e2e/fullgame -run '^TestTaskFiveExecutionSemanticsRemainGreen$' -count=1 -v` then passed all two-clean/one-restart artifacts, unchanged winner/hashes/budgets. `/tmp/city-development-isolated-race.log`, `EXIT=0`.
- Vet/build, 87-file production freeze, changed-Go gofmt and diff checks pass. Normal
  unchanged pre-commit Swagger generation ran on every atomic implementation/test
  commit. Temporary freeze refresh commands were removed.
- Coverage: 81.2% overall; engine 86.6%. The pre-existing 90% engine target deficit
  is unchanged in policy and not hidden by lowering a gate or dropping tests.
- No client source, live service, user save, database repair or installed asset write.
