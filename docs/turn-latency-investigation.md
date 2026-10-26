# Why passing turns takes seconds

**Historical investigation.** The follow-up implementation and measured equivalence/results are documented in [near-instant turns](instant-turns.md). Findings below describe the pre-fix behavior.

## Scope and conclusion

Investigated server `b8fbda9` and client `256c995`, including the running local game's read-only database records. No gameplay commands were submitted, no application records were changed, and no runtime code was changed. Unrelated client working-tree changes appearing during the investigation were left untouched.

**There are two substantial, independent contributors:**

1. **The client deliberately spaces viewer refreshes 3.1 seconds apart.** AI progress is delivered as invalidations, not state, so the displayed turn can lag the actual server turn. A refresh waiting for admission also blocks the client's command worker.
2. **The server does considerable repeated work for each AI command.** Three AI opponents advance serially. Recorded AI transaction starts are commonly about 0.45–0.69 seconds apart; an offline pass command's core CPU pipeline alone measured about 206 ms on a small map. Repeated rules canonicalization/compilation and viewer projection are substantial costs, not AI strategy.

These costs overlap; do not simply add all figures below to predict an exact click-to-display duration. The reported one-to-two-second pauses are credible from server work alone, and the client can add seconds of stale display or queued-command latency.

## Client findings

Paths in this section are relative to `$CTPCLIENT`.

### 1. Refresh admission creates seconds of latency (confirmed)

`src/gameplay/runtime/network/recovery.rs`:

- `READ_INTERVAL_MS = 3_100`.
- `ReadGate::new()` starts the next available slot at 3,100 ms because bootstrap already used the same endpoint.
- `reserve_at()` spaces subsequent requests another 3,100 ms apart.
- The event watcher and current-snapshot requests share this gate.

This is not an arbitrary timer to remove in isolation. The server registers `GET /game/{game_id}/events` under `NewLobbySSELimiter()`, which allows **20 connections/minute**. Both persistent subscriptions and `X-Gameplay-Resync: current` reads use that route (`internal/runtime/runtime.go`, `internal/ratelimit/ratelimit.go`).

`src/gameplay/server/stream.rs` consumes revision/cursor hints only. The server's normal events likewise contain metadata, not a replacement viewer state (`visibility.Projector.ProjectEvent`). Consequently, when an AI advances the game, the client must issue a separate current read before showing the result.

In steady state, a recently consumed read slot can leave a subsequent refresh waiting almost 3.1 seconds. At startup, if the watcher reserves the first slot, it cannot connect before approximately +3.1 seconds, and a needed current read reserves no earlier than approximately +6.2 seconds. Reconnection reservations can extend waits. Idle sessions whose next slot is already in the past do not incur the full delay on every turn.

A current read made midway through the AI sequence may return an intermediate revision. The next event then needs another admitted read. Coalescing prevents redundant reads for already-known revisions, but does not make an intermediate snapshot the final turn result.

### 2. Refresh waiting blocks commands too (confirmed by call chain)

`src/gameplay/runtime/network.rs::Network::start` uses one worker for both command execution and current reads:

1. `requests.recv_timeout(100 ms)` receives user commands immediately when available.
2. On timeout, newer observed revisions trigger `refresh()`.
3. `HttpTransport::resync()` calls `gate.wait()` before the blocking HTTP request.
4. While that worker is waiting or reading, it cannot receive another command.

Thus, even though `submit()` is nonblocking on the SDL thread, a user command can sit behind a multi-second background refresh. The one-element command channel does not provide preemption.

The **100 ms receive timeout is an additional event-refresh scheduling delay**, not a universal 100 ms delay on every command. An idle worker is woken by an arriving command. Also, the stream's 250 ms socket read timeout is handled by continuing to read; it does **not** deliberately reconnect every 250 ms.

### 3. Ordinary receipt/state processing is much smaller (measured)

`runtime/sync.rs` applies complete snapshots monotonically. The ordinary successful command path uses its receipt directly and does not require a background current read. State replacement clones the previous state, builds a render model, and retains a cloned viewer projection. The session also owns a projection.

