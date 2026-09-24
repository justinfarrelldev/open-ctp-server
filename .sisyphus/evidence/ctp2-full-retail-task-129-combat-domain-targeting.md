# CTP2 Full Retail Task 129: grouped combat domain target selection

## Implementation
- Combat strikes now select the first eligible living target for each attacker instead of always striking the first defender in ID order.
- Bombardment now filters target stacks to eligible targets before resolving active defense and bombard rounds.
- Submarine targets now require explicit underwater attack capability; sea-only attack flags no longer bypass submarine stealth/domain rules.

## Tests
- `go test ./internal/gameplay/engine -run 'TestRetail(CombatDomainMasksSelectEligibleTargets|BombardDomainMasksSelectEligibleTargets|SubmarineRequiresUnderwaterAttackCapability)' -count=1`
- `go test ./internal/gameplay/engine -count=1`

## Remaining gaps
- Full CTP2 stack ordering, retreat rules, stealth visibility exceptions, and air/naval interception details remain future grouped work.
