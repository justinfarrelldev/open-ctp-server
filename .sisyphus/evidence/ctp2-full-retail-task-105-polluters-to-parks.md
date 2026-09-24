# CTP2 Full Retail Task 105: polluters-to-parks wonder effect

## Source references
- `Player.cpp` handles `PollutersToParks` by selecting the most polluted cities, excluding the city that built the wonder, and calling `CityToPark`.
- `CityData::CityToPark` removes city-radius improvements/units and kills the city.

## Implementation
- Wonder completion now applies `polluters_to_parks` by selecting the top polluted city tiles (stable by pollution descending, city ID tie-breaker), excluding the wonder city.
- Converted cities are removed, their city tiles' pollution is cleared, non-park overlays on those tiles are replaced by park overlays owned by the wonder owner, units on converted city tiles are removed, and wonders in removed cities are marked destroyed.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailPollutersToParksWonderConvertsMostPollutedCities -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Exact retail radius-wide improvement/unit removal around converted cities is simplified to city-tile removal until city-radius ownership/worked-tile rules are fully modeled.
