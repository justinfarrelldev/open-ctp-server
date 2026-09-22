# README World Section Completion

## TL;DR
> **Summary**: Update exactly the two remaining unchecked `README.md` World section bullets for world generation and player spawn locations, using conservative base-game/CTP2-first wording backed by the repo's runtime generator and tests.
> **Deliverables**:
> - Two completed README World checklist bullets.
> - Evidence-backed wording that does not claim unsupported exact Spyroviper or byte-for-byte base-game parity.
> - Verification evidence from diff review and `task test`.
> **Effort**: Quick
> **Parallel**: NO
> **Critical Path**: Review evidence → update README bullets → validate diff/tests

## Context
### Original Request
The user asked to implement the remaining two parts of the World section in `README.md`, with the caveat that world generation should follow the accessible base game instead of Spyroviper edition. The user also required stopping immediately if obvious bugs are found in base-game worldgen, with options explaining the bug and gameplay ramifications.

### Interview Summary
- Scope is README documentation completion, not source implementation.
- The two remaining World items are `README.md:231` world generation and `README.md:232` player spawn locations.
- Base-game worldgen bug assessment found `BUG_FOUND: no` for the runtime gameplay path.
- Verification uses tests-after: `task test` is required. No dedicated README validator exists.

### Metis Review (gaps addressed)
- Avoid overclaiming exact parity beyond evidence.
- Map every README claim to specific implementation/test references.
- Restrict edits to the two target bullets and immediately necessary surrounding wording only.
- Add hard escalation if an obvious gameplay-affecting worldgen defect or unsupported parity claim is discovered.

## Work Objectives
### Core Objective
Complete the README World section by replacing the two unchecked World bullets with completed, evidence-backed statements about base-game-aligned world generation and player spawn locations.

### Deliverables
- `README.md` has the two remaining World bullets checked/completed.
- Bullet 1 separately addresses world generation availability/base-game alignment.
- Bullet 2 separately addresses player spawn location behavior/base-game alignment.
- No source code changes unless the implementer triggers the stop condition and receives later user approval in a separate plan/session.

### Definition of Done (verifiable conditions with commands)
- `git diff -- README.md` shows changes only inside the World section around the two target bullets, unless an immediately adjacent sentence must be updated to remove Spyroviper-specific wording.
- `README.md` no longer presents the two World items as unchecked.
- The updated text contains no `Spyroviper` wording in the two target bullets.
- The updated text uses conservative language: `base-game`, `CTP2-first`, `base-game-aligned`, `implemented`, or `test-backed`; it must not claim `exact parity`, `identical`, or `byte-for-byte` behavior unless the implementer records stronger evidence in `.sisyphus/evidence/task-1-readme-world-review.md`.
- `task test` exits `0`.

### Must Have
- Claim-to-evidence mapping for both bullets before editing.
- Hard stop if a gameplay-affecting worldgen defect is discovered.
- Hard stop if the README wording cannot be truthfully supported by existing runtime code/tests.
- Agent-executed QA evidence saved under `.sisyphus/evidence/`.

### Must NOT Have
- No Spyroviper-edition parity claim for the two target bullets.
- No broad README cleanup.
- No source code, test code, Swagger, generated docs, or Taskfile changes.
- No unsupported absolute parity claims.
- No `task docs` requirement unless source comments/generated Swagger inputs are unexpectedly changed, which this plan forbids.

## Verification Strategy
> ZERO HUMAN INTERVENTION - all verification is agent-executed.
- Test decision: tests-after/docs-only validation with Go test framework via `task test`.
- QA policy: Every task has agent-executed scenarios.
- Evidence: `.sisyphus/evidence/task-{N}-{slug}.{ext}`.

## Execution Strategy
### Parallel Execution Waves
> Target: 5-8 tasks per wave. <3 per wave (except final) = acceptable here because the change is intentionally small and sequential.
> Extract shared dependencies as Wave-1 tasks for max parallelism.

Wave 1: Task 1 evidence review and stop-gate classification (`quick`)
Wave 2: Task 2 README edit and validation (`quick`)

### Dependency Matrix (full, all tasks)
| Task | Depends On | Blocks |
| --- | --- | --- |
| 1. Evidence review and stop gate | none | 2 |
| 2. Update README World bullets and validate | 1 | Final Verification Wave |

