# CTP2 Retail Gameplay Completion Sync Checklist

## Reconciliation

- [x] Saved dirty project-main worktree state before cleanup.
- [x] Saved dirty gameplay-systems worktree state before cleanup.
- [x] Compared dirty main files against gameplay-systems content and kept the newer gameplay-systems branch as the authoritative baseline for overlapping gameplay files.
- [x] Merged main-only `AGENTS.md` guidance into `gameplay-systems-completion`.
- [x] Fast-forwarded `game-server-world`, `gameplay-systems-completion`, and local `main` to the same reconciled commit.
- [x] Verified project root and gameplay worktree were clean after sync.
- [x] Created the new top-level plan and checklist.

## Completed Increments

- [x] Task 001: Long-run retail AI progress test now guards against zero-unit/inert autoplay; AI now queues emergency defenders for populated undefended cities while preserving expansion-settler preference, stale projected combat is suppressed, and combat can target currently visible units without requiring prior fog memory.
- [x] Task 002: Retail AI progress records now include command throughput, active queues, research targets/progress, treasury/public works, trade/diplomacy/war, pollution, combat, terminal, and state-hash metrics.
- [x] Task 003: Retail AI progress records now checkpoint canonical JSON restart equality on every recorded turn.
- [x] Task 004: Accepted combat now establishes mutual war state, clears trade treaty state, and cancels combatant trade routes, matching CTP2 hostile-action diplomacy/trade cleanup.
- [x] Task 005: Trade routes now carry source goods, add source gold yield, and let destination production satisfy NeedsCityGood prerequisites from imported goods.
- [x] Task 006: Public-works terrain improvements can be queued without worker units on owned/city-radius tiles, are projected as legal actions, and are selected by AI planners.
- [x] Task 007: Retail AI strategy economy sliders are imported and the planner uses the strategy science percent instead of forcing 100% science/no-gold economy.
- [x] Task 008: Public-works rate is a gameplay command, AI uses retail strategy public-works percent, and city production is diverted into nation public works before queue progress.
- [x] Task 009: Rush-buy is a legal projected economy action; AI can spend accumulated gold to buy queued production front items into completion via the normal turn pipeline.
- [x] Task 010: Science-rate commands now persist the government-capped science percentage, matching retail tax-rate clamping.
- [x] Task 011: Government trade caps from retail `govern.txt` are projected into runtime rules, enforced when creating trade routes, and suppress invalid projected legal actions.
- [x] Task 012: Retail government rush-buy coefficients now project from `govern.txt`, engine rush-buy costs use the active government/item type modifier, and legal actions suppress unaffordable retail-modified rush buys.
- [x] Task 013: AI planners now consume projected legal trade-pact proposal actions so trade agreements can precede goods trade routes without hidden-state shortcuts.
- [x] Task 014: AI planners now consume projected legal peace proposal actions so combat-created wars have a public-projection treaty recovery path.
- [x] Task 015: AI planners now consume projected legal bombard actions so bombard-capable units participate in combat planning without hidden-state shortcuts.
- [x] Task 016: AI terrain-improvement planning now preserves retail `PublicWorksReserve` instead of spending public works/materials down to zero.
- [x] Task 017: Retail AI `SliderElement` social-policy deltas are imported, projected, exposed as legal actions, and consumed by AI planning.
- [x] Task 018: AI planners now consume projected legal `upgrade-unit` actions so retail `UpgradeTo` chains can advance autonomous units.
- [x] Task 019: AI planners now consume projected retail special city/tile actions for soothsaying and park creation.
- [x] Task 020: Retail AI social-slider projection now requires visible city fit evidence before changing workday/wages/rations, matching governor fit checks and preserving long-game production.
- [x] Task 021: Terrain improvement `HealRate` now executes during round advancement by healing owner-compatible units on completed improvement overlays.
- [x] Task 022: Terrain improvement `VisionRange`/`RadarRange`/`ListeningPost` now extend player visibility from completed improvement overlays.
- [x] Task 023: Terrain improvement `Airport` now acts as an owner-compatible refuel point during round advancement.
- [x] Task 024: AI now suppresses fragile legacy `terraform` unit commands and relies on validated public-works `build-improvement` terraform paths instead.
- [x] Task 025: Trade-route legal actions now require current visible source/destination cities, preventing stale fog-memory routes from reaching AI submission.
- [x] Task 026: Public conquest capture recovery now keeps reserve capture units out of combat, switches to visible mobile legal reserves when needed, routes remembered city targets one projected legal step at a time, and avoids stale-memory frontier oscillation.
- [x] Task 027: Advance-enabled capitalization and infrastructure production now projects from retail advances/government coefficients, remains public-advance gated, and converts city production into gold or public works with retail ceiling semantics.
- [x] Task 028: Retail `TriggerFeat` advances now import feat definitions, apply timed nation effects on research completion, and execute the gunpowder city-wall reduction path.
- [x] Task 029: Retail feat catalog fields are included in the field inventory snapshot and coverage report now that `feat.txt` is parsed for advance-triggered feats.
- [x] Task 030: Retail building-triggered feats now import `Building` thresholds and apply timed nation effects when city production completes the required building count.
- [x] Task 031: Retail `FEAT_CONQUERED_BY_FORCE` now imports `MinimumSizeOfCiv`, tracks nation max city count, and applies the hit-point feat when a qualifying civilization loses its final city.
- [x] Task 032: Retail feat boat-movement effects now contribute to sea/amphibious movement refresh alongside wonder boat-movement bonuses.
- [x] Task 033: Retail feat hit-point effects now add to produced units' current and maximum hit points while active.
- [x] Task 034: Retail city founder identity is preserved and `FEAT_CITY_RECAPTURED` is awarded when a founder retakes a city.
- [x] Task 035: Retail feats now record global achievement IDs and cannot be awarded repeatedly after first success.

## Full-Game Mechanical Execution First

- [x] Add/restore a RED long-run retail AI progress test that fails when autoplay becomes inert, loses all units, stops producing/researching, or issues only round-complete commands late-game.
- [x] Extend the full-game progress record with cities, units, active queues, research progress, advances, gold/treasury, public works, trade routes, diplomacy/war, pollution, combat counts, command throughput, terminal state, and state hash.
- [x] Make the progress harness checkpoint canonical restart equality.
- [x] Fix AI economy/production continuity using only public projections and legal actions.
- [x] Fix AI exploration, settling, worker/public-works improvement, transport/cargo/fuel dead ends, and pathing stalls.
- [x] Fix coarse diplomacy/trade/goods loops so trade agreements, routes, goods-gated production, and gold flow materially affect long games.
- [x] Fix coarse combat loops for war/contact, legal targeting, bombardment/active defense, ZOC, stealth detection, city defense/capture, and conquest progress.
- [x] Fix pollution/disaster/government-slider/readiness/happiness/crime/endgame loops for long-game stability.

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

- [x] Run targeted RED/GREEN/REFACTOR evidence for every behavior slice.
- [x] Run package tests for changed packages after each slice.
- [x] Run freeze guard after production `.go` changes.
- [x] Run `git diff --check` before every commit.
- [x] Run long retail AI artifact recording after full-game progress fixes.
- [x] Run full repository validation where feasible before final review.
- [x] Complete and record a final maintainability/source-parity code review.
