# Task 085 source notes — multi-government unit production gates

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.cpp:1695-1700` parses `GovernmentType` with `ParseRecordInArray`, so a unit may list multiple allowed governments.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:10183-10196` rejects production only if the unit has one or more `GovernmentType` entries and none of them matches `m_government_type`; it loops all `GetNumGovernmentType()` entries.

Implementation notes:

- `RuntimeUnitDefinition` now carries `RequiredGovernmentIDs` in addition to the legacy first `RequiredGovernmentID` for compatibility.
- Retail runtime projection preserves every repeated `GovernmentType` value while still setting the legacy first value.
- Engine production, visibility legal actions, and unit upgrade government checks accept any listed government and reject unlisted governments.

Validation:

- RED evidence: `.sisyphus/evidence/ctp2-retail-progress-task-085-red.txt`.
- GREEN targeted evidence: `.sisyphus/evidence/ctp2-retail-progress-task-085-green-targeted.txt`.
- GREEN package evidence: `.sisyphus/evidence/ctp2-retail-progress-task-085-green-packages.txt`.
- Freeze evidence: `.sisyphus/evidence/ctp2-retail-progress-task-085-freeze.txt`.
- Full suite evidence: `.sisyphus/evidence/ctp2-retail-progress-task-085-green-full.txt`.
