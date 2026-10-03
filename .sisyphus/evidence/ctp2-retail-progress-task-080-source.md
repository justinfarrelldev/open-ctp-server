# Task 080 source check: readiness support floors after summing fractional costs

Retail source evidence:

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Readiness.cpp:116-146` computes support cost as a `double` per unit, including readiness coefficient, special-forces war coefficient, wonder readiness reduction, and government `SupportCoef`.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Readiness.cpp:197-207` accumulates those per-unit double costs into `m_cost` without per-unit integer truncation.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:2138-2140` floors the summed readiness cost with `(sint32)(m_readiness->GetCost() + 0.0000001)` when charging production.
- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Readiness.cpp:220-226` and `:285-293` compare `sint32(m_cost)` to the production budget while killing unsupported units, so fractional costs remain summed before integer comparison.

Implementation notes:

- Readiness support candidates now retain fractional `float64` costs.
- Production support tax floors only after summing all candidate costs.
- Over-budget disband checks also compare the floored summed cost against the budget while subtracting full fractional candidate costs.

Validation evidence:

- RED: `.sisyphus/evidence/ctp2-retail-progress-task-080-red.txt`
- GREEN engine targeted: `.sisyphus/evidence/ctp2-retail-progress-task-080-green-engine.txt`
- GREEN engine package: `.sisyphus/evidence/ctp2-retail-progress-task-080-green-packages.txt`
- GREEN full repo: `.sisyphus/evidence/ctp2-retail-progress-task-080-green-full.txt`
- Freeze guard: `.sisyphus/evidence/ctp2-retail-progress-task-080-freeze.txt`
- Diff check: `.sisyphus/evidence/ctp2-retail-progress-task-080-diff-check.txt`
