# CTP2 Full Retail Task 154: government war/empire happiness fields

## Implementation
- Runtime government definitions now preserve war-discontent, conquest distress, overseas/home defeat, at-home radius, empire-distance, and too-many-cities fields from retail governments.
- Retail import maps these values directly from `govern.txt`.
- Synthetic/manual runtime compilation accepts the same optional fields.

## Tests
- `go test ./internal/gameplay/rules -count=1`

## Remaining gaps
- Applying war-discontent, conquest distress, empire-distance, and too-many-cities happiness modifiers remains future parity work.
