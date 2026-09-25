# CTP2 Full Retail Task 186: riot chance rolls

## Source evidence
- `CityData.cpp::CheckRiot` only checks rioting when happiness is below `ConstDB::RiotLevel`.
- Retail `Const.txt` sets `RIOT_LEVEL 73`.
- `CheckRiot` rolls `g_rand->Next(100)` against `(RiotLevel - happiness) * Player::GetRiotChance()`.
- Retail `DiffDB.txt` sets `RIOT_CHANCE 10.0` for released difficulties.

## Implementation
- Added an explicit canonical `riot` RNG domain to new state initialization; validation recognizes it while allowing legacy three-domain snapshots.
- Low-happiness riot handling now rolls deterministic canonical RNG against retail chance rather than always rioting below 73.
- Riot RNG advances during city recalculation only for cities below the riot level.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailRiotChanceCanLeaveUnhappyCityCalm -count=1`
- `go test ./internal/gameplay/engine -run TestRetailLowHappinessCityRiotsAndLosesProductionGold -count=1`
- `go test ./internal/gameplay/engine -run TestRNGDomainsAreNamedAndSerializable -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/game -count=1`
