# Task 054: retail DeathEffectsHappy gates combat peace-movement losses

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:987-989`
  - `Unit::DeathEffectsHappy()` returns the unit DB `DeathEffectsHappy` flag.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityEvent.cpp:337-352`
  - Nuke city kill lists call `RegisterLostUnits` only for units where `DeathEffectsHappy()` is true.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp:5488-5499`
  - Nuke kill lists apply the same `DeathEffectsHappy()` gate before registering lost units for morale/peace movement.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Unit.cpp:120-135`
  - Normal combat kill causes trigger unit death scripting, but do not unconditionally register peace-movement lost-unit counters.

## Implementation notes

- Combat casualty removal now records peace-movement unit losses only for units carrying `EffectDeathEffectsHappy`.
- Direct helper behavior and existing nuke blast gating remain unchanged.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-054-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailCombatDeath.*DeathEffectsHappy|TestCombatLossesClassifyPeaceMovementHomeAndOverseas|TestAdvancePeaceMovementAppliesDefeatDecayAndResetsLossCounters' -count=1`
