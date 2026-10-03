# Task 100 source notes — field coverage ledger regression guard

Change:

- Added `TestRetailFieldCoverageLedgerMentionsEveryParsedField` beside the released-data field inventory snapshot.
- The guard reloads the retail catalogs, computes the same unique parsed field inventory, then fails if `.sisyphus/evidence/ctp2-full-retail-task-22-field-coverage-inventory.md` omits any parsed catalog field name.
- This supplements the existing count/sentinel snapshot so new parser coverage must also be reflected in the coverage/classification ledger.

Validation:

- `go test ./internal/gameplay/rules -run 'TestRetailFieldInventorySnapshot|TestRetailFieldCoverageLedgerMentionsEveryParsedField' -count=1`
