# CTP2 Retail Gameplay Completion Sync Checklist

## Reconciliation

- [x] Saved dirty project-main worktree state before cleanup.
- [x] Saved dirty gameplay-systems worktree state before cleanup.
- [x] Compared dirty main files against gameplay-systems content and kept the newer gameplay-systems branch as the authoritative baseline for overlapping gameplay files.
- [x] Merged main-only `AGENTS.md` guidance into `gameplay-systems-completion`.
- [x] Fast-forwarded `game-server-world`, `gameplay-systems-completion`, and local `main` to the same reconciled commit.
- [x] Verified project root and gameplay worktree were clean after sync.
- [x] Created the new top-level plan and checklist.

## Full-Game Mechanical Execution First

- [ ] Add/restore a RED long-run retail AI progress test that fails when autoplay becomes inert, loses all units, stops producing/researching, or issues only round-complete commands late-game.
- [ ] Extend the full-game progress record with cities, units, active queues, research progress, advances, gold/treasury, public works, trade routes, diplomacy/war, pollution, combat counts, command throughput, terminal state, and state hash.
- [ ] Make the progress harness checkpoint canonical restart equality.
- [ ] Fix AI economy/production continuity using only public projections and legal actions.
- [ ] Fix AI exploration, settling, worker/public-works improvement, transport/cargo/fuel dead ends, and pathing stalls.
- [ ] Fix coarse diplomacy/trade/goods loops so trade agreements, routes, goods-gated production, and gold flow materially affect long games.
- [ ] Fix coarse combat loops for war/contact, legal targeting, bombardment/active defense, ZOC, stealth detection, city defense/capture, and conquest progress.
- [ ] Fix pollution/disaster/government-slider/readiness/happiness/crime/endgame loops for long-game stability.

## Retail Parity Closure

- [ ] Complete runtime execution for all gameplay-affecting advance fields.
- [ ] Complete runtime execution for all gameplay-affecting unit fields and actions.
- [ ] Complete runtime execution for all gameplay-affecting wonder fields.
- [ ] Complete runtime execution for all gameplay-affecting building fields.
- [ ] Complete runtime execution for all terrain, worker, public-works, goods, and trade effects.
- [ ] Complete runtime execution for all government, diplomacy, slider, readiness, happiness, crime, and pollution effects.
- [ ] Complete runtime execution for all victory/endgame/space/Gaia/final-turn mechanics.
- [ ] Update the retail field coverage report so no behavior-bearing retail field remains unimplemented or parse-only.

## Verification and Review

- [ ] Run targeted RED/GREEN/REFACTOR evidence for every behavior slice.
- [ ] Run package tests for changed packages after each slice.
- [ ] Run freeze guard after production `.go` changes.
- [ ] Run `git diff --check` before every commit.
- [ ] Run long retail AI artifact recording after full-game progress fixes.
- [ ] Run full repository validation where feasible before final review.
- [ ] Complete and record a final maintainability/source-parity code review.
