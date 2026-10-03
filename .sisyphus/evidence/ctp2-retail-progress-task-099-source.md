# Task 099 source notes — remaining gameplay-bearing advance flags verified

Retail source evidence:

- `AdvanceRecord.cpp:181-197` and `AdvanceRecord.h:110-131` parse/expose `Capitalization`, `Infrastructure`, `PlantNuke`, `PollutionSizeModifier`, and `PollutionProductionModifier`.
- `Advances.cpp:247-253` enables capitalization/infrastructure production flags when the advance is gained.
- `CityData.cpp:2113-2117` converts stored production to gold/public-works materials using government `CapitalizationCoefficient` and `InfrastructureCoefficient`.
- `advanceutil.cpp:14-18`, `UnitData.cpp:5408-5410`, and `ArmyData.cpp:8294` use `PlantNuke`/plant-nuke data to authorize plant-nuke actions.
- `Advances.cpp:752-787` sums known advances' pollution size/production modifiers; `Player.cpp:7793-7812` exposes those sums to city/pollution code.
- `Player.cpp:9975-9983` checks `AdvanceRecord::GetTriggerFeatIndex` and registers the feat when an advance is granted.

Server parity:

- Runtime projection maps these advance fields to `AllowsPlantNuke`, `EnablesCapitalization`, `EnablesInfrastructure`, `PollutionSizeModifier`, `PollutionProductionModifier`, and `TriggerFeatID`.
- Engine execution uses those fields for plant-nuke eligibility, capitalization/infrastructure conversion, city pollution modifiers, and feat effects on research completion.
- Visibility exposes capitalization/infrastructure production only when public research state satisfies their prerequisites.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailPlantNukeRequiresNuclearAdvanceWhenRuntimeDeclaresIt|TestAdvanceInfrastructureProductionConvertsShieldsToPublicWorks|TestAdvanceCapitalizationProductionConvertsShieldsToGold|TestAdvanceProductionRequiresUnlockedAdvance|TestRetailCityPollutionAppliesAdvancePollutionModifiers|TestAdvanceTriggerFeatAddsTimedNationEffects' -count=1`
- `go test ./internal/gameplay/visibility -run TestAdvanceProductionActionsRequireProjectedAdvance -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeMapsSelectedRetailValues -count=1`
