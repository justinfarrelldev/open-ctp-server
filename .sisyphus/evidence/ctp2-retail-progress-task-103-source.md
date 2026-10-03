# Task 103 source notes — special/spy action vector coverage verified

Retail source evidence:

- `UnitRecord.cpp` parses unit special-action structs/flags for the vectors in released data, including nested chance/death/effect/timer/amount fields.
- `UnitData.cpp:3718-4161` executes city-target actions such as investigate, steal technology, incite revolution, assassinate ruler, establish embassy, and throw party.
- `UnitData.cpp:5406-5551` exposes target legality checks for plant nuke, create parks, underground railway, embassy, franchise, assassination, steal, injoin, incite, cause unhappiness, and expel.
- `ArmyData.cpp:1144-1902`, `ArmyData.cpp:2027-2410`, and `ArmyData.cpp:2536-3868` select eligible units and execute/project the full army-level action family: investigate, steal technology, incite revolution, assassinate ruler, create franchise, sue/sue franchise, expel, cause unhappiness/advertise/soothsay, plant nuke, slave raids/settler slave raids, underground railway, establish embassy, throw party, bio/nano/plague terror, convert city, reform, indulgence sales, and create park.

Server parity:

- Engine tests cover every imported special-action vector and nested success/death/effect/timer/amount behavior that mutates canonical state.
- Legal-action projection uses public projection data and suppresses stale/hidden targets; the AI consumes the projected action list rather than hidden state.
- Presentation-only nested `Sound` fields have already been classified as UI/audio metadata.

Validation:

- `go test ./internal/gameplay/engine -run 'TestRetailPlantNukeEscapeChanceTeleportsToNearestOwnCity|TestRetailSueUnitRemovesCanBeSuedTargetsOnTile|TestRetailUndergroundRailwayFreesSlaveToNearestHomeCity|TestRetailReformCityRemovesInjunctionAndCrime|TestRetailExpelMovesAdjacentEnemyUnit|TestRetailAssassinateRulerDestabilizesTargetNation|TestRetailCreateParkCleansPollutedAdjacentTile|TestRetailInvestigateCityStoresPrivateCityReport|TestRetailSlaveUprisingTransfersCityToRebels|TestRetailIndulgenceSalesTransferGoldAndTimedHappiness|TestRetailInciteRevolutionUsesChanceAndTransfersCityToRebels|TestRetailCauseUnhappinessAddsTimedHappinessAttack|TestRetailCreateFranchiseMarksCityOnce|TestRetailSlaveRaidsReducePopulationAndHonorProtection|TestRetailBioNanoAndPlagueSpecialsAffectCityPopulation|TestRetailConvertCityHonorsProtectionAndDisclosure|TestTask11EmbassySoothsayingAndDisclosureVectors|TestRetailStealTechnologySpecificChanceStealsRequestedAdvance|TestTask11StealTechnologySortedRNGFailureCaptureAndDisclosure' -count=1`
- `go test ./internal/gameplay/visibility -run 'TestAdvertiseLegalActionUsesRetailAdvertiseRoute|TestThrowPartyLegalActionUsesPublicGoldAndDiplomacy|TestRetailSueFranchiseActionProjectsForOwnFranchisedCity' -count=1`
- `go test ./internal/gameplay/http -run 'TestHandlerDelegatesThrowPartyRoute|TestHandlerDelegatesAdvertiseRoute' -count=1`
