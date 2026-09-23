# CTP2 full-retail parity task 20 research trade yield evidence

Scope: make research progression consume city trade/gold yield in addition to explicit research effects.

Original source checked:
- `ctp2_code/gs/gameobj/CityData.cpp::SplitScience` computes science from city trade using the science tax rate, applies building/wonder science percentages, adds population science, subtracts crime, then applies government knowledge coefficient.
- `ctp2_code/gs/gameobj/CityData.cpp::GetScienceFromPops` applies building science per-pop and government knowledge coefficient to population science.

Implemented server-compatible slice:
- Research rate now includes each owned city's current `Yields.Gold` as the bounded server trade/science proxy.
- Nation-level research effects are applied once instead of being counted again through city effects.
- City/building/wonder research effects remain included through a local city research effect helper.

Validation commands:
- `go test ./internal/gameplay/engine -run 'RetailResearchRateUsesCityTradeYield|Task6ResearchCompletesAdvancesAndUnlocksProduction' -count=1`
