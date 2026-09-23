# CTP2 Full Retail Gameplay Parity Plan

## Purpose
Bring the gameplay-systems branch from the current retail catalog/runtime foundation to full released-game CTP2 multiplayer parity for server-authoritative gameplay. This means the server must be able to load user-owned released-game data, start a game with the retail ruleset, let AI-controlled nations play complete games, and execute every retail gameplay mechanic needed for a complete game cycle without synthetic-only assumptions.

## Current Stable Baseline
- Branch: `gameplay-systems-completion`.
- Current work is committed and stable at `dd3b2be feat(ai): value retail government economy bonuses`.
- Existing retail support includes raw/typed catalog parsing, runtime projection/pinning, retail startup opt-in, basic production cost/effects, terrain-improvement restrictions, unit fuel/ranged projection, government coefficient projection, and AI use of projected retail-like rules.
- This is **not** full retail parity yet. Many behavior-bearing fields are preserved but not executed.

## Non-negotiable Methodology
- Use strict TDD for every behavior change: RED test, GREEN implementation, REFACTOR/cleanup, then validation.
- Commit atomically after every stable behavior slice.
- Check original CTP2 source/data before and after ambiguous retail mechanics.
- Do not commit proprietary retail data; tests may use local user-owned retail data and skip when absent.
- Keep viewer projections and AI inputs client-safe: no hidden canonical state, no raw catalog leakage, no omniscient targeting.
- Keep `.sisyphus/evidence/*` source and validation notes for every full-retail task.
- Update `.sisyphus/plans/checklists/ctp2-full-retail-parity.md` continuously.

## Definition of Full Retail Parity
Full retail parity is reached only when all of the following are true:
1. The server can import/pin the released-game default gamedata catalogs and compile a runtime ruleset without synthetic record requirements.
2. All 74 retail units, 107 advances, 30 wonders, 52 city improvements, 40 terrain improvements, 26 terrains, 12 governments, pollution trigger lists, and major AI build/strategy lists are parsed and behaviorally represented.
3. The engine executes all retail fields that affect authoritative gameplay: movement, fuel, visibility, stealth/detection, domains, transport/cargo, bombard/ranged combat, special attacks/actions, upgrades, obsolescence, production, upkeep, governments, science/economy, happiness, crime, pollution, disasters, diplomacy/trade, wonders/endgame, terrain/public works, and victory.
4. Public projection/API/SSE surfaces expose only viewer-safe data while providing enough legal-action metadata for a client and AI to play.
5. AI can play a full retail-runtime game using only public projections/history.
6. A deterministic AI-vs-AI full-game E2E records every turn to an artifact and reaches at least turn 100; a longer game either reaches a valid retail victory/end condition or a documented retail final-turn score condition.
7. Full repository validations pass: freeze guard, all unit/e2e tests, `git diff --check`, and long-running retail AI scenario.

## Major Workstreams

### A. Retail Data Coverage and Runtime Model
- A1. Complete typed catalog fields for `unit.cdb`: flags, movement types, carry/attack/see classes, fuel, stealth, special attacks, upgrade paths, martial law, city founding/build-city flags, terraforming, spy actions, trader flags, active defense, bombard, space/domain fields.
- A2. Complete typed catalog fields for `building.cdb`: every scalar/flag/nested effect including happiness, crime, pollution, production, commerce, science, specialist, defense, refuel, conversion/slavery protection, capitol, airport/port, population, food, and prerequisites.
- A3. Complete typed catalog fields for `wonder.cdb`: every world/city/nation effect, obsolescence, Gaia/endgame interactions, free buildings, pollution/happiness/science/trade/production effects, diplomatic effects.
- A4. Complete typed catalog fields for `terrain.cdb`: yields, movement, defense, transforms, resources/goods hooks if relevant, tile ecology values.
- A5. Complete typed catalog fields for `government.cdb`: sliders/rates, workday/rations/wages, readiness, martial law, war discontent, support, crime, coefficients, pollution, city-count penalties, ranks.
- A6. Parse AI lists: `UnitBuildLists.txt`, `BuildingBuildLists.txt`, `WonderBuildLists.txt`, `AdvanceLists.txt`, `ImprovementLists.txt`, strategy/personality files, diplomacy goal files as needed.
- A7. Add runtime validation for all cross references and retail quirks (including intentionally hidden/self-prereq advances) without false rejection.
- A8. Add compatibility metadata showing which fields are executed vs parsed-only; full parity requires no gameplay-affecting parsed-only fields remain.

