# CTP2 Retail Gameplay Completion Sync Plan

## Baseline and Reconciliation

- The `.sisyphus/open-ctp-server-gameplay-systems` worktree must start from the same tracked content as the project main branch before any new gameplay work.
- Preserve potentially valuable dirty work before destructive cleanup. The reconciliation created stashes for the prior main-worktree dirt and prior gameplay-worktree dirt before changing branch pointers.
- Keep the meaningful main-only change: `AGENTS.md` project guidance.
- Prefer the newer gameplay-systems branch content over older dirty main copies where the same files diverged.
- Fast-forward the project branches to the reconciled baseline, then continue only from the gameplay worktree.

## Goal

Make the server-authoritative Call to Power II gameplay runtime complete enough for an appropriate client to play a full base-game match. The server must accurately execute retail CTP2 rules for advances, units, wonders, buildings, stealth, pollution, unit actions, goods, bombardment, zone of control, nukes, eco-terrorist conversion, terrain, governments, workers, government sliders, diplomacy, victory, and all other gameplay-affecting systems.

## Methodology

- Use red-green-refactor for every behavior slice.
- Source-check ambiguous mechanics against `ctp2/ctp2_source_code` and retail data under `ctp2/ctp2_released_game` before implementing.
- Keep each behavior slice atomic and commit after the green/refactor validation.
- Keep `plans/checklists/ctp2-retail-gameplay-completion-sync.md` current after every slice.
- Do not modify or bypass the pre-commit hook.

## Execution Order

1. **Reconciliation closeout**: record the synced baseline, create this plan/checklist, and commit it.
2. **Full-game progress harness first**: add a failing test/harness that proves autonomous retail-runtime play remains mechanically active over long runs with useful metrics for cities, units, production, research, gold, trade, diplomacy, pollution, combat, commands, restart equality, and terminal state.
3. **Economy and production continuity**: fix AI and engine blockers that leave nations without units, cities, queues, research, public works, or gold flow in long games.
4. **Exploration, workers, terrain, goods, and trade**: ensure workers and public works improve terrain, goods feed city requirements, trade routes matter, and AI can keep expanding/exploring without hidden-state shortcuts.
5. **Diplomacy, war, combat, bombardment, ZOC, stealth, and nukes**: close coarse full-game loops for contact, treaties, war, targeting, bombardment, active defense, city capture, stealth detection, and nuclear/special attacks.
6. **Pollution, disasters, governments, sliders, happiness, crime, and endgame**: ensure the late-game systems run without crashes and use retail coefficients and source-backed effects.
7. **Field-coverage closure**: return to field-by-field catalog parity only after full-game mechanics progress; no gameplay-affecting retail field may remain parse-only or unexecuted.
8. **Final verification and review**: run freeze guard, relevant package tests, full repository tests where feasible, long retail AI artifacts, `git diff --check`, and a full code review against the plan.

## Current Deferred/Preserved Work

- A pre-sync worktree stash contains an in-progress retail AI long-run production/unit-presence test and AI production-order tweak. Reapply only after this plan/checklist is committed and treat it as the RED starting point for the full-game progress harness.
- A pre-sync main-worktree stash preserves older dirty files. Do not reapply wholesale; compare individual files only if a missing behavior is suspected.
