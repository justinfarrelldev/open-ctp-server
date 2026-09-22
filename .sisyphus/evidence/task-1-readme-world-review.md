# Task 1, README world review

## World generation evidence

README.md lines 231 and 232 contain two unchecked world claims. Runtime start-world evidence is `internal/game/player_endpoints.go:StartGame`: after setup, ownership, lobby, player-count, and duplicate-country checks, it obtains a seed, normalizes generation input, calls `world.GenerateWorld`, applies player countries, persists the snapshot and turn in the same transaction, and marks the game started. Generation or persistence errors roll the transaction back.

`internal/world/generator.go:GenerateWorld`, `generateWorld`, and `validateGenerationInput` validate game configuration, seed, player order, map profile, and climate before deterministic generation. Tile generation deterministically classifies terrain and computes connected land continents. Spawn selection accepts only spawnable land-continent tiles, uses profile-aware separation with configured horizontal wrapping, and fails rather than returning a partial set when every player cannot be placed.

`internal/world/spawn_test.go` covers player-order preservation, bounds, backing tiles, spawnable terrain, land continents, unique coordinates, wrapped separation, larger profiles, deterministic repeated generation, and retry-cap failure without partial spawns. `internal/game/start_game_world_test.go` covers normalized defaults, seeded serialization, country application, transactional persistence, rollback on generation failure, and prevention of a second start.

The scoped runtime generation and spawn code exposes no other obvious gameplay-affecting defect. The cited tests support deterministic, validated server behavior only. They do not compare output with Spyroviper or a base game.

## Player spawn evidence

`NormalizeWorldGenerationInput` in `internal/game/world_contract_stubs.go` provides start-world defaults, transfers joined-player account IDs into player order, and passes the generated seed to the runtime generator. `ApplyWorldSpawnCountries` replaces temporary generator country labels with each selected country by account ID before persistence. Spawn tests verify that each generated location is valid, unique, land-based, and adequately separated.

The live world-read path is `internal/game/world_endpoints.go:GetWorld`. It calls `world.NewStore(db).Read(gameID)` and serializes the stored snapshot using `storedWorldToReadResponse`, so runtime clients receive the persisted started-world data. `internal/world/contracts.go:GetWorldHandler` and `deterministicContractWorld` are contract-handler fixtures, not the production handler or runtime world-generation evidence.

## Wording restrictions

Do not use Spyroviper behavior as the implementation source of truth. Do not claim that output is `identical`, `byte-for-byte`, or in `exact parity` with Spyroviper or either base game. Permitted documentation is limited to deterministic server-side world generation and tested valid, separated player starts for supported configurations. Determinism and validity do not prove cross-edition equivalence or matching seeded locations.

## Stop-gate decision

STOP_GATE: pass

Conservative base-game-aligned wording is supportable: the runtime server deterministically generates, persists, and returns a validated world with tested valid player spawn placement. This pass does not authorize claims of Spyroviper or base-game parity. The README remains unchanged by this task.