### B. Advances, Research, Obsolescence, and Unlocks
- B1. Implement complete advance availability semantics from `Advances.cpp`, including hidden/non-researchable entries and prerequisites.
- B2. Ensure every unlock path (units, buildings, wonders, governments, terrain improvements, special actions, tile transforms) uses runtime retail definitions.
- B3. Implement advance completion side effects, notifications/events, and deterministic auto-selection of next research for AI/governor.
- B4. Implement obsolescence for units/buildings/wonders, including effects disabling and production legality changes.
- B5. Add tests for complete retail tech tree reachability and representative historical paths through late-game advances.

### C. Units, Armies, Movement, Visibility, and Combat
- C1. Complete unit production/spawning, continuous queues, population-removing units, only-build-one, city-cell capacity, trader exceptions.
- C2. Implement domains and movement classes: land, mountain, sea, shallow water, air, space, amphibious, tunnels/roads/rails/maglev, city/transport constraints.
- C3. Implement fuel/refuel faithfully: fuel costs, crash/no-crash behavior, city/building/carrier refuel, turn refresh semantics.
- C4. Implement visibility classes, stealth, can-see, detectors, hidden units, fog/memory, and viewer-safe projections.
- C5. Implement cargo classes/sizes, transport capacity, carriers, unload/load rules, ocean/space cargo distinctions.
- C6. Implement full army stack limits, formation/line selection, active defense, ranged phase, bombardment, bombard rounds/range, city bombard, counterfire.
- C7. Implement domain-specific attack/defense flags (`CanAttack`, `Defend`, `AttackCity`, etc.) and fail-closed illegal combat.
- C8. Implement special unit actions: settle, disband, fortify, pillage, enslave, convert, sue/franchise, steal technology, plant nuke, bio/nano attacks, incite revolution, investigate, establish embassy, underground railroad, soothsaying, terraforming, space launch/build hooks.
- C9. Implement upgrade paths and veteran/health/repair/refuel mechanics.
- C10. Add unit-level public projection/action tests and AI legal-use tests for every action family.

### D. Cities, Economy, Buildings, Wonders, Happiness, Crime
- D1. Complete city growth, food storage, starvation, specialists, population caps, overcrowding, city tile working, workday/rations/wages.
- D2. Implement production pipeline parity: shield stores, rush buying, queue replacement/cancellation, capitalization/infrastructure, continuous units, wonder races/losses.
- D3. Implement every building effect from retail `buildings.txt`/`building.cdb` with CTP2 source citations.
- D4. Implement every wonder effect from retail `Wonder.txt`/`wonder.cdb`: uniqueness, global/national/city scopes, obsolescence, free buildings, diplomacy/world/Gaia effects.
- D5. Implement happiness and unrest: religious/entertainment/building/wonder effects, war discontent, martial law, crime, pollution unhappiness, conquest distress.
- D6. Implement city defense bonuses, walls/forcefields, anti-nuke/bio/nano protections, refuel buildings, airports/ports/coastal requirements.
- D7. Ensure city projections expose viewer-safe but client-complete city state.

### E. Terrain, Public Works, Tile Improvements, Goods/Trade Routes
- E1. Execute all terrain yields/movement/defense directly from retail terrain runtime instead of fallback buckets.
- E2. Implement terrain improvements with all terrain-specific effects, levels/classes/exclusions, construction costs/times, roads/rail/maglev/tunnels, ocean improvements, detectors, farms/mines/ATM/structures.
- E3. Implement transform/terraform actions and advance restrictions.
- E4. Implement pillage/removal and tile ownership interactions.
- E5. Implement goods/trade route creation, route pathing, piracy/blockade where present, and trade yields/agreements.
- E6. Public projection must expose only visible tile improvements/pollution/goods.

### F. Governments, Diplomacy, Trade, Readiness
- F1. Complete all government coefficients and sliders: science/tax/workday/rations/wages, production/food/gold/science, pollution, support, crime, city-count penalties.
- F2. Implement readiness states, turns-to-readiness, HP/readiness modifiers, war/peace discontent.
- F3. Implement diplomacy relationship states: contact, treaty types, ceasefire/peace/war/alliance/trade/embargo, trust/attitude fields used by AI.
- F4. Implement proposals/counterproposals, acceptance/rejection, treaty expiry/breaking, reputation penalties.
- F5. Implement government-specific diplomacy constraints and wonder effects on diplomacy.
- F6. Add public diplomatic projection and AI legal-use tests for every treaty action.

