# Task 079 source classification: government `ProfessionalUnits`

Retail source/data evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/GovernmentRecord.cpp:280` and `:887-888` parse `ProfessionalUnits` into government records.
- `ctp2/ctp2_source_code/ctp2_code/robotcom/Alloc/wotp.cpp:316-320` uses `GetProfessionalUnits(gov)` only in AI government/readiness utility scoring.
- `ctp2/ctp2_source_code/ctp2_code/robotcom/Agent/scienceagent.cpp:787-790` also uses `GetProfessionalUnits(gov)` only for AI science/government utility.
- Runtime support/disband behavior uses the per-unit `IsProfessional` flag: `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Readiness.cpp:129-130` makes `u.GetIsProfessional()` support-free, and `:252-256` separates already-professional units during over-budget disband checks.
- Retail default `govern.txt` sets `ProfessionalUnits 0` for all 12 governments.

Classification:

- Government `ProfessionalUnits` has no literal retail gameplay execution in default data beyond AI evaluation and all default values are zero.
- The server already executes the gameplay-affecting per-unit `IsProfessional` support exemption. Government `ProfessionalUnits` remains projected metadata/AI-scoring input until broader AI parity requires it.

Validation evidence:

- Source/data grep and counts: `.sisyphus/evidence/ctp2-retail-progress-task-079-grep.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-079-diff-check.txt`
