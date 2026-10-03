# Task 074 source classification: `Explodes` is combat presentation only

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:679` and `:1786-1787` parse the `Explodes` unit flag into the unit record.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CTP2Combat.cpp:1079-1086` and `:1181-1188` add `BATTLE_EVENT_TYPE_EXPLODE` battle events when a damaged defender survives ranged/melee hits; these blocks do not modify HP, units, terrain, cities, diplomacy, or other gameplay state.
- `ctp2/ctp2_source_code/ctp2_code/ui/interface/battle.cpp:379-380` gates the explosion animation on `theUnit.GetDBRec()->GetExplodes()` and returns early for non-exploding unit records.
- Searches for `GetExplodes`/`Explodes` in `ctp2_code/gs` found parser/getter references and combat battle-event/UI animation plumbing, with no gameplay state mutation tied to the flag.

Classification:

- `Explodes` is safely classified as presentation-only combat animation metadata for server gameplay parity.
- The server continues importing/projecting the marker so clients can animate it if needed, but no additional gameplay execution is required.

Validation evidence:

- Source grep/inspection: `.sisyphus/evidence/ctp2-retail-progress-task-074-grep.txt` and this file.
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-074-diff-check.txt`
