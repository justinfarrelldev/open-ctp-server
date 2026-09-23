# CTP2 full-retail parity task 17 AI government economy evidence

Scope: make the AI government choice value the retail-like economy bonuses projected in runtime governments.

Source/data context:
- Retail governments expose production/commerce/science coefficients in `govern.txt`.
- Task 16 projects those coefficients into `RuntimeGovernmentDefinition` and engine nation effects.

Implemented server-compatible slice:
- AI government scoring now includes `ProductionPercent` and `GoldPercent` alongside research/public-works bonuses.
- AI still consumes only `visibility.Projection`; no hidden canonical state or raw retail catalog is exposed.
- Existing unlock filtering remains intact.

Validation commands:
- `go test ./internal/gameplay/ai -run PlannerUsesProjectedGovernmentUnlocksAndBonuses -count=1`
- `go test ./internal/gameplay/ai -count=1`
