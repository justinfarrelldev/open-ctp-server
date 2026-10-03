# Task 081 source classification: building `AllowGrunts`

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/BuildingRecord.cpp:242` and `:630-631` parse the `AllowGrunts` building flag.
- `ctp2/ctp2_source_code/ctp2_code/gs/outcom/C3BlgDB.cpp:371` exposes `BLG_EFFECT_ALLOW_GRUNTS` to the AI/outcom building database.
- `ctp2/ctp2_source_code/ctp2_code/robotcom/Agent/CityAgentPop.cpp:963-970` computes AI grunt utility by asking `m_population->GetCityGruntOutput(...)`.
- `ctp2/ctp2_source_code/ctp2_code/gs/outcom/C3Population.cpp:533-536` implements `GetCityGruntOutput` as `return 0;`, so the parsed flag has no gameplay output in this source release.
- Retail default data uses `AllowGrunts` only on `IMPROVE_FACTORY` (`buildings.txt:364`).

Classification:

- `AllowGrunts` is safely classified as non-executed/AI-placeholder metadata for literal retail gameplay. It remains imported/projected as an effect marker, but no city yield/specialist execution is required unless a future optional mod ruleset supplies a nonzero grunt-output implementation.

Validation evidence:

- Source/data grep: `.sisyphus/evidence/ctp2-retail-progress-task-081-grep.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-081-diff-check.txt`
