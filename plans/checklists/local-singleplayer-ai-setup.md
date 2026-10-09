# Local Singleplayer and Autonomous AI Setup Checklist

## Planning

- [x] Created branch `singleplayer-ai-local-setup` from the current branch.
- [x] Created the local singleplayer/AI setup plan.
- [x] Created this implementation checklist.

## Red-Green-Refactor Slices

- [x] Add failing setup tests for creating server-managed AI players through `/game/add_ai` without placeholder accounts.
- [x] Implement server-managed AI account/player creation while preserving existing target-account AI conversion.
- [x] Refactor AI setup helpers and validate max-player/country/owner/setup rules.
- [x] Add failing due-AI worker tests for sequential, simultaneous, hybrid, inactive, queued, and closed AI cases.
- [x] Implement due-AI discovery and autonomous worker execution through the existing AI driver.
- [x] Refactor worker/store boundaries and harmless concurrency conflict handling.
- [x] Add failing runtime/main wiring tests or source guards proving the AI worker is exposed and started.
- [x] Wire the production AI worker into runtime construction and `main.go`.
- [x] Add/update local singleplayer documentation.
- [x] Run targeted package tests after each slice.
- [x] Run full repository tests where feasible.
- [x] Run `git diff --check`.
- [x] Complete final code review against the plan.

## Final Review Notes

- [x] Verified server-managed AI setup remains an ordinary setup-stage game/player flow and does not introduce a separate gameplay rules mode.
- [x] Verified AI commands are still submitted only through `ExecuteAICommand` via the production AI gateway/driver.
- [x] Verified the production `main.go` starts the autonomous AI worker and runtime tests guard that wiring.
- [x] Verified one-human/seven-AI local setup and autonomous multiplayer AI execution with public-runtime E2E coverage.

## Acceptance Criteria

- [x] A local owner can create a no-lobby game, choose a human country, add 0-7 server-managed AI slots with countries, and start the game through existing endpoints.
- [x] A one-human/seven-AI setup starts with eight active canonical players and AI controllers for the seven server-managed slots.
- [x] Existing joined-player-to-AI conversion remains supported.
- [x] AI players in multiplayer and singleplayer started games are driven automatically by the server worker.
- [x] AI commands use the existing authoritative command service, receipts, event outbox, visibility projection, and idempotency behavior.
