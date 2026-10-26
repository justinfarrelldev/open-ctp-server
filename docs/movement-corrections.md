# Retail movement corrections

New retail rule pins advertise `movement.retail_movement_version: 1`.

- Ordinary foreign terrain is enterable regardless of territorial ownership.
  Enemy occupants, domain restrictions, full stacks and foreign cities remain
  separate gates. Peaceful foreign cities cannot be entered; wartime capture
  additionally respects every member's `CantCaptureCity` capability.
- Owner-private `unit.first_move` is an explicit nullable boolean, not a deduction
  from `movement == max_movement`. Creation and real turn refresh set it; movement,
  upgrades, launch, bombardment and movement-consuming special actions clear it.
  Each army member admits/spends its own cost. An expensive first step saturates
  movement at zero; later unaffordable steps reject atomically or wait in a
  retained order. Fuel remains an independent execution restriction.
- Aircraft pay the released fixed 100-point step cost. Amphibious units can enter
  land or sea terrain. Original stationary city-unit records retain zero movement.

## Routing authority

The released client calls `UnitAstar::FindPath` in
`ctp2_source_code/ctp2_code/robot/pathing/unitastar.cpp`, backed by `Astar` in
`robot/pathing/astar.cpp`. The server now uses **A\***, not greedy stepping or
Dijkstra, for retained routes and corrected-pin projected current-turn planning:

- `ComputeValidMovCost` caps terrain routing cost at the slowest member's original
  maximum allowance; this is not the actual cost deducted during execution.
- All-air stacks use fixed 100-point routing cost. Human `pretty_path` preferences
  reproduce `DecayOrtho`'s 5% directional discount; AI routes omit it.
- The priority queue uses `f = g + h`, with deterministic cost/ID tie breaking.
  Costs use integer hundredths rather than floating point. Improved paths reopen
  nodes rather than preserving the original disabled cost-relaxation behavior.
- The distance heuristic uses the server's existing bounded eight-neighbor grid
  and a minimum cost from currently visible cells (90 points for all-air stacks,
  bounded by the actual lower bound). It does not consult the released engine's
  hidden-world nearby-cost cache. No new wrapping/topology semantics are implied.
- Server routing and projection share sight/detection rules. Routes use current
  visible terrain and detected occupants only. Unseen destinations progress via
  reachable visible frontiers and replan as sight changes. Remembered terrain is
  not replaced with hidden live facts. No route or ETA is exposed by the retained
  destination contract. Occupants/ZOC discovered during actual execution can still
  stop an order; retained movement never automatically attacks.

## Compatibility and delivery

Existing rule pins are immutable. Version-zero pins retain legacy ownership and
strict-budget policies (and legacy projected route selection); the supported way to obtain the corrected retail policy
is to start a new game with a newly imported retail pin. Do not edit a live pin,
rewrite its hash, or infer first-move history from a full movement budget.

The state field is additive: missing `first_move` stays unknown and grants no
special allowance. A real refresh under a version-one pin initializes it. Ordinary
snapshot persistence, canonical hashes and transactional command replay include
known true/false values. Newly executed retained orders use the updated routing
code, including when their pin retains legacy access/budget rules.

Regression coverage includes one subtest for **each of all 74 released unit
records**, checked against original `Units.txt`/`Terrain.txt` data (72 map-unit
records including three stationary cities, plus the original trade-point behavior
for caravan and freight transport), plus access,
cities, cargo, mixed armies, first-move/fuel isolation, restart equality, detours,
changing obstacles, frontier exploration, hidden-unit/terrain isolation, projected
planning and AI exploration. This is not a claim of complete transport, treaty or
fuel-system parity beyond the movement rules described above.

These changes do not deploy/restart the live server, mutate game95, or modify the
client. Live-game/client acceptance must be performed after deploying the server
and starting a game with the corrected retail pin.
