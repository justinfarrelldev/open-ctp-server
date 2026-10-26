# Near-instant authoritative turn delivery

Implementation of [the latency investigation](turn-latency-investigation.md) and [plan](../plans/instant-turns.md). No AI action budget/ranking, pass/closeout command, upkeep, turn sequencing, combat/RNG, persisted pin or player authority is relaxed.

## What changed

- Exact-input, bounded caches reuse verified canonical/compiled rules, compatibility metadata and decoded states. Incorrect hashes, changed bytes at the same revision, malformed inputs and unsupported versions still fail. Every mutable return owns its complete graph. Warm requests do not wait behind unrelated cold compilation/decoding. Eviction affects cost only.
- Execution cloning preserves the former JSON round trip (omission, null/empty, raw-message escaping/order, custom codecs, UTF-8 and audit time normalization) without serializing/reparsing the whole world. Durable canonical encoding and gameplay/event hashes are unchanged. A pre-existing shallow fog-effect copy was repaired so serialization cannot reorder caller-owned memory.
- AI can consume the **same viewer projection as a typed model** in-process rather than serialize it, read its metadata, then parse it again. Byte-input controllers keep their legacy closeout semantics. Both paths share candidate construction/ranking and exact immutable submissions.
- A coordinate index replaces repeated whole-map scans when enumerating A* neighbors. Original ID tie order, duplicate-coordinate behavior, costs, heuristic, accessibility, first-move, domains, occupancy and ZOC checks remain the same.
- A committed command immediately wakes the AI worker and SSE broker. AI remains serial through the same transactional gateway. Bounded coalesced game hints/FIFO plus periodic durable discovery recover missed hints/restarts. Discovery transfers scheduling metadata, not every started world's terrain/rules/fog; targeted wakeups read one game. AI no longer waits for global outbox draining between players. The durable retry dispatcher and outbox remain intact.
- `snapshot_stream` negotiates complete viewer-safe models on the persistent SSE connection, eliminating the second throttled current read. A separate connection limiter allows the first subscription immediately after bootstrap; reconnects stay bounded. Legacy streams/one-shot reads retain their contract.
- The client uses independent command/background transports and a bounded latest-model mailbox. Snapshots coalesce, but receipts/rejections/recovery cannot disappear or acknowledge the wrong operation. A stale in-flight read/failure cannot reopen recovery after a newer complete delivery. UI mailbox polling uses `try_lock`: if a producer is backpressured after the UI sampled its outcome queue empty, the optional model is retained for another frame rather than blocking SDL. Validation precedes cursor/model delivery, incomplete/malformed data fails closed, and diagnostics do not echo response contents.
- Interactive commands allow 30 requests/second per IP rather than the old two/second bottleneck. Authentication, phase/revision/ownership checks, transactional game locks, client one-outstanding-operation and flood rejection remain enforced. Other limiter classes are unchanged.

## Measured equivalence and latency

Read-only local database snapshots were replayed **offline**, without submitting live commands or changing games. A fixed clock/ID/RNG source ran the real engine, visibility updates/projections, AI driver/planner, action/closeout commands and serialization. Seven warm repetitions per implementation used the same snapshot, `GOMAXPROCS=4`, and the same source lineage. The baseline binary was built using original source overlays before optimization. Every repeated optimized run retained the same gameplay/event artifact hash; the baseline and optimized command sequences and hashes matched exactly.

| Four-player saved game | Original median | Optimized median | Commands |
| --- | ---: | ---: | --- |
| Earlier game95 | 1,591 ms | **80 ms** (78–84 ms) | human pass; each of 3 AIs selects science rate then passes |
| Larger game96 | 1,521 ms | **123 ms** (121–129 ms) | human pass; 2 AIs reform then pass; final AI passes |

Both maps have 384 cells and three AI opponents. The larger sample includes 18 units and 424 fog observations. These are **server computation/serialization timings, not an end-to-end live SQL/network/SDL guarantee**. SQL commits, actual hardware/load, network RTT and rendering add time; larger games necessarily do more work. No fixed multi-second replication wait remains on the negotiated healthy path. The sub-100 ms warm computation target is met by the earlier sample, not claimed universally.

Matching gameplay/event hashes:

- Earlier: `ac103a570abd2e7d02b0ab6228babebc8a77e2813b28bc9ab1dad8cfb0be7bfe`
- Larger: `82f550d648cc19902e41c0df8e26ff49688b20c40069b87440f4d16b059a43f3`

An additional read-only query of 27 started games measured scheduling transfer at **19,932 bytes instead of 3,135,732 bytes** (about 99.4% less), before targeted-game admission. The client fake-time test requires the negotiated first subscription at exactly time zero while keeping the legacy interval unchanged. Its production-worker/isolated-HTTP regression requires a command receipt and a subsequent AI-completed model over one subscription with **no current read**. These verify removal of the delays, not a synthetic claim of measured live gameplay.

Local diagnostic sources/binaries/results remain under `/tmp/ctp-turn-round.go`, `/tmp/ctp-turn-round-{baseline,fast}`, `/tmp/ctp-turn-final-benchmark.txt`, and `/tmp/ctp-turn-baseline-overlay.json`. No canonical snapshot or credentials were saved with this report or committed to the repository.

## Wire/compatibility

See [client gameplay contract](client-gameplay-contract-v2.md). A new client opts in only after seeing `contract.features.snapshot_stream == true`; old servers/clients keep legacy hint/current behavior. Complete snapshot data is not a command receipt, and SSE sequence is never a revision. Client negotiated data is bounded to 16 MiB per complete model; oversized/malformed/inconsistent models require safe recovery, never fabricated local progress.

Restart/rebuild both processes to use the complete path (the development server may auto-reload). Existing pinned games need no migration, reset or new ruleset. An already-running old client binary does not gain the feature merely because its source or server changed.

## Verification status

- All server package tests passed after updating the exact SQL-memory adapter and autonomous-worker/production-file inventory assertions to the implemented interfaces; no freeze scope/import policy or hook was weakened.
- Race tests passed for rules, engine, service, events, store, runtime and rate limiting.
- Internal statement coverage from the broad run: **81.46%**. Engine-only tests: **86.5%**; engine-instrumented unit/integration campaign: **87.9%**. The optional Taskfile engine target of 90% is **not met**; the older engine files independently measure about 86.5%, so this is not claimed as a passing 90% gate or hidden by changing thresholds.
- Client commits **`48687bb` + `5cc0bd4`** on **`perf/instant-turns`**, verified in `/tmp/ctp-client-instant-turns`: required lint-fix/strict lint, focused regressions, and a fresh final **833-test / 86.58% LLVM line coverage** run pass. Initial coverage had 37 missing-reference-fixture failures, repaired using a read-only image link; the final source change was verified afresh, without old mappings. Both unchanged normal commit hooks pass; the final hook runs **833 library + 21 integration tests and native build**. Final review reproduced/fixed the held-ledger/full-outcome-channel UI race with an inert-thread RED/GREEN. The original checkout's turn source files match the isolated branch; unrelated active font work remains untouched, so branch integration is still owner-coordinated.
- Source review: original released `EndTurnButton::EndTurnButtonActionCallback` → director → `NewTurnCount::StartNextPlayer` → `Player::EndTurn`; released UnitAstar costs/order remain unchanged. This is transport and redundant-work optimization, not a second client rules engine.

Live click-to-present and large-world/remote-network acceptance remain explicit follow-up measurements. No live game mutation, asset copying, changed command semantics or arbitrary fixed UI delay is used to manufacture a fast result.
