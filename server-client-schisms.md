# Server/client schisms — server resolution

Server-only delivery for the reported screenshots and source mismatches. No client
code, installed assets, live server or user game was changed. The original
investigation is retained in git (`fbc9c0c`); screenshot attribution was never
established by a live wire capture and is not invented here.

Current contracts: `docs/server-contract-repairs.md` and
`docs/client-gameplay-contract-v2.md`. Implementation/test evidence:
`plans/checklists/server-contract-repairs.md`.

## 1. River pools at fog boundaries

- [x] Source-check original `TiledMap::TryRiver`/tile postprocessing and
  `UnseenCell`'s retained `TileInfo`, not community/Apolyton river behavior.
- [x] Expose versioned `river_geometry` for legally observed center-tile artwork:
  SW/NW/NE/SE edge mask, beach-mouth discriminator and observed revision.
- [x] Persist remembered geometry without reading hidden-world changes; preserve
  unknown legacy observations instead of treating them as isolated river pools.
- [x] Test all 16 edge patterns, beach mouths, a visible center with a hidden
  continuation, hidden-neighbor exclusion, fog/reload/hidden changes/reveal.

No hidden tile IDs/terrain/deposits/units/ownership are projected to patch artwork.
The artwork fact is legitimately visible on the center tile, even at the frontier.

**Client-only follow-up:** consume the received geometry when choosing installed
original overlays; unknown neighbors must not become known no-river. No client
renderer/SDL change or visual screenshot acceptance is claimed.

## 2. Technical IDs in city display names

- [x] Separate authoritative default names from unit/account/city identities in
  found-action suggestions, AI founding and ruin-created cities.
- [x] Share bounded Unicode/control-safe normalization between founding and rename.
- [x] Remove the founding global-name uniqueness oracle; duplicate display names
  are permitted, while entity IDs remain stable and authoritative.
- [x] Test default/custom/foreign-name independence, invalid-name rollback, ruin
  names, own-only suggestions and persisted observed name behavior.

New retail pins now contain original civilization-specific city lists and English
strings from `civilisation.txt` / `civ_str.txt`, with source-order selection and
owner-only collision checks. Legacy pins retain neutral `City N`; existing
saved/user/foreign names are not silently overwritten. Original `CivilisationData`
chooses per-player names, and `CityData::SetName` does not make names global
identities. See `docs/city-development-contract.md` for naming, AI production and
authoritative national/city-influence borders plus stable player/color identities.

**Client-only follow-up:** render received names separately from IDs, retain naming
input/rename flows and fix caption/panel clipping. No UI work was done here.

## 3. Slow end turns / acting-nation strip

- [x] Publish revision/turn-bound committed `turn_context`, with sequential/long-turn
  public active actor and explicit status, without private planning queues or
  hybrid conflict participants.
- [x] Explicitly advertise `live_resolution_progress: false`; no made-up acting-AI
  sequence, queue contents, positions or mid-lock progress is returned.
- [x] Add command/read `Server-Timing` and private injected-clock command stage
  metrics for admission/lock-load, engine resolution, visibility and persistence.
- [x] Test phase modes, locked/no-live-progress semantics, bounded timing fields,
  receipt replay equality and observer clock injection.

There was no live latency trace, so no measured engine speedup is claimed. Server
measurements separate gateway work from client enqueue/read-budget/render delays.
The server resolver may publish only after commit; committed scheduler identity is
not a continuous progress stream.

**Client-only follow-up:** measure scheduling/read-worker delays and implement the
original nation strip only where legal committed identity exists, with explicit
waiting/unavailable behavior otherwise. No client strip implementation is claimed.

## 4. Impossible-looking settler path across water

- [x] Preserve authoritative movement domain/access/path validation and retained
  journey stopping semantics; existing movement/fog/replay regressions remain.
- [x] Explicitly advertise `unit_route_preview: false` and document that destination
  intent/general terrain cost is not a route or turn-count guarantee.

No server ocean-walking execution was demonstrated in the original report. A
client straight-line green preview bypassing intermediate passability is not
repaired by weakening server movement or exposing the hidden map.

**Client-only follow-up:** stop displaying fabricated legal routes/turn counts,
separate unverified destination feedback from route authorization, and consume
exact received domains/known blockers. No omniscient route endpoint was added.

## 5. Oil-looking good on plains-looking terrain

- [x] Compare original terrain Resources/goods IDs with the server's extended GOG
  installation; record the original catalog independently of installed sprite art.
- [x] Introduce new-game `ctp2-original-goods-v4`, pinned to
  `ctp2-original-resources-v1`: no fabricated THREE/FOUR variants, no shelf goods,
  and only ONE for tundra/swamp. Other generated terrain families use original
  ONE/TWO IDs. Legacy generator output is unchanged.
- [x] Fix a concrete additional lineage bug: synthetic startup flattened generated
  desert/mountain terrain to plains while retaining its deposit. New fixture pins
  support generated terrain families with explicit synthetic costs/yields; startup
  rejects unsupported terrain and invalid original initial pairs.
- [x] Expose safe generator/resource-catalog provenance without seeds, and test
  100-seed placement, determinism/other-version lineage, unknown/undefined/GOG-only
  IDs, startup handoff, fog facts and persisted reloads.

The screenshot's exact pair remains unobserved. The fixed flattening path is real,
but not asserted as that screenshot's proven cause. Original `Cell::SetTerrain`
retains deposits: later authoritative terrain changes may legitimately leave a
good on a different terrain. No silent resource migration/deletion or legacy-save
regeneration masks this distinction.

**Client-only follow-up:** check the selected installation's terrain/art/good-ID
mapping and show received authoritative facts. Do not generate/move/delete goods
locally to disguise a catalog or engine mismatch.

## 6. Repeated INFO “resyncs” / refresh versus recovery

- [x] Fix a concrete server lost-wakeup race: streaming now subscribes **before**
  its initial durable read. A commit during initial projection triggers a reread,
  rather than waiting for an unrelated later command.
- [x] Fix ambiguous failed initial reads: finite current and initial stream query
  failures return safe 503 **before** success headers, not empty 200 responses.
- [x] Expose `X-Gameplay-Read-Mode: current|stream`, no-store and initial-read timing.
- [x] Correct resync revision/cursor confusion: data carries durable `event_cursor`,
  SSE `id` uses that sequence, and subsequent wakeups resume there rather than
  repeatedly emitting full snapshots. Test unequal revision/event counters.
- [x] Test projection-time commit delivery, disconnect cleanup, initial read errors,
  durable event IDs, finite current responses and projected terminal behavior.

Durable event ordering, idempotency/receipt identity, queued-versus-executed
semantics and event rate limits are unchanged. The resync-ID bug is corrected;
wakeups remain latency hints, not event authority.
Ordinary finite current replication is not a fault or proof of rollback.

**Client-only follow-up:** distinguish routine snapshot refresh from genuine
recovery/rejection in logs/UI and investigate the shared read-worker admission
budget. No client logs were suppressed/relabelled and no live healthy-turn
acceptance session is claimed.

## Delivery boundary

All server items above are implemented with regression coverage. Client decoder,
rendering, accessible Add/Insert, rejection persistence, strip/path presentation,
UI clipping and live visual/latency acceptance are separate work, explicitly
excluded by the user's server-only scope. The server cannot honestly mark those
client/live outcomes complete; they are not unchecked server implementation tasks.
