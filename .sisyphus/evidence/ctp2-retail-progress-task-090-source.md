# Task 090 source notes — advance UI/advice metadata classification and Age split

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/AdvanceRecord.cpp:306-312` parses `Branch` and `Age` as database references; `AdvanceRecord.cpp:385-388` parses `KeyAdvanceAdvice` as an advice string ID; `AdvanceRecord.h:104-108,122,144-147` exposes icon/branch/age/GLHidden/advice accessors.
- `ctp2_source_code/ctp2_code/ui/interface/sciencewin.cpp:839-842` uses `Branch` for science-window branch grouping/imagery.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrarywindow.cpp:413-421` uses advance `Icon` metadata for Great Library still/movie/gameplay/history/prereq/variable files.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrary.cpp:2172-2178` uses `GLHidden` to hide database entries from Great Library topic lists.
- `ctp2_source_code/ctp2_code/ai/diplomacy/diplomat.cpp:4104-4113` and `4159-4185` use `KeyAdvanceAdvice` to choose/populate diplomacy/science-advice messages.
- `ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:513-557`, `10046-10057`, `CityData.cpp:562-572`, `Advances.cpp:493`, and `gameinit.cpp:1659-1660` show `Age` is gameplay/setup-affecting: starting age grants/skips advances/buildings, ending age can constrain research, and age transitions depend on researched advances.

Classification/split:

- `Branch`, `Icon`, and `GLHidden` are presentation/setup metadata for server gameplay.
- `KeyAdvanceAdvice` is diplomacy/AI advice text metadata, not a deterministic gameplay rule mutation.
- `Age` is not classified away; it remains a gameplay/setup field that must be implemented/verified separately for start/end-age and age-transition parity.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
