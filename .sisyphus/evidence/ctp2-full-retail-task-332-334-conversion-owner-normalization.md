# CTP2 Full Retail Tasks 332-334: conversion cleanup on ownership changes

## Source notes
- Retail `CityData` clears conversion when a city is converted to its current owner.
- City ownership changes from conquest/revolt can therefore make prior conversion state invalid.

## Implementation
- Added canonical conversion-owner normalization that clears `converted_to_player_id` / `converted_by` whenever the converted-to player becomes the city owner.
- City capture invokes conversion normalization after owner transfer.
- Revolt/rebel ownership transfer invokes conversion normalization after owner transfer.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetailCityCaptureClearsConversionWhenConvertedOwnerCaptures|TestRetailInciteRevolutionHonorsRevoltingCitiesJoinPlayerWonder' -count=1`
- `go test ./internal/gameplay/engine -count=1`
