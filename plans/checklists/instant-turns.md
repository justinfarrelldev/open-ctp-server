# Near-instantaneous turns implementation

Plan: [instant turns](../instant-turns.md). Results/limits: [evidence](../../docs/instant-turns.md).

## Verified rules and state reuse
- [x] RED: missing cache/clone helpers; cached/uncached equivalence, deep ownership, hash/content corruption, concurrent callers, bounded eviction/oversize and nil-versus-empty RawMessage regression.
- [x] GREEN: bounded exact-input canonical/compiled rules, metadata and decoded-state reuse; all verification/validation remains enforced and mutable returns own their graphs.
- [x] RED/GREEN execution-copy equivalence: JSON omission/null/empty, escaping/key order, invalid UTF-8, custom scalar/slice codecs and audit times. Canonical serialization's shallow fog-effect mutation reproduced/fixed without changing persisted bytes.
- [x] REFACTOR/review: allocation-free exact-byte lookups, FIFO eviction releases old keys, cold loader serialization never holds the warm-reader map lock; blocked-cold/warm-hit RED and race checks pass. Schema/ownership tests guard future fields.
- [x] Typed AI projection and indexed A* neighbor enumeration preserve exact candidate/closeout submissions, legacy metadata-only closeout, route costs/order and gameplay/event hashes.

## Server scheduling and delivery
- [x] RED: missing commit wakeups, targeted discovery, snapshot negotiation, interactive-rate refusal and global-outbox waiting observed before implementation.
- [x] GREEN: immediate commit hints, bounded FIFO/coalescing and durable periodic recovery; one-game scheduling reads, no unrelated world/rules/fog transfer.
- [x] GREEN: authenticated negotiated full snapshots, subscribe-before-read and independent connection budget; legacy SSE/current behavior and durable outbox unchanged.
- [x] REFACTOR/review: cancellation, bounded work/queue, controller/phase/revision authority, same action budget/order, terminal/privacy and no synchronous global outbox drain between AI players.
- [x] All server packages pass after updating stale SQL-memory adapter, exact autonomous-entry assertion and frozen production inventory100; no freeze scope/import policy/hook/threshold weakened. Relevant race suites and go vet pass.

## Client
- [x] Existing client PLAN/checklist updated; unrelated font work left untouched.
- [x] RED: missing full-snapshot/admission/worker helpers, malformed diagnostic privacy and late optional-read failure; strict fragmented bounds/metadata, receipt ordering and legacy behavior.
- [x] GREEN: immediate capability-gated subscription, full-model mailbox/coalescing and independent background/command transports; no second current read on healthy path.
- [x] REFACTOR/review: bounded memory/channels, shutdown, monotonic state, receipt-only acknowledgement and serialized completion generation prevent same-revision/stale failure races without locks during network waits. Held-ledger/full-channel SDL race reproduced RED and fixed by nonblocking try_lock mailbox polling; model remains retained for another frame.
- [x] Isolated client branch `perf/instant-turns`, commits `48687bb` + `5cc0bd4`: required lint-fix/strict lint, focused regressions and fresh final833-test/86.58% LLVM line coverage pass. Initial37 image-fixture failures were repaired using only a missing read-only reference symlink; no old mappings reused after the final source change. Both normal unmodified commit hooks pass; final hook runs833 library+21 integration tests and native build.

## Integration
- [x] Original released EndTurnButton/director/NewTurnCount/Player and A* chain reviewed; no AI/turn/upkeep/physics/authority/UI/scaling change.
- [x] Fixed-snapshot seven-repetition offline rounds:1591->80ms (earlier4-player),1521->123ms (larger); exact AI command sequences and final gameplay/event hashes match. One-subscription/receipt/AI-next-turn/no-current client workflow passes. Live SQL/network/GPU/large-world latency is explicitly unmeasured, not claimed instantaneous.
- [x] Protocol/investigation evidence and production digests updated. Server internal statement coverage81.46%; engine cross-suite87.9%, optional Taskfile90% engine target remains unmet and is reported without lowering it.
- [x] Client atomic commit/hooks verified on an isolated clean worktree to avoid concurrent font edits; its turn source files match those already present in the main client checkout. Server atomic commit uses its unchanged normal hook; unrelated work is preserved.

Remaining deployment/acceptance: integrate the independently committed client branch with the concurrent font work when its owner is ready, restart both binaries, and measure live click-to-present latency. These are not represented as completed live acceptance or a universal latency guarantee.
