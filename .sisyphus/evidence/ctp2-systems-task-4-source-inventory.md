# Task 4 source inventory

Pinned upstream commit: `0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`
Repository: `https://github.com/civctp2/civctp2`

## Primary citations and implementation consequences

| Family / contract | Upstream path | Symbol | Task 4 consequence |
|---|---|---|---|
| units | `ctp2_code/gs/gameobj/Player.cpp` | `Player::CreateUnit` | Kept the unit contract grounded in creation-time movement/domain gating, so the synthetic unit contract requires a stable ID prefix plus minimum movement metadata instead of copied retail stats. |
| advances | `ctp2_code/gs/gameobj/Advances.cpp` | `Advances::ResetCanResearch` | Marked advances as `deferred` and required explicit prerequisite/cost tags in the synthetic contract, because upstream research availability is prerequisite-driven and not safely representable by prose. |
| buildings | `ctp2_code/gs/gameobj/buildingutil.cpp` | `buildingutil_GetTotalUpkeep` | Defined the building contract around upkeep plus effect vectors, matching the current bounded runtime/e2e slice without claiming full retail building behavior. |
| wonders | `ctp2_code/gs/gameobj/WonderTracker.cpp` | `WonderTracker::AddBuilt` | Required a uniqueness-oriented wonder contract and kept wonders `included-now` only for bounded production/effect wiring and global uniqueness tracking. |
| terrain | `ctp2_code/gs/world/WrlImprove.cpp` | `Terrain::GetFood` | Anchored terrain records to stable movement/yield semantics, which is why the terrain contract requires movement metadata and stays bounded rather than copying full terrain tables. |
| terrain improvements | `ctp2_code/gs/world/WrlImprove.cpp` | `World::InsertImprovement` | Separated terrain-improvement records from base terrain records and marked them `deferred`, because upstream inserts improvements as distinct world objects. |
| governments | `ctp2_code/gs/gameobj/Player.cpp` | `Player::InitPlayer` | Marked governments `deferred` and required unlock/cap tags in the future contract, because upstream startup binds government state to advances, readiness, and tax caps. |
| trade routes | `ctp2_code/gs/gameobj/TradeRouteData.cpp` | `TradeRouteData::GeneratePath` | Combined route/agreement scope under one deferred family so later tasks must prove deterministic path/cost state before claiming readiness. |
| agreements | `ctp2_code/gs/gameobj/AgreementData.cpp` | `AgreementData::BeginTurnOwner` | Added lifecycle pressure to the same deferred trade/agreement family because upstream agreements carry owner/recipient/breach state over time. |
| pollution sources | `ctp2_code/gs/gameobj/Pollution.cpp` | `Pollution::GetGlobalPollutionLevel` | Defined pollution sources as their own deferred synthetic family because upstream aggregates multiple player/world contributors instead of a single record kind. |
| pollution terrain effects | `ctp2_code/gs/world/WrldPoll.cpp` | `World::GlobalWarmingEvent` | Kept pollution bounded and deferred because upstream world mutation is event-driven and would overclaim parity if imported now. |
| special actions | `ctp2_code/gs/gameobj/ArmyData.cpp` | `ArmyData::EstablishEmbassy` | Kept special actions `included-now` as a bounded vector contract already mirrored by the local manifest. |
| special actions | `ctp2_code/gs/gameobj/ArmyData.cpp` | `ArmyData::Soothsay` | Confirmed that special actions must stay action-specific rather than being folded into generic unit stats. |
| special actions | `ctp2_code/gs/gameobj/UnitData.cpp` | `UnitData::StealTechnology` | Justified a dedicated synthetic action/vector contract for tech theft rather than copying retail advance names. |
| special actions | `ctp2_code/gs/gameobj/Player.cpp` | `Player::CanTerraform` | Justified keeping terraform in the bounded special-action family with separate future terrain-improvement work. |

## Local implementation consequences

- Added a machine-readable `MechanicMatrix` to `SupportedRulesManifest`, bounded to the exact statuses `included-now`, `deferred`, and `explicitly-excluded`.
- Added `SyntheticContracts` to the same manifest so imports, canonical JSON, runtime validation, and e2e consumers all see one stable Task 4 contract surface.
- Required every `included-now` row to declare non-empty importer/runtime/engine/api/ai readiness fields.
- Required complete coverage of the ten synthetic families: units, advances, buildings, wonders, terrain, terrain-improvements, governments, trade-routes-agreements, pollution-sources, and special-actions.
- Preserved current runtime compatibility by leaving actual record compilation behavior unchanged; Task 4 only adds explicit bounded inventory/contract validation.
- Updated only the acceptance hashes that drifted because canonical rules metadata legitimately changed after the new manifest contract was added.
