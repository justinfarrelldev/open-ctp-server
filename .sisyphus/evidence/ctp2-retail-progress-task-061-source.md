# Task 061 source check: city-capture wonder transfer

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/UnitData.cpp:2011-2023` in `UnitData::ResetCityOwner` iterates the captured city's built-wonder bitset, calls `g_player[m_owner]->RemoveWonder(w, FALSE)`, and then calls `g_player[newo]->AddWonder(w, me)` for each wonder in the captured city.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7435-7658` shows `Player::AddWonder` re-applies owner-facing wonder add effects such as polluters-to-parks, free slaves, HP bonuses, full happiness, nuke elimination, global radar, illegal-item cleanup, and building-everywhere bits.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:7661-7690` shows `Player::RemoveWonder` clearing old owner wonder state including building-everywhere bits/global radar.

Implementation notes:

- `captureCityAt` now transfers non-destroyed wonders whose `CityID` is the captured city to the conqueror.
- After transfer, `applyWonderCompletionEffects` runs for the new owner, matching retail's `AddWonder` call on capture.
- Owner-scoped effects such as building-everywhere prerequisites now immediately apply to the conqueror and stop applying to the defeated owner via the changed owner ID.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-061-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-061-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-061-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-061-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-061-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-061-diff-check.txt`
