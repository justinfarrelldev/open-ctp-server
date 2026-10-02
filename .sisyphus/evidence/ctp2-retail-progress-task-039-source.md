# Task 039 source check: direct nuclear-attack units

Retail CTP2 source/data points:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/ArmyData.cpp` checks `GetNuclearAttack()` for units that can nuke cities/locations.
- Retail nuclear city targeting rejects own cities, checks `city.SafeFromNukes()`, consumes the nuclear unit even when the target city is protected, and queues `GEV_NukeCityUnit` when the target is vulnerable.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/unitevent.cpp` `NukeCityUnitEvent` invokes the city nuke event and kills the nuclear unit with `CAUSE_REMOVE_ARMY_NUKE`.

Server alignment:

- Units with the imported `nuclear_attack` effect can execute canonical `nuclear-attack` against a foreign city tile.
- Protected target cities (`protect_from_nukes`/force-field effects) consume the missile without city damage.
- Vulnerable cities take the same population/pollution/blast-unit/improvement effects used by server nuke resolution, and the attacking nuclear unit is consumed.
- Safe projection exposes `nuclear-attack` only for visible foreign city targets, and the authenticated HTTP route forwards only `unit_id`/`target_tile_id` to the canonical command.
