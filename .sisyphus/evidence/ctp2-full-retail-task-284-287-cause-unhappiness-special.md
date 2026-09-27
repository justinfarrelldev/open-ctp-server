# CTP2 Full Retail Tasks 284-287: cause-unhappiness / advertise special execution

## Implementation
- Added `cause-unhappiness` as an executable special-action command using retail `CauseUnhappiness` chance/timer/amount metadata.
- Successful actions spend special movement, set the target city watchful, and add a timed negative happiness effect without disclosing canonical city IDs in the public event payload.
- Added generic timed effect expiry (`Effect.Turns`) so happiness attacks expire deterministically on subsequent city recalculations.
- Safe projections and HTTP routes expose the action only for owned units with the `special_cause_unhappiness` capability; visible enemy projections still suppress unit capabilities.
- AI command scoring recognizes the projected action.

## Source notes
- Retail `ArmyData::CauseUnhappiness` reads `Chance`, `Timer`, and `Amount`, applies a happiness timer with negative amount, and sets the target city watchful. Retail `UNIT_CORPORATE_BRANCH` uses the same block for advertising.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailCauseUnhappinessAddsTimedHappinessAttack -count=1`
- `go test ./internal/gameplay/visibility -run TestTask12ProjectionIncludesRulesAndLegalActionMetadata -count=1`
- `go test ./internal/gameplay/http -count=1`
- `go test ./internal/gameplay/ai -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`
