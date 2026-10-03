# Task 049: retail EliminateNukes globally blocks nuclear-attack unit production

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp:4177-4181`
  - City unit production rejects records with `NuclearAttack` when `wonderutil_GetNukesEliminated(g_theWonderTracker->GetBuiltWonders())` is true.
  - The check uses the global built-wonder mask.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10203-10205`
  - Player production validation applies the same global nuke-elimination gate.

## Implementation notes

- Nuclear-attack unit production now checks active global `EffectEliminateNukes` wonders directly, in addition to canonical nation nuke-elimination markers applied on wonder completion.
- This keeps initial/imported canonical states with an existing active nuke-defuser wonder from allowing new nuclear-attack units before a completion event reruns.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-049-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine -run 'TestRetailEliminateNukesWonderGloballyBlocksNuclearAttack|TestRuntimeEliminateNukesBlocksNuclearAttackProduction' -count=1`
