# CTP2 Full Retail Task 142: AI uses projected establish-contact actions

## Implementation
- AI planning now consumes projection-authorized `establish-contact` legal actions.
- Candidates are derived from the projected payload only, preserving the projection trust boundary.
- Duplicate contact targets are collapsed deterministically before normal candidate sorting.

## Tests
- `go test ./internal/gameplay/ai -run TestPlannerUsesProjectedEstablishContactAction -count=1`
- `go test ./internal/gameplay/ai -count=1`

## Remaining gaps
- Richer diplomatic acceptance/regard strategy remains future AI parity work.
