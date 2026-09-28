# Task 018 source check: unit upgrade legal actions

Retail data/source evidence:

- Retail unit records carry `UpgradeTo` chains. Example: `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt` has upgrade chains such as archer-era units upgrading into later mounted/armored units.
- The released source parser stores this field on unit records (`ctp2/ctp2_source_code/ctp2_code/gs/newdb/unit.cdb` defines unit database fields, and the server catalog inventory already snapshots `UpgradeTo` as a behavior-bearing unit field).
- The server runtime already projects `UpgradeTo`, engine execution already validates `upgrade-unit`, and visibility already exposes legal `upgrade-unit` actions only when prerequisite/government gates are met.

Implementation notes:

- AI planning now consumes projected legal `upgrade-unit` actions instead of ignoring unit upgrade chains that visibility has already proven legal.
- The planner validates the public payload (`unit_id`, `upgrade_to`) and submits the same public action without hidden-state access.
