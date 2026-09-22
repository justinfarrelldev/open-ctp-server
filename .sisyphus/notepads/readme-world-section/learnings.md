# Learnings

- Task 1 gate failed: deterministic generation and valid-spawn tests support conservative server behavior only, not Spyroviper parity. `GetWorldHandler` currently returns a fixed `deterministicContractWorld` instead of the persisted started-world snapshot, so world reads can disagree with the actual game.

- Correction: that earlier route conclusion was false. `main.go` registers `game.GetWorldHandler`, which dispatches to `internal/game/world_endpoints.go:GetWorld`; it reads the persisted snapshot with `world.NewStore(db).Read(gameID)` and serializes it with `storedWorldToReadResponse`. The contract helper is not production routing, so the conservative evidence gate passes.

- Task 2 completed the two World bullets with conservative wording: deterministic server-side generation and persistence through the base-game/CTP2-first path, plus test-backed valid, separated starts for supported configurations. `GIT_MASTER=1 git diff -- README.md` was limited to those bullets and `task test` exited 0.

