# CTP2 Full Retail Task 158: social policy command and projection

## Implementation
- Added `set-social-policy` command with bounded workday/wages/rations payload validation and `government.social_policy_changed` events.
- Owner diplomacy projections now include explicit social slider values when set.
- Legal actions expose a safe `set-social-policy` action with the current/default slider payload and editable fields.

## Tests
- `go test ./internal/gameplay/engine -run TestRetailSetSocialPolicyCommandPersistsSliders -count=1`
- `go test ./internal/gameplay/visibility -run TestSocialPolicyProjectsOwnSliderValuesAndLegalAction -count=1`
- `go test ./internal/gameplay/engine -count=1`
- `go test ./internal/gameplay/visibility -count=1`

## Remaining gaps
- AI optimization and economy-side effects for social sliders remain future parity work.
