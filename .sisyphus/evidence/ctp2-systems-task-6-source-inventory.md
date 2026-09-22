# Task 6 CTP2 source inventory: advances, research, and unlocks

Source root inspected before implementation and rechecked after implementation: `/home/justi/Projects/open-ctp-server/ctp2/ctp2_source_code`.

## Advance schema
- `ctp2_code/gs/newdb/advance.cdb` defines `Advance` with `Int Cost`, `Record Advance[0..4] Prerequisites`, `Dependencies`, `LeadsTo`, `Branch`, and `Age`.
- This supports the server synthetic contract implemented here: bounded advance IDs with positive costs and prerequisite advance IDs, while explicitly not copying the retail tree.

## Research eligibility and prerequisite reset
- `ctp2_code/gs/gameobj/Advances.cpp` `Advances::ResetCanResearch` clears already-owned advances, checks every `AdvanceRecord` prerequisite against `m_hasAdvance`, and marks researchable advances in `m_canResearch`.
- The same function also keeps a limited offered-choice set; the server intentionally uses deterministic sorted availability instead of CTP2's random offer trimming.

## Research selection and progress
- `ctp2_code/gs/gameobj/Player.cpp` `Player::SetResearching` obtains `m_advances->CanResearch()` and refuses a human/network selection when the advance is not researchable.
- `ctp2_code/gs/gameobj/Player.cpp` science update logic adds science, compares `m_science->GetLevel()` against `m_advances->GetCost()`, subtracts the cost, and grants the researched advance.
- Server alignment: `set-research` requires a runtime-defined advance whose prerequisites are satisfied, and round completion adds deterministic research points until the target cost completes.

## Granting advances
- `ctp2_code/gs/gameobj/Advances.cpp` `Advances::GiveAdvance` returns if the advance is already owned, calls mod validation, sets `m_hasAdvance`, increments discovery counters, and notifies networking.
- Server alignment: nation `Advances` remains a sorted owned-ID set; completion does not duplicate already-owned advances and proceeds through deterministic next available advances.

## Unlock/obsolete semantics
- `ctp2_code/gs/newdb/UnitRecord.cpp` has `EnableAdvance` and `ObsoleteAdvance` tokens and resolves them to advance IDs.
- `ctp2_code/gs/outcom/C3UnitDB.cpp` `C3UnitDB::CityCanBuildUnit` denies obsolete units when the player has an obsolete advance and requires the enable advance when present.
- `ctp2_code/gs/newdb/WonderRecord.cpp` has the same `EnableAdvance`/`ObsoleteAdvance` shape for wonders.
- Server alignment in this task: production prerequisites/obsolete IDs now recognize owned nation advances, and runtime unit prerequisites are enforced even if a client omits the prerequisite from the public command payload.

## Hidden-information boundary
- CTP2 research choice is per-player through `Player::SetResearching`; the server preserves this by projecting only the viewer nation's owned advances, research target/progress, and available research IDs.
- Existing server steal-technology continues to choose deterministically from target-owned advances without disclosing the list in event payloads.
