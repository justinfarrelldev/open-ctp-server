# Task 124 source evidence: readiness, martial law, war discontent, conquest distress

## Retail source checks

- `GovernmentRecord` parses readiness coefficients/HP modifiers, turns to new readiness, support/defense coefficients, martial-law fields, war-discontent fields, conquest distress/decay, at-home/overseas defeat coefficients and decays (`GovernmentRecord.cpp`/`.h`).
- Retail readiness/support/HP and transition behavior lives in `Readiness.cpp`; combat HP modifiers are applied through combat/unit HP paths.
- Retail happiness/war-discontent/martial-law/conquest-distress mechanics are handled through `Happy.cpp`, `CityData.cpp`, `CityEvent.cpp`, and peace-movement/defeat tracking in player/government paths.

## Server coverage

- Runtime imports readiness (`ReadyPeaceCoef`, `ReadyAlertCoef`, `ReadyWarCoef`, HP modifiers, `TurnsToNewReadiness`) and government support/defense coefficients.
- Engine command support persists readiness state and projects legal readiness actions. Support tax/disband behavior, professional-unit marker exemption, and HP combat modifier execute.
- `ReduceReadinessCost` wonder effects stack and cap safely.
- Martial law, war discontent, conquest distress/decay, home/overseas defeat coefficients/decays, and owned/home/overseas distinctions are covered by city happiness/peace-movement tests.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-124-green.txt`
