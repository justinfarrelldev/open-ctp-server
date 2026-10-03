# Task 041: retail wonder IncreaseRegard diplomacy projection

## Retail source/data checked

- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/diplomat.cpp:1196-1221`
  - `Diplomat::RecomputeRegard` gets `wonderutil_GetIncreaseRegard(g_player[m_playerId]->GetBuiltWonders())`; for each contacted foreigner it logs a `REGARD_EVENT_WONDER` on the foreigner's diplomat toward the wonder owner for duration `1`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/wonderutil.cpp:295-297`
  - `wonderutil_GetIncreaseRegard` sums active `IncreaseRegard` wonder values.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Wonder.txt:92-98,328-337`
  - Released wonders include `IncreaseRegard 100` on `WONDER_EMPIRE_STATE_BUILDING` and `IncreaseRegard 50` on `WONDER_THE_FORBIDDEN_CITY`.

## Implementation notes

- Public diplomacy projection now treats an active `EffectIncreaseRegard` wonder on the relation target as positive regard from the viewer/foreigner toward that wonder owner.
- Direction remains retail-accurate: the owner's own regard toward the foreigner is unchanged by the owner's wonder.
- Existing war/union precedence remains intact (`hotwar` and `love` still override friendly wonder regard).

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-041-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/visibility -run 'TestRetailIncreaseRegardWonder|TestThrowPartyRegardProjectionIsDirectional' -count=1`
