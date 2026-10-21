# Retained movement, ruin rewards and city governors

Implement the user's three formerly unsupported contracts server-side. This plan
supersedes their deferral in client-report-contracts.md; other deferrals stand.
Use red → green → refactor for each slice, isolated tests, atomic commits and the
unchanged pre-commit hook. Track progress in the matching checklist.

1. Retain owner-private destination coordinates for a unit or its exact army
   membership. Advance an affordable prefix immediately and once per subsequent
   round after movement refresh, in deterministic order. Use a bounded geometric
   route (no omniscient terrain search), validate every step with existing movement
   rules, stop on generic obstruction/ZOC, never auto-attack. Cancel/replace
   explicitly; manual orders, displacement, membership/ownership changes cancel.
   Preserve atomic legacy move semantics and all turn-policy authority.
2. Introduce versioned ruin placement without changing old generator output.
   Persist unrevealed per-ruin rolls; exploring by movement consumes a ruin once.
   Implement the released villages-only risk profile (gold, advance, mercenary,
   settler, city, empty); do not claim barbarian-profile support. Apply real pinned
   content/effects and private, stable reward notifications. Project ruin presence
   with current sight and stored fog observations, never hidden rolls/rewards.
3. Add persisted city governor enablement and priority. Use pinned retail building
   and wonder lists for growth/production/gold/science/happiness/defense/offense;
   fill empty queues using existing production validation, never overwrite manual
   queues, rush-spend gold, or alter national policy. Run on enable and each city
   round, reset on ownership transfer. Advertise only supported priorities.
4. Wire command routes, phase-aware legal actions/intents, capability flags and
   owner-private state through existing transactional receipts/snapshots. Document
   precise scope rather than claiming complete original AI/slider optimization.
5. Review against CTP2 ArmyData::ExecuteMoveOrder/Move, GoodyHuts::ChooseType/
   OpenGoody, wldgen::GenerateGoodyHuts, Governor::FillEmptyBuildQueues and retail
   risk/build-list files. Test restart equality, invalid/foreign commands,
   cancellation, fog isolation, idempotent service replay, turn modes and terminal
   boundaries. Run full tests, vet, build, freeze guard and diff checks.