### Agent Dispatch Summary (wave → task count → categories)
| Wave | Task Count | Categories |
| --- | ---: | --- |
| 1 | 1 | quick |
| 2 | 1 | quick |
| Final | 4 | oracle, unspecified-high, unspecified-high, deep |

## TODOs
> Implementation + Test = ONE task. Never separate.
> EVERY task MUST have: Agent Profile + Parallelization + QA Scenarios.

- [x] 1. Review World Evidence and Apply Stop Gate

  **What to do**: Review the exact README World section and the cited worldgen implementation/tests. Confirm the two target bullets can be truthfully updated as base-game-aligned/test-backed documentation. Create `.sisyphus/evidence/task-1-readme-world-review.md` containing: target README lines, evidence for world generation claim, evidence for spawn claim, and `STOP_GATE: pass` or `STOP_GATE: fail`.

  **Must NOT do**: Do not edit `README.md` in this task. Do not run broad source refactors. Do not inspect Spyroviper-specific behavior as the source of truth.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: small bounded evidence review with explicit file anchors.
  - Skills: [] - No specialized skill needed.
  - Omitted: [`frontend-ui-ux`, `git-master`] - No UI or git operations required.

  **Parallelization**: Can Parallel: NO | Wave 1 | Blocks: [2] | Blocked By: []

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `README.md:229-238` - World checklist scope; only the two unchecked bullets should be targeted.
  - Pattern: `internal/game/player_endpoints.go:523-536` - Runtime start-game path that seeds, normalizes, generates, applies countries, persists the world, then starts the game.
  - Pattern: `internal/world/generator.go:66-115` - Deterministic world snapshot/turn generation.
  - Pattern: `internal/world/generator.go:117-187` - Generation input validation and supported profile enforcement.
  - Pattern: `internal/world/generator.go:238-440` - Terrain generation, continent merging, and spawn placement.
  - Pattern: `internal/game/world_contract_stubs.go:32-93` - Startup defaults and player-country-to-spawn mapping.
  - Test: `internal/world/spawn_test.go:11-184` - Spawn validity, determinism, and retry-cap coverage.
  - Test: `internal/game/start_game_world_test.go:81-420` - Game-start world generation, deterministic seed usage, wetness serialization, and persistence coverage.
  - Test: `internal/game/world_contract_test.go` - World generation contract coverage.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `.sisyphus/evidence/task-1-readme-world-review.md` exists.
  - [ ] Evidence file contains `STOP_GATE: pass` if no gameplay-affecting defect or unsupported claim is found.
  - [ ] If evidence file contains `STOP_GATE: fail`, Task 2 is not executed.
  - [ ] Evidence file separately lists support for world generation and player spawn location claims.
  - [ ] Evidence file explicitly states that README wording must avoid Spyroviper parity and unsupported exact parity language.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Happy path evidence supports README update
    Tool: Bash
    Steps: Review referenced files; write .sisyphus/evidence/task-1-readme-world-review.md with STOP_GATE: pass, world generation evidence, spawn evidence, and wording restrictions.
    Expected: Evidence file exists and contains STOP_GATE: pass plus separate evidence sections for both README bullets.
    Evidence: .sisyphus/evidence/task-1-readme-world-review.md

  Scenario: Defect or unsupported claim found
    Tool: Bash
    Steps: If review finds an obvious gameplay-affecting worldgen bug or the desired README claim is unsupported, write STOP_GATE: fail with bug/claim details, gameplay ramifications, and 2-3 remediation options; do not edit README.
    Expected: Evidence file exists with STOP_GATE: fail; Task 2 remains unchecked and unexecuted.
    Evidence: .sisyphus/evidence/task-1-readme-world-review.md
  ```

  **Commit**: NO | Message: `docs(readme): document base-game world support` | Files: [`.sisyphus/evidence/task-1-readme-world-review.md`]

- [x] 2. Update README World Bullets and Validate

  **What to do**: If Task 1 has `STOP_GATE: pass`, edit only the two target World checklist bullets in `README.md`. Mark both as completed and use conservative base-game-aligned wording. The wording must separately communicate: (1) world generation is implemented through the base-game/CTP2-first runtime generator and covered by game-start/worldgen tests; (2) player spawn locations are assigned through the base-game-aligned spawn generation/mapping path and covered by spawn/game-start tests. Then validate the diff and run `task test`.

  **Must NOT do**: Do not change source code. Do not change tests. Do not change generated Swagger docs. Do not make broad README edits. Do not claim exact base-game parity unless Task 1 evidence explicitly proves exact parity.

  **Recommended Agent Profile**:
  - Category: `quick` - Reason: targeted README edit plus standard test validation.
  - Skills: [] - No specialized skill needed.
  - Omitted: [`frontend-ui-ux`, `git-master`] - No UI or git operations required.

  **Parallelization**: Can Parallel: NO | Wave 2 | Blocks: [Final Verification Wave] | Blocked By: [1]

  **References** (executor has NO interview context - be exhaustive):
  - Pattern: `README.md:229-238` - Preserve World section style: concise checklist bullets with explicit behavior and verification note.
  - Pattern: `.sisyphus/evidence/task-1-readme-world-review.md` - Must show `STOP_GATE: pass` before editing.
  - Pattern: `internal/game/player_endpoints.go:523-536` - Runtime evidence for generation/persistence path.
  - Pattern: `internal/world/generator.go:66-440` - Runtime evidence for deterministic base-game-aligned generator and spawn placement.
  - Pattern: `internal/game/world_contract_stubs.go:32-93` - Runtime evidence for defaults and country-to-spawn mapping.
  - Test: `internal/world/spawn_test.go:11-184` - Spawn test evidence.
  - Test: `internal/game/start_game_world_test.go:81-420` - Game-start world generation test evidence.
  - Command: `task test` - Primary verification command; expected exit code `0`.

  **Acceptance Criteria** (agent-executable only):
  - [ ] `.sisyphus/evidence/task-1-readme-world-review.md` contains `STOP_GATE: pass` before the README edit.
  - [ ] `git diff -- README.md` only changes the two target World bullets and any immediately necessary adjacent wording.
  - [ ] The two World bullets are checked/completed.
  - [ ] The first updated bullet separately addresses world generation availability/base-game alignment.
  - [ ] The second updated bullet separately addresses player spawn location behavior/base-game alignment.
  - [ ] Neither target bullet contains `Spyroviper`.
  - [ ] Neither target bullet contains unsupported exact-parity terms: `identical`, `byte-for-byte`, `exact parity`, or equivalent absolute wording, unless Task 1 evidence file explicitly authorizes it.
  - [ ] `task test` exits `0`.
  - [ ] Save command output or summary to `.sisyphus/evidence/task-2-readme-world-validation.txt`.

  **QA Scenarios** (MANDATORY - task incomplete without these):
  ```
  Scenario: Happy path README update validates
    Tool: Bash
    Steps: Confirm Task 1 STOP_GATE: pass; edit README target bullets; run git diff -- README.md; run task test; save validation summary to .sisyphus/evidence/task-2-readme-world-validation.txt.
    Expected: Diff is limited to target World bullets; both bullets are completed; wording is base-game-aligned and avoids Spyroviper/exact-parity overclaims; task test exits 0.
    Evidence: .sisyphus/evidence/task-2-readme-world-validation.txt

  Scenario: STOP_GATE is missing or failed
    Tool: Bash
    Steps: Check .sisyphus/evidence/task-1-readme-world-review.md before editing. If missing or not STOP_GATE: pass, do not edit README and write validation note explaining why.
    Expected: README remains unchanged; validation evidence records blocked status and required next action.
    Evidence: .sisyphus/evidence/task-2-readme-world-validation.txt
  ```

  **Commit**: NO | Message: `docs(readme): document base-game world support` | Files: [`README.md`, `.sisyphus/evidence/task-2-readme-world-validation.txt`]

## Final Verification Wave (MANDATORY — after ALL implementation tasks)
> 4 review agents run in PARALLEL. ALL must APPROVE. Present consolidated results to user and get explicit "okay" before completing.
> **Do NOT auto-proceed after verification. Wait for user's explicit approval before marking work complete.**
> **Never mark F1-F4 as checked before getting user's okay.** Rejection or user feedback -> fix -> re-run -> present again -> wait for okay.

- [x] F1. Plan Compliance Audit — oracle

  **Tool/Agent**: Direct review agent type `oracle` via the executor's supported subagent dispatch interface.

  **Prompt to send**: `Review the final README World section implementation against .sisyphus/plans/readme-world-section.md. Inputs: final diff, .sisyphus/evidence/task-1-readme-world-review.md, .sisyphus/evidence/task-2-readme-world-validation.txt. Verify every implemented change maps to the plan and no forbidden source/test/Swagger edits occurred. Return APPROVED or REJECTED with critical deviations.`

  **Steps**: Launch direct `oracle` reviewer with the prompt above. Include the final diff, `.sisyphus/evidence/task-1-readme-world-review.md`, `.sisyphus/evidence/task-2-readme-world-validation.txt`, and this plan path.

  **Expected Approval Output**: Oracle response contains `APPROVED` and lists no critical deviations. If Oracle returns `REJECTED` or any critical deviation, fix the deviation and re-run F1-F4.

  **Evidence**: Save response to `.sisyphus/evidence/final-f1-plan-compliance.md`.

- [x] F2. Code Quality Review — general

  **Tool/Agent**: Direct review agent type `general` via the executor's supported subagent dispatch interface.

  **Prompt to send**: `Review the final README World section diff for wording quality, README consistency, absence of AI-slop phrasing, absence of unsupported exact parity claims, and presence of task test evidence in .sisyphus/evidence/task-2-readme-world-validation.txt. Return APPROVED or REJECTED with required fixes.`

  **Steps**: Launch direct `general` reviewer with the prompt above. Include the final diff in the prompt.

  **Expected Approval Output**: Reviewer response contains `APPROVED` and no required wording/test fixes. If reviewer returns `REJECTED` or required fixes, apply fixes and re-run F1-F4.

  **Evidence**: Save response to `.sisyphus/evidence/final-f2-code-quality.md`.

- [x] F3. Real Manual QA — general

  **Tool/Agent**: Direct review agent type `general` via the executor's supported subagent dispatch interface.

  **Prompt to send**: `Run or verify these checks for the README World section change: git diff -- README.md and task test. Confirm the diff is limited to the target World bullets/immediately necessary adjacent wording, both target bullets are completed, no forbidden parity words appear, and task test exits 0. Return APPROVED or REJECTED with command results and required fixes.`

  **Steps**: Launch direct `general` QA reviewer with the prompt above. The QA reviewer must execute or verify `git diff -- README.md` and `task test` and report concrete results.

  **Expected Approval Output**: QA response contains `APPROVED`, confirms the limited diff, confirms the wording constraints, and reports `task test` exit code `0`. If any command fails or diff scope is too broad, fix and re-run F1-F4.

  **Evidence**: Save response to `.sisyphus/evidence/final-f3-real-manual-qa.md`.

- [x] F4. Scope Fidelity Check — oracle

  **Tool/Agent**: Direct review agent type `oracle` via the executor's supported subagent dispatch interface.

  **Prompt to send**: `Review the original user request, final README diff, .sisyphus/evidence/task-1-readme-world-review.md, and .sisyphus/evidence/task-2-readme-world-validation.txt. Verify the completed work follows base-game/CTP2-first worldgen rather than Spyroviper, respects the bug-stop instruction, avoids unsupported exact parity claims, and does not silently expand scope. Return APPROVED or REJECTED with required fixes.`

  **Steps**: Launch direct `oracle` reviewer with the prompt above. Include original user request and final diff in the prompt.

  **Expected Approval Output**: Reviewer response contains `APPROVED` and confirms: base-game caveat honored, bug-stop rule honored, no unsupported exact parity claim, and no out-of-scope changes. If reviewer returns `REJECTED` or required fixes, apply fixes and re-run F1-F4.

  **Evidence**: Save response to `.sisyphus/evidence/final-f4-scope-fidelity.md`.

## Commit Strategy
- Do not commit unless explicitly requested by the user.
- If the user later requests a commit, use message: `docs(readme): document base-game world support`.
- Intended files: `README.md` plus `.sisyphus/evidence/*` if the project convention includes evidence artifacts; otherwise only `README.md`.

## Success Criteria
- The README World section has no remaining unchecked items for world generation or player spawn locations.
- The wording reflects base-game/CTP2-first implementation and tests without Spyroviper-specific claims.
- Any discovered gameplay-affecting bug results in a stop report instead of a misleading README update.
- `task test` passes after the README change.
