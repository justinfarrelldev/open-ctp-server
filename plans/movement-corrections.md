# October 25 movement corrections

Supersedes the geometric-routing constraint in retained-movement-ruins-governors.md.
Use red → green → refactor, an updated checklist and atomic commits with the normal hook.

1. Version the retail movement policy: ordinary foreign land is enterable independently of foreign city, occupant and stack restrictions. Share access with viewer planning. Old immutable pins retain their old policy; starting a new game is the supported upgrade, not editing live pins.
2. Persist explicit owner-private per-unit first-move state. New units and turn refresh initialize it; movement/action consumption clears it. Share per-member admission/spending, saturate movement at zero, preserve independent fuel/phase/domain/ZOC safeguards. Unknown legacy state grants no allowance until refresh. Update planning/AI and restart/hash coverage.
3. Replace retained greedy stepping with the released UnitAstar A* routing algorithm, including capped terrain costs, fixed air costs and human pretty-path preference. Adapt the distance lower bound to the bounded server grid and viewer-visible costs rather than a hidden-world cache. Replan after steps/turns and changing obstacles, explore toward unseen destinations via visible frontiers, never disclose hidden routes or auto-attack. Preserve exact stack identities and distinguish waiting from illegal movement.
4. Review released CellUnitList::CanMoveIntoCell/IsMovePointsEnough, UnitData::ResetMovement/DeductMoveCost, and SelItem pathfinding. Run focused and full tests, vet, build, freeze checks and diff checks; document compatibility and delivery boundaries.