A standalone in-memory diagnostic linked against the existing **debug** client library used a recorded human viewer projection, not a live HTTP request. Thirty repetitions, including result destruction, gave:

| Operation | Median |
| --- | ---: |
| Deserialize 234,983-byte viewer projection | 10.93 ms |
| Clone viewer projection | 2.07 ms |
| Clone `GameState` | 1.80 ms |
| `GameState::from_server_projection` | 3.11 ms |

The sample contained 71 exposed/remembered tiles, two units and two cities. These are individual measurements, not a measurement of the entire reconciliation callback or SDL frame. They do not suggest that state conversion alone explains seconds on this small game.

The frame loop polls one result per frame and sleeps 16 ms **in addition to** rendering/presentation time (`runtime/session.rs`). Rendering, GPU presentation and first-use art uploads were not profiled; larger maps, large managers or new art can still hitch. Textures are retained in caches across state replacements, rather than universally reloaded on every turn.

Each blocking HTTP request opens a new TCP connection with `Connection: close` (`server.rs::request`). This adds overhead but is unlikely to dominate a local request compared with the explicit read gate and measured server work. Network latency on remote deployments was not measured.

## Server findings

### 4. Three AI turns really do take substantial time (recorded evidence)

Read-only local records for game 96 show a 384-tile, four-player game. One observed sequence:

| Revision | Controller | Event | Gap from preceding transaction start |
| --- | --- | --- | ---: |
| 93 | human | `command.queued` | — |
| 94 | AI | `command.queued` | 686.0 ms |
| 95 | AI | `command.queued` | 483.3 ms |
| 96 | AI | `turn.resolved` | 446.0 ms |

Another sequence, revisions 88–91, had gaps of 606.3, 468.3 and 457.9 ms.

**Boundary caveat:** receipt `created_at` defaults to PostgreSQL `now()`, which is the transaction start timestamp, not the commit timestamp. These sequences show roughly 1.5–1.6 seconds from the human command's transaction start to the final AI command's transaction start, with the final transaction still to complete. They are not direct end-to-end stopwatch or per-command execution measurements.

`main.go::runGameplayWorker` polls at 250 ms. `AIWorker.RunOnce` can drive multiple players in one invocation, so there is **not necessarily a fresh 250 ms sleep between every opponent**. However, initial scheduling can wait for the worker, and work across started games is serial. Selection repeatedly takes `due[0]` in game order, with a 32-drive per-invocation budget. Busy earlier games can delay later games.

### 5. Repeated rules compilation/projection is a major CPU cost (measured)

Offline Go diagnostics used a read-only copy of canonical state streamed through stdin. The command diagnostic ran the real pure engine against that copy, using diagnostic clock/ID/RNG implementations; it never persisted results or contacted gameplay HTTP endpoints. Each operation ran twelve times with normal Go build optimization, while local applications remained running.

Representative second sample: 384 tiles, four players, four cities, five units, 116 fog observations, sequential mode, 216,299 bytes of pinned rules and 334,991 bytes of canonical state.

| Operation | Median |
| --- | ---: |
| `DecodeCanonicalStateJSON` | 22.31 ms |
| `ValidateState` | 6.51 ms |
| `rules.CanonicalDefinitions` | 7.20 ms |
| `rules.CompileRuntime` | 11.04 ms |
| One `visibility.Update` | 19.97 ms |
| `CanonicalStateJSON` | 8.41 ms |
| `engine.Apply` of active player's pass | 57.46 ms |
| Core command CPU pipeline, excluding database | **205.88 ms** |
| Human `visibility.AIInput` | 96.71 ms |
| Individual AI `visibility.AIInput` | 65.46–66.05 ms |
| Individual AI `Plan` | **2.72–3.50 ms** |

The core pipeline was decode → pre-command visibility → engine apply → post-command visibility → actor viewer projection → canonical serialization. It omits service/store validation, snapshot rules-pin processing, receipt/event encoding, SQL, commit and dispatch. Stage measurements overlap internally and should not be summed with the whole-pipeline result. An earlier sample produced projections around 64–94 ms and AI planning around 2–3 ms.

