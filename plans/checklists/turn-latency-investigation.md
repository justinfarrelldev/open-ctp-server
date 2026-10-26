# Turn latency investigation

Investigation only; findings and follow-up recommendations: [report](../../docs/turn-latency-investigation.md).

- [x] Read server/client instructions and inspect both working trees without changing client files.
- [x] Trace end-turn input, immutable command submission, receipt handling and UI reconciliation.
- [x] Trace SSE invalidation, shared read admission, refresh scheduling and command-worker contention.
- [x] Trace server AI scheduling, transactional command execution, persistence and event delivery.
- [x] Compare turn progression with the original released CTP2 source.
- [x] Inspect local database sizes, command chronology and aggregate SQL timings using read-only queries.
- [x] Measure canonical decoding, rules compilation, visibility, projections, engine pass and AI planning in memory, without submitting live commands.
- [x] Profile the offline server workload and measure existing debug-client snapshot decoding/cloning/rebuilding in memory.
- [x] Run focused existing fake-time/client reconciliation and fake-store/server AI-worker tests.
- [x] Review causal claims, timing boundaries and recommendations; distinguish measurements from unmeasured possibilities.
- [x] Record findings, prioritized fixes, verification commands and remaining limitations.

No runtime implementation was requested or performed. Any follow-up fixes require new failing regression tests before implementation, followed by GREEN, refactoring and normal repository verification.
