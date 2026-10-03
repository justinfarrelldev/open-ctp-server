# Task 091 source notes — government rank/advice/icon/IsEcotopia metadata classification

Retail source evidence:

- `ctp2_source_code/ctp2_code/gs/newdb/GovernmentRecord.cpp:980-1029` parses government rank display fields (`GrowthRank`, `ProductionRank`, `ScienceRank`, `GoldRank`, `MilitaryRank`, `CommerceRank`, `PollutionRank`, `LoyaltyRank`, `MartialLawRank`).
- `ctp2_source_code/ctp2_code/gs/newdb/GovernmentRecord.cpp:1100-1119` parses `IsEcotopia`, `SameRankAdvice`, and `HigherRankAdvice`; `GovernmentRecord.h:176-202` exposes icon/rank/advice/IsEcotopia accessors.
- `ctp2_source_code/ctp2_code/ui/interface/GovernmentTab.cpp:304-333` and `428-474` use the rank fields to populate government comparison arrows and descriptive UI strings.
- `ctp2_source_code/ctp2_code/ai/CityManagement/governor.cpp:266-279` uses `Rank`, `SameRankAdvice`, and `HigherRankAdvice` to choose government-change advice text.
- `ctp2_source_code/ctp2_code/ui/interface/greatlibrarywindow.cpp:435,511` uses government `Icon` metadata for Great Library presentation.
- `GetIsEcotopia` has no non-parser gameplay call sites in the released source tree; it is a database marker with no executed state mutation in retail source.

Classification:

- Government `Icon`, rank display fields, `Rank`, `SameRankAdvice`, and `HigherRankAdvice` are UI/advice metadata for server gameplay.
- Government `IsEcotopia` is source-classified as non-executed marker metadata for server gameplay.

Validation:

- `go test ./internal/gameplay/rules -run TestRetailFieldInventorySnapshot -count=1`