The offline workload's CPU profile attributed about **51% cumulative sampled CPU to `rules.CompileRuntime`**, dominated by canonical JSON processing and decoding. This is a profile of the diagnostic workload, not a production-wide CPU percentage or an allocation profile.

Relevant repeated work:

- `service.executeControllerCommand` decodes state, updates visibility **before and after** every command, builds a full actor projection and serializes the aggregate/receipt, all within the command transaction.
- `engine.DecodeCanonicalStateJSON` validates and then performs a canonical JSON round trip to clone state; `Apply` also canonical-clones state.
- `visibility.Update` and `Project` construct inspectors, which compile pinned rules.
- Projection helpers call `runtimeRules` separately for rules/research/other action metadata; `runtimeRules` compiles again. Each `CompileRuntime` canonicalizes and verifies the definitions before decoding the runtime.
- `AIGateway.Projection` separately reads/decodes the aggregate and builds a full projection before AI planning. The subsequent command creates another full projection, stored in the receipt; the AI gateway then returns only revision/terminal status to the driver.
- Some AI drives submit an action **and** a closeout, each through the full command gateway. The recorded sequences above happened to use one command per AI.

The response projections in the observed chronology were roughly 207–254 KB. Full rules/display data and full viewer receipts are processed repeatedly even for pass/closeout commands.

### 6. Worker discovery repeatedly reads every started game (measured)

`AIWorker.findDuePlayers` calls `FindStartedGameplayAggregates`, deserializes all returned states and selects the first due player. It repeats this after every drive, including one final scan to discover there is no due AI.

The local database contained **27 started games**, approximately **2.96 MB** of aggregate JSON per scan, many unrelated to the currently played game. Deserializing the complete array without full engine validation measured **27.28 ms median**. PostgreSQL's existing `pg_stat_statements` reported the discovery query at about **20.76 ms mean / 107.80 ms maximum**, over about 16,000 calls. These are historical aggregate SQL statistics, not a profile of this one turn.

At a nominal four idle scans per second that is approximately 12 MB/sec of state text before allowing for actual worker scheduling; active AI drives perform additional scans. Abandoned started games therefore add work to the current game's AI latency even when they have no due AI.

For comparison, historical local aggregate row-lock/load SQL averaged about 6.75 ms and aggregate UPDATE about 5.30 ms. Database work and commit overhead still matter, but these individual queries do not account for the entire roughly half-second AI cadence. Network transfer, other SQL, locks and service-side work are not included in those two means.

### 7. Event dispatch adds polling latency, but is not the only issue

The standalone outbox dispatcher polls at 250 ms; dispatch work and scheduling can extend that interval. AI worker `AfterCommit` also dispatches after a successful drive, so AI notifications need not always wait for the standalone poll. The broker wakes the stream, which rereads durable events. Those events still only invalidate the client model, leading back to the throttled current read.

Do not remove outbox durability or acknowledgement/revision checks to improve responsiveness.

## Original released CTP2 comparison

Authoritative local source: `ctp2/ctp2_source_code/ctp2_code/`.

- `ui/interface/EndTurnButton.cpp::EndTurnButtonActionCallback` verifies the visible/current player, registers manual end-turn and queues `g_director->AddEndTurn()`.
- `gs/utility/newturncount.cpp::NewTurnCount::StartNextPlayer` validates ending the turn and calls the current player's `EndTurn`, then advances the next player through director/network handling.
- `gs/gameobj/Player.cpp::Player::EndTurn` performs city/unit/diplomacy/pollution turn work and host networking.

Opponent processing and authoritative turn sequencing are legitimate work. The **3.1-second HTTP refresh budget is an integration artifact of this implementation**, not a retail turn mechanic. Optimize redundant processing and delivery without bypassing authority, changing AI action semantics, skipping upkeep or inventing a local turn.

## Recommended fix order (not implemented)

