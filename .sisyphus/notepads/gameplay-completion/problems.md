# Problems


## SQL-free real-handler E2E composition blocker — 2026-07-19

A deterministic storage replacement below the handler layer cannot be isolated to `internal/e2e`: account/lobby/game/world/auth/gameplay/events public paths own concrete `*sqlx.DB` transactions and queries. The viable work is a cross-package production abstraction effort (repository/session/game/world/gameplay/event interfaces and SQL adapters), followed by an in-memory implementation. Until that work is done, a full signup/lobby/game/start/SSE/restart flow cannot both use real handlers and avoid all SQL-backed infrastructure.


## Task 17 catalog completion blocker — 2026-08-10

The full `./internal/e2e/fullgame` package remains fail-closed: fixed catalog rows 2–60 lack observed rules hashes, terminal metadata, and state/event artifacts, so `Validate(Catalog)` rejects them. Those rows cover modes, country counts, mechanic families, and victory evaluators the current production public setup/executor does not yet support; no synthetic goldens were written.


## Task 17 mode-adapter prerequisite — 2026-08-10

`internal/gameplay/engine/turns` contains all four policy implementations, but production `engine.Apply` does not invoke it: `CommandRoundComplete` always calls `advanceRound` and victory evaluation directly, regardless of `State.Phase.Mode`; no production runtime scheduler invokes `TurnPolicy.Expire`. The only public `end-early` route maps to prohibited `CommandEndGame`, not the policy `CommandEarlyEnd`. Thus long-turn/simultaneous/hybrid cannot gain real public phase/deadline semantics by changing the E2E runner; Task 7 policy integration and a production deadline dispatcher are prerequisites.


## Production policy resolver integration blocker — 2026-08-10

Canonical policy persistence is now available, but `turns.Command` deliberately contains ordering metadata only and omits a command payload; `turns.Result.ResolvedCommands` therefore cannot replay queued simultaneous actions through `engine.Apply`. Hybrid resolution also derives conflict components from `turns.War`, while canonical `engine.State` has no diplomatic-war authority to persist or project. Invoking the resolver now would either discard real queued actions or mutate them before policy resolution, violating deterministic gateway semantics. The next prerequisite is payload-carrying queued command envelopes plus canonical/persisted diplomatic war state, followed by an internal non-policy replay path.


## Scheduler slice deferred — 2026-08-10

The incomplete deadline scheduler was intentionally removed after repair. A complete follow-up must add a locked deterministic due-deadline scanner, wire it into the production post-commit loop, retain payload-bound system receipts, and cover DSN-free fake-clock restart/SSE behavior. Queue replay remains verified and unchanged.


## Row 2 reciprocal-embassy terminal blocker — 2026-08-10

After repairing monotonic public projection refreshes, the fixed 48-turn Row 2 flow can establish both human embassies through visible cities, but cannot obtain the evaluator-required reciprocal AI embassies. Completing it without violating scope requires a production AI exploration/reciprocal-diplomacy capability that reaches the visible human city within the immutable scenario budget; changing runner scripts, hidden targeting, victory semantics, vision, or `MaxTurns` is prohibited.

## Task 5 acceptance contradiction — 2026-08-11T12:19:00Z

Task 5 simultaneously requires empty `CapturedCityID` to preserve the projection fallback unchanged and fixed Row 1 to terminate after only adding a non-empty-ID guard. Fixed Row 1 produces an empty ID after `defender_held` and loses the referenced attacker unit, so the fallback must fail. Completing the scenario would require newly authorized orchestration for combat-loss recovery or a changed combat setup, both outside the active item.

## Task 5 acceptance contradiction resolved — 2026-08-11T12:35:23Z

The apparent contradiction came from treating the selected starter as the complete army despite other eligible co-located starters being present in the public bootstrap projection. Publicly composing the full starter army and adding projection/history-only exploration recovery resolves Row 1 within unchanged scenario inputs and budgets. The remaining unfiltered package failure is unrelated catalog fixture authoring (`TestCatalogHasExactlySixtyFixedCoveredCases`), explicitly outside Task 5.

## Task 7 clean freeze boundary — 2026-08-11T14:26:00Z

The last committed boundary is HEAD `212e070880dbce3be56b74446c6fc00a4a2cb97e`, while the accepted Tasks 5–6 `internal/e2e/fullgame/**` package remains inherited and untracked. A clean committed-tree Task 7 baseline or narrow Task 7 commit cannot include that acceptance implementation without absorbing unrelated untracked work, which is forbidden. Coverage/format blockers additionally prevent a truthful pre-freeze PASS.

## Task 7 final boundary — 2026-08-11T15:09:30Z

No technical pre-freeze acceptance blocker remains: coverage is 80.4%, gofmt/diff/build/vet/LSP are clean, lock-race and both race matrices pass, migrations clean up, and broad testing has only the explicit Task 9 catalog red. The inherited fullgame tree remains untracked by mandate, so Task 8 must use the documented HEAD plus Task 7 acceptance-input digest boundary rather than assuming HEAD alone contains the harness.

## Task 9 Row 03 remaining blocker — 2026-08-29T06:57:11Z

Rows 03–12 remain incomplete. After the AI open-fallback correction, Row 03 advances past the previous invalid-pass defect but still stops on `movement obstructed` after an ordinary AI `round-complete` in simultaneous planning. No new row artifacts were written from that nonterminal failure.

- 2026-08-29: Task 9 Row 08 remains provisional after rollback stabilization. Removed speculative Row 08 hybrid bloodlust scheduler and diagnostic work and preserved accepted AI, script, restart, and freeze fixes, but go test ./internal/e2e/fullgame still fails in TestTaskFiveExecutionSemanticsRemainGreen on the accepted conquest path with defender army resolution failure. A contained harness-side retry was reverted after it only changed the failure to turn-budget exhaustion. Row 08 acceptance was not attempted or pinned.

- 2026-08-29: Correction — the first Row 08 stabilization cleanup was incomplete. The executor still contained speculative bloodlustHybridAction / chooseBloodlustHybridAction / driveBloodlustScriptedHuman logic and a call site. This pass removed those leftovers completely, restored the accepted Row 01 conquest path, and only then revalidated the preserved AI, restart, and freeze fixes.

- 2026-08-29: Second cleanup correction — Phase-1 review found leftover Row 08 hybrid scheduler residue after the first correction. This pass removed hybridHumanDebugState, humanDriveDebug, dynamic hybrid planning-status gating, early hybrid-only continue handling, remaining bloodlust helper remnants, and hybrid error augmentation so mixed scheduling is back to the accepted shared simultaneous-or-hybrid planning baseline.
