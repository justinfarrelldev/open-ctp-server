# CTP2 Full Retail Task 172: conquest distress

## Source evidence
- `UnitData.cpp` calls `ResetConquestDistress(g_player[newo]->GetConquestDistress())` when a city changes owner.
- `Happy.cpp::CalcConquestDistress` applies conquest distress to city happiness and decays it by `ConquestDistressDecay`.

## Implementation
- Canonical cities now store `ConquestDistress`.
- Government conquest distress and decay fields become active nation effects.
- Capturing a city initializes conquest distress from the new owner's government effects.
- City happiness applies conquest distress and turn recalculation decays it toward zero.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailConquestDistressAppliesToCapturedCityAndDecays -count=1`
- `go test ./internal/gameplay/engine -count=1`