1. **Repair the update-delivery contract jointly.** Separate subscription/reconnect admission from current-state read capacity, with explicit bounded per-viewer budgets. Alternatively/additionally deliver complete viewer-safe snapshots on the existing stream with revision/cursor metadata. The current parser deliberately discards snapshot payloads, so a server-only payload change is insufficient. Do not merely lower the client interval against the unchanged 20/minute limiter.
2. **Prevent background recovery from monopolizing command submission.** Make admission waiting preemptible or separate read/command execution, retaining one outstanding user operation, immutable retries, explicit receipt acknowledgement and monotonic revision reconciliation.
3. **Reuse verified immutable rules and request-scoped derived state.** Avoid repeated canonicalization/compilation across the same command/projection. Any cache must be bounded, content/hash/version verified, concurrency-safe, and protected against accidental mutation of shared slices/maps. Preserve visibility/privacy and deterministic replay.
4. **Discover only due games/players and schedule fairly.** Avoid full JSON scans of every started game after each AI. Consider commit-driven per-game scheduling plus bounded periodic recovery. Do not parallelize mutations within a game or remove expected-revision/receipt transactions.
5. **Reduce redundant AI projection and large static-metadata work.** Reuse authorized projection results where safe and explore negotiated rules/static-data caching. Do not silently omit fields from the current full-snapshot contract or weaken durable replay receipts.
6. **Measure end-to-end before declaring success.** Track command capture, worker queue wait, admission, HTTP/service stages, AI chain completion, event observation, snapshot revision and UI presentation. Existing client DEBUG logs include queue/admission/HTTP/reconciliation durations; HTTP responses include `Server-Timing`. Server `ObserveCommand` has stage timings, but its default observer uses `slog.Debug`, filtered by the default logger; `main.go` currently does not install a debug-enabled handler. Do not assume setting an arbitrary log-level environment variable enables server stage logs.

For each implementation batch: write failing fake-time regressions first, implement GREEN, then refactor and run the appropriate gates. Important cases include prompt AI-state visibility, command admission during a background wait, request-budget exhaustion, disconnects, intermediate snapshots, old receipts, phase restrictions, and hidden-state/receipt privacy. Never use live gameplay or filesystem-modifying tests to reproduce performance bugs.

## Verification and limitations

Existing focused tests passed:

```sh
# Server: pure/fake-store AI worker behavior
 go test ./internal/gameplay/service -run 'Test(DueAIPlayersForState|AIWorker)' -count=1
# Client: fake-time read admission/revision coalescing (3 tests)
 cd "$CTPCLIENT"
 cargo test --lib gameplay::runtime::network::recovery_tests
# Client: injected outcome/receipt reconciliation (2 tests)
 cargo test --lib gameplay::runtime::sync_tests
```

The admission regression explicitly requires successive slots at least three seconds apart, corroborating that the current delay is intentional behavior rather than an accidental sleep elsewhere. These tests verify existing semantics, not an implemented performance fix.

Local offline diagnostic binaries/profile remain at `/tmp/ctp-turn-latency`, `/tmp/ctp-turn-latency.pprof` and `/tmp/ctp-client-turn-latency`; diagnostic sources are `/tmp/ctp-turn-latency.go` and `/tmp/ctp-client-turn-latency.rs`. They print sizes/timings only, not session IDs or canonical contents. The Go binary accepts one canonical state on stdin, or an array of states with argument `scan`; its optional first argument otherwise names a CPU profile file. Inspect the retained CPU profile with:

```sh
go tool pprof -top -cum /tmp/ctp-turn-latency /tmp/ctp-turn-latency.pprof
```

No automated end-to-end click-to-present recording, live process CPU profile, GPU trace, remote-server benchmark, large-map benchmark or post-fix verification was performed. Local state evolved while the user played, so the two server samples, receipt chronology and client sample are different snapshots; their exact numbers should not be treated as a single synchronized trace. No new executable tests or runtime changes were introduced; a full build/lint/coverage pass was not necessary for this documentation-only investigation.
