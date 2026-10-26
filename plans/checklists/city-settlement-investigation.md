# City settlement rejection investigation — game97

Investigation only; do not submit commands to or modify the live game. Findings: [report](../../docs/city-settlement-rejection.md).

- [x] Read supplied logs and inspect both server/client working trees without disturbing pending client changes.
- [x] Trace client Settle input through the exact advertised found-city payload and server rejection classification.
- [x] Inspect game97 state/receipt/event history using read-only queries, distinguishing national ownership from city influence.
- [x] Audit original released CTP2 CanSettle/Settle guards and current territory initialization.
- [x] Reproduce the exact advertised command from copied revision17/turn3 state: foreign-national/empty-influence site rejects; changing only national owner to neutral/own succeeds; removing settler capability still rejects. All results remain in memory.
- [x] Repeat all four cases with original b8fbda9 compiler overlays; identical results. Founding/territory implementations unchanged by performance commits.
- [x] Trace the full released army/event/player/unit/create-city/insert-city chain; distinguish national owner from city influence and preserve terrain/movement/adjacent-city constraints in follow-up recommendations.
- [x] Record the server/advertised-action/AI mismatch, safe error/log behavior, receipt-history limitations and a source-backed repair outline without implementing it.
- [x] Existing focused engine founding and isolated native client settlement tests pass. Report/links/whitespace reviewed; runtime code, live data and pending client source changes remain untouched.
