# CTP2 Full Retail Task 128: grouped special-action death chances

## Implementation
- Added action-specific `*_death_chance` runtime effects for retail special action blocks with `DeathChance`.
- Runtime importer now projects `DeathChance` for multi-action retail units (for example Cyber Ninja investigation and Abolitionist underground railway).
- Engine now resolves deterministic post-action death chances for grouped special actions:
  - convert city,
  - create franchise,
  - investigate city,
  - plant nuke,
  - bio/nano/plague terror,
  - assassinate ruler,
  - slave/settler slave raids,
  - underground railway.
- Death resolution removes the acting unit/cargo after the action outcome while preserving the original disclosure-safe event.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSpecialDeathChanceRemovesActingUnitAfterResolvedActions -count=1`
- `go test ./internal/gameplay/rules -run TestProjectRetailRuntimeIncludesCompleteReleasedCatalog -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Retail elite chances, escape chances, and diplomacy regard side effects remain future grouped work.
