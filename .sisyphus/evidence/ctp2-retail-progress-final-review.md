# CTP2 retail gameplay completion final review

Review date: 2026-09-28

Scope reviewed:
- Task 026 conquest/capture-route recovery commit `071f900`.
- Public-projection capture movement helpers in `internal/e2e/fullgame/runner.go` and `exploration.go`.
- Stale-tile planning domain ordering in `internal/gameplay/visibility/visibility.go`.
- Acceptance hash/freeze-lock updates for the deterministic row-one conquest scenario.

Findings:
- Capture planning continues to consume only public projections and projected `move-unit` legal actions for the step it actually executes. Remembered/stale tiles are used only for route selection; execution is limited to the current projected first step.
- Co-located reserves remain available after conquest combat because attacker army assembly now uses only the selected attacker. Capture fallback can switch to a visible legal reserve when the preferred unit is spent.
- Stale-memory oscillation is bounded by visited-tile exclusion, recovery reset, and a bounded move trace in errors.
- `PlanKnown` stale tile domains are sorted before binary-search membership checks.
- The previous diagnostic capture-budget inflation is absent; conquest still uses configured action/turn budgets.
- Freeze and acceptance hashes were regenerated after the production visibility change.

Validation reviewed:
- Targeted Task 026 green evidence: `.sisyphus/evidence/ctp2-retail-progress-task-026-green.txt`.
- Package evidence: `.sisyphus/evidence/ctp2-retail-progress-task-026-package-tests.txt`.
- Full repository validation: `.sisyphus/evidence/ctp2-retail-progress-final-validation.txt` (`go test ./... -count=1`).
- Whitespace validation: `git diff --check` passed before commit.

Conclusion:
- No follow-up defects found in the reviewed Task 026 implementation.