### G. Pollution, Global Warming, and Disasters
- G1. Wire retail pollution model into authoritative turn processing rather than standalone rule tests.
- G2. Compute city and world pollution from population, production, buildings, wonders, terrain improvements, governments, and active suppressors.
- G3. Maintain 5-turn pollution history and strict `>` trigger semantics.
- G4. Execute disasters from retail trigger lists: floods, ozone depletion, nuclear winter/global warming and all source-verified disaster outcomes.
- G5. Implement tile/city/unit effects of disasters: terrain changes, city damage/pop loss, building destruction, unit damage/death, pollution changes.
- G6. Implement pollution reducers/suppressors from wonders/buildings/governments and terrain cleanup.
- G7. Add public world pollution endpoints/projections and AI cleanup behavior.
- G8. Add deterministic disaster tests for each map-size trigger list and at least one full-game disaster occurrence.

### H. Victory, Endgame, Space/Gaia, and Full Game Cycle
- H1. Complete conquest, score/final-turn, diplomatic union, Gaia/controller and any retail endgame object mechanics.
- H2. Implement Gaia countdown prerequisites and cancellation from real retail records.
- H3. Implement space/endgame object construction and victory triggers.
- H4. Ensure terminal state, scoring, rankings, and event output match source-backed semantics.

### I. Retail AI Parity
- I1. Parse and project retail AI build lists/strategy lists.
- I2. Replace heuristic-only AI with strategy/list-driven choices for research, governments, units, buildings, wonders, terrain improvements, diplomacy, war/peace, pollution cleanup, and endgame.
- I3. Ensure AI uses only public projection and legal actions.
- I4. Add deterministic AI-vs-AI tests for early, mid, late, and endgame situations.
- I5. Add long-run AI-vs-AI E2E: minimum 100 turns, output every turn, assert no hidden-state use, no invalid commands, and meaningful progress.
- I6. Add full completion E2E that runs until retail victory/final-turn if runtime permits within test budget; otherwise make it a nightly/long tag command with artifact output.

### J. Public API, SSE, Persistence, and Client Readiness
- J1. Extend HTTP command schemas/projections for all new actions.
- J2. Extend SSE event redaction/projection semantics for combat, diplomacy, disasters, pollution, wonders, research, city changes.
- J3. Ensure restart/canonical-state equality for every mechanic family.
- J4. Store/persist runtime pins and retail direct runtime definitions safely.
- J5. Ensure diagnostics expose rules identity/counts but not hidden gameplay state.

### K. Verification, Evidence, and Review
- K1. Every mechanic has source inventory evidence mapping CTP2 source/data to implementation.
- K2. Every mechanic has happy path/failure/restart/projection/AI tests.
- K3. Maintain engine freeze lock for production `.go` changes.
- K4. Add retail field coverage report: all behavior-bearing retail fields either executed or explicitly proven presentation-only.
- K5. Add final comprehensive code review evidence.
- K6. Add final fullgame artifact bundle with at least 100-turn AI-vs-AI JSONL transcript and visual summary.

## Execution Order
1. Planning/checklist commit.
2. Full retail field coverage inventory and gap report.
3. Add deterministic AI-vs-AI 100-turn harness and transcript artifact using current partial retail runtime; this becomes the long-run regression target.
4. Replace terrain yields/movement with retail runtime values.
5. Complete pollution/disaster integration.
6. Complete diplomacy/government/readiness/trade.
7. Complete unit special actions/visibility/stealth/cargo/combat.
8. Complete buildings/wonders/happiness/crime/economy.
9. Complete AI list parsing and strategy execution.
10. Complete victory/endgame/Gaia/full cycle.
11. Final retail field coverage closure and full-game completion run.

## Required Long-Run Fullgame Artifact
- Test/harness name: `TestRetailAIAutoPlayRecordsAtLeastHundredTurns` (or successor).
- Players: all AI if possible; otherwise deterministic mock controller for player 1 using only public projection.
- Minimum assertions:
  - reaches turn >= 100;
  - no invalid commands;
  - no hidden-state AI input;
  - records one JSONL summary per turn with turn, active player, command counts, city/unit/advance/wonder counts, pollution, diplomacy summary, and state hash;
  - transcript path is printed by the test and can be written to `.sisyphus/evidence/` by an explicit recording command.
- Final parity assertion additionally requires complete-game terminal victory/final-turn artifact.

## Current Known Risks
- Full retail parity is a large multi-system implementation, not a single small patch.
- Retail data includes quirks (hidden/self-prerequisite advances) that must be preserved without corrupting runtime validation.
- Long AI games can be slow/flaky unless deterministic policy replay and bounded command budgets are maintained.
- Some retail mechanics may depend on UI/SLIC/presentation assets; only authoritative gameplay effects belong on the server.
