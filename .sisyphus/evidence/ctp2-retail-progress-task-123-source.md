# Task 123 source evidence: social slider parity closure

## Retail source checks

- Retail government records define `WorkdayExpectation`, `WagesExpectation`, `RationsExpectation`, and positive/negative coefficients for each slider (`GovernmentRecord.cpp`/`.h`).
- Retail tax/social-policy source applies workday/wages/rations to production, gold/wages, food/rations, and happiness/unrest calculations (`TaxRate.cpp`, `CityData.cpp`, `Happy.cpp`, `GovernmentTab.cpp`).
- AI strategy data contains slider settings/elements that influence AI social-policy choices (`aidata/strategies.txt`, AI strategy projection paths).

## Server coverage

- Runtime import preserves all slider expectations and coefficients.
- `set-social-policy` validates/persists workday/wages/rations, emits a public event, and projections expose only owner slider state plus legal actions.
- City yield/economy applies workday/wages/rations to production, gold, and food consumption.
- Happiness applies expectation delta with positive/negative coefficients.
- AI legal-action projection uses retail strategy public inputs and current visible city fit; no hidden-state shortcut is used.

## Validation

- GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-123-green.txt`
