# Task 019 source check: remaining projected retail special actions

Retail data/source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/unit.cdb` defines `CanSoothsay` and `CreateParks` as unit special-action fields.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/UnitRecord.h` exposes `GetCanSoothsay()` and `GetCreateParks()`.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/order.cdb` defines `UnitPretest_CanSoothsay`, tying soothsaying to an order pretest.
- `ctp2/ctp2_source_code/ctp2_code/gs/newdb/terrimprove.cdb` defines the `Terraform` terrain-improvement class and `TerraformTerrain` target field.
- `ctp2/ctp2_source_code/ctp2_code/ai/CityManagement/governor.cpp` evaluates terraform improvements with `terrainutil_GetTerraform*Improvement()` and `CanCreateImprovement`.

Implementation notes:

- AI planning now consumes the remaining projected city/tile special legal actions (`soothsaying`, `create-park`, `terraform`) rather than ignoring them.
- The planner validates only public payload fields (`unit_id`, `city_id`, `tile_id`) and reuses the projected payload for submission.
