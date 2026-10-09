# Local Singleplayer and Autonomous AI Setup Plan

## Goal

Allow a locally running Open CTP server to host a single-human game with server-managed AI opponents without requiring fake human clients. The same autonomous AI driver must also run for AI-controlled slots in ordinary multiplayer games.

## Requirements

- Preserve the existing multiplayer setup and command authority model; singleplayer is a setup/orchestration path, not a separate gameplay rules mode.
- Allow an authenticated owner to add server-managed AI players to a setup-stage game without pre-existing placeholder accounts.
- Support any AI count currently bounded by the CTP2 setup limit: up to seven AI opponents with one human player, with lower counts also supported.
- Preserve the existing owner flow for converting an existing joined player to AI control.
- Persist AI slots as ordinary `game_players` with `controller_kind = 'ai'` so world generation, start state, persistence, events, and command receipts stay canonical.
- Reject duplicate/blank AI country selections, non-owner AI setup, non-setup games, and games that already have the supported maximum player count.
- Start games with mixed human/AI slots using the existing `/game/start_game` path and initialize AI controllers in canonical gameplay state.
- Add a production AI worker that discovers due AI players in started games and drives them through the same gateway/receipts/events as test/runtime AI.
- Drive AI in sequential, long-turn, simultaneous planning, hybrid planning, and hybrid conflict phases without submitting commands for inactive or closed AI players.
- Wire the AI worker into `main.go` and expose it through `runtime.Runtime` for integration tests.
- Document the localhost singleplayer setup flow.
- Use red-green-refactor with automated tests for setup, worker due-player selection, and runtime wiring.
- Complete a final review against this plan before stopping.
