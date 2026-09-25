# CTP2 Full Retail Task 174: specialist yields

## Source evidence
- Retail `Pop.txt` defines specialist outputs: entertainers add happiness, farmers food, laborers production, merchants commerce, and scientists science.
- `CityData.cpp` adds farmer/laborer specialist outputs to city food/production and scientist output to research paths.

## Implementation
- Canonical cities now persist specialist assignments in a safe `specialists` map.
- Specialist assignments are validated for known retail specialist IDs, nonnegative counts, and total specialists not exceeding city population.
- City yields now include farmer/laborer/merchant/scientist outputs; entertainers contribute city happiness.
- Research rate includes city science output from scientists.
- Owner city projections expose specialist assignments.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSpecialistsContributeYieldsHappinessAndResearch -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
