# CTP2 Full Retail Task 116: IncreaseScientists wonder research

## Source notes
- Retail `Wonder.txt` contains `IncreaseScientists 5`.
- CTP2 `CityData::GetScienceFromPops` adds `wonderutil_GetIncreaseScientists(m_builtWonders)` into each city's pop-science path.

## Implementation
- Research rate now adds the owner's active `increase_scientists` wonder value to each owner city before city research modifiers.
- This keeps the effect city-scoped rather than a single empire-wide flat bonus.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailIncreaseScientistsWonderAddsCityResearch -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact specialist assignment/pop database behavior is still simplified; this implements the deterministic city research contribution in the current economy model.
