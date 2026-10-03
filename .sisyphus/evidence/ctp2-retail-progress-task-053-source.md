# Task 053: retail CanBeGifted unit gift action

## Retail source checked

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp:4435-4445`
  - `Player::GiveUnit` selects the unit, gets the giver's capitol position, resets the unit owner to the target player, and sets the unit position to the giver capitol.
- `ctp2/ctp2_released_game/ctp2_data/default/gamedata/Units.txt:4454`
  - The retail data marks gift-capable units with `CanBeGifted`; the importer preserves it as `can_be_gifted`.

## Implementation notes

- Added canonical `gift-unit` command / `unit.gifted` event for owned units with `EffectCanBeGifted`.
- Gift execution transfers ownership to an active target player and moves the unit to the giver's capitol/fallback city tile, matching retail `GiveUnit` placement.
- The action fails closed for transported, in-space, cargo-carrying, self-targeted, non-giftable, or no-capital cases.
- Public legal actions expose gift choices only to diplomacy targets already present in the viewer's public diplomacy projection.
- Added HTTP `/game/{id}/gift-unit` forwarding with `unit_id` and `target_player_id` only.

## Red/green evidence

- Red test: `.sisyphus/evidence/ctp2-retail-progress-task-053-red.txt`
- Green targeted test:
  - `go test ./internal/gameplay/engine ./internal/gameplay/visibility ./internal/gameplay/http -run 'TestRetailCanBeGifted|TestRetailGiftUnit|TestGiftUnit|TestHandlerDelegatesGiftUnitRoute' -count=1`
