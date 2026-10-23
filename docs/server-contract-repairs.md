# Server contract repairs: production, observations and diagnostics

This supplements [gameplay contract v2](client-gameplay-contract-v2.md) and supersedes
historical implementation claims in `known-server-issue.md` and
`server-client-schisms.md`. Server only: no client code or live games were changed.
New additive features remain under `contract.version: 2`.

## Complete owned-city production options

When `contract.features.production_options` is true, each currently observed owned
city has `production_options`, in the **same viewer snapshot** as its queue:

```json
{
  "revision": 7,
  "city_id": "home",
  "content_hash": "<exact pinned SHA-256>",
  "operation": "append",
  "status": "complete",
  "items": [{"id": "UNIT_WARRIOR", "kind": "unit", "cost": 10}]
}
```

- `complete` means every supported pinned unit/building/wonder/conversion candidate
  was considered, using execution's supported eligibility checks (not unimplemented
  asset SLIC policy or a claim of full retail-script parity). No first-unit limit and no
  nonempty-queue omission. Items are unique by content ID and sorted by ID.
- `complete` with `items: []` is a genuine empty result (including capacity 256).
- `unavailable` with `items: []` is **not** a complete empty or unlocked catalog:
  rules may be absent, the actor may be phase-restricted/queued/closed/locked, or
  retained queue entries may no longer pass atomic-edit revalidation. No partial
  result is published. Missing/unknown feature or field is unsupported/unknown.
- Foreign/remembered cities never expose options. No disabled hidden reason,
  hidden wonder identity, owner, location or missing foreign prerequisite leaks.
- Costs/upkeep/effects in items are server-owned. Entry IDs are absent from new
  candidate items; new entries receive identities only on accepted execution.
  Prospective completion turns are not supplied. Use the existing queue forecast
  only for its documented target/assumption, not for arbitrary new items.
- `revision`, `city_id` and `content_hash` must match the selected city snapshot.
  `turn_context` and viewer `planning_status` supply the matching phase context.
  Actor eligibility is not permission to bypass human/AI controller authentication.

**Operation boundary:** this is eligibility to retain the entire current ordered
queue and append **one** new item to an atomic replacement at that revision. It
revalidates retained entries once. It is not a general Insert, reorder, clear,
saved-template, multiple-add or funded-head-transfer quote. Duplicate non-units,
capacity, ordered retained validation and head-progress rules still apply to the
whole replacement. An unavailable append query does not prohibit submitting an
otherwise valid clear/removal/replacement through the generic edit action.

Submit exact received entry IDs plus the new `item_id` through `edit-production`.
There is no full-edit preview endpoint: every submitted replacement is checked
atomically, and rejection leaves queue, identities, funding and revision unchanged.
A new choice after refreshing needs its own command/idempotency identity; never
rebase or silently retry an old rejected choice against a newer revision.

`queue-production` legal actions now derive from this same complete result. The
retention-shaped generic `edit-production` action remains shape/edit access, not
permission to construct any catalog item. The rules catalog remains useful for
Library/template authoring, not current city authorization.

## Safe command rejections

Gameplay command errors retain their HTTP status and compatible `error` string:

```json
{
  "error": "gameplay request rejected",
  "code": "production_prerequisite",
  "message": "The proposed production queue is not currently buildable.",
  "correlation_id": "<server-generated opaque identity>"
}
```

`X-Gameplay-Correlation-ID` matches the body. Codes/messages come only from a
server allowlist; internal engine/database/decoder text never becomes public.

| Code | Meaning |
| --- | --- |
| `revision_conflict` | Refresh; the expected canonical revision is stale. |
| `idempotency_conflict` | An existing immutable identity was reused differently. |
| `controller_restricted` | Membership/controller/owner authorization restriction. |
| `phase_restricted` | Current scheduler/planning phase rejects the command. |
| `queue_identity` | Invalid city/foreign/missing/repeated retained entry identity. |
| `queue_shape` | Invalid ordered selection shape, including null/missing entries. |
| `queue_capacity` | More than 256 queue entries. |
| `content_unavailable` | No supported canonical production definition. |
| `production_prerequisite` | Whole proposed queue is not buildable; hidden causes remain private. |
| `command_invalid` | Other authoritative command validation failure. |
| `request_invalid`, `request_too_large` | Malformed request/unsupported field or 256 KiB body limit. |
| `authentication_required`, `not_found`, `state_conflict`, `rate_limited`, `internal_error` | Safe transport/service categories. |

Authentication/rate-limit middleware can reject before gameplay dispatch and may
retain its separate existing response envelope. Unknown future codes should be
handled as a generic persistent rejection, not permission to retry/rebase.

Private rejection logs correlate game ID, expected revision, correlation ID and
SHA-256 command fingerprint. Server-authored production diagnostics preserve the
internal branch without naming hidden entities. Request bodies, session credentials,
raw command/idempotency IDs and arbitrary database/decoder error strings are not
logged. Authentication store success logs no longer print bearer sessions/IDs.
An idempotent accepted receipt replay remains byte-identical; diagnostics and
wall-clock timing do not enter canonical state or receipt bytes.

A rejected command and subsequent successful state refresh are separate facts.
Current snapshots are also ordinary replication; neither a finite read nor an
unchanged planning queue proves that a command was rejected or executed.

## Rules and generation provenance

`rules.family` is `retail`, `synthetic`, or `unknown`. New retail runtime pins encode
the family inside their hashed definitions. Synthetic manifest format is explicit;
unlabelled older runtime pins remain unknown rather than guessed from IDs or
compatibility version. `rules.content_hash` remains the exact executable identity.
Old games are never migrated merely by restarting with another rules mode.

Synthetic rules are bounded, asset-independent development/regression fixtures,
**not** original CTP2 gameplay or a separate engine. The development `task run`
configuration explicitly selects retail and its installed gamedata directory.
That does not migrate an existing synthetic-pinned game. Start a new game to use
new retail content; verify its family and hash instead of assuming compatibility
version proves provenance.

The subsequent startup omission is also repaired for new retail pins: original
DiffDB human/AI chance/cap grants, current age and legal initial research are
initialized before persistence. Toolmaking is guaranteed at all six difficulties.
Population-one Warrior and population-two Settler options/Add are regression-tested
through real server growth, without relaxing the population-one Settler rejection.
Existing saves are not modified. See [retail startup contract](retail-startup-contract.md)
and `plans/checklists/retail-starting-advances.md` for exact scope and source evidence.

New games use `ctp2-original-goods-v4`. Public `world.generator_version` and
`world.resource_catalog: "ctp2-original-resources-v1"` pin the generation policy
without exposing seeds or unexplored geography. V4 retains V1 base terrain/spawns,
V2 river output and V3 ruins, replacing only resource selection with an explicit
registry of original release Resources for the generator's terrain families:
forest/plains/grassland/desert/mountain/deep ocean ONE/TWO, swamp/tundra ONE,
no shelf resources. Undefined THREE/FOUR variants are never synthesized. GOG or
community extensions and client sprite mappings cannot redefine this policy.

Startup rejects invalid initial terrain/resource pairs or unsupported pinned
terrain rather than silently relocating a desert deposit onto plains. New
synthetic fixture pins explicitly support all generated terrain families with
bounded synthetic costs/yields. They are still labelled synthetic, not retail.
Older generators/saved deposits remain unchanged and unknown metadata stays absent.
A later authoritative terrain change may retain a deposit, as original
`Cell::SetTerrain` does; this is **not** an instruction to delete transformed goods
or alter remembered observations to match current-world terrain.

## Observed river geometry

`observed_river_geometry` advertises `tile.river_geometry` for a known river tile:

```json
{"version":"ctp2-river-edges-v1","edges":10,"beach":false,"observed_revision":7}
```

`edges` is a four-bit artwork discriminator, bounded to 0..15:

| Bit | Original edge | Server `(dx,dy)` |
| --- | --- | --- |
| 0 | SW | (0,+1) |
| 1 | NW | (-1,0) |
| 2 | NE | (0,-1) |
| 3 | SE | (+1,0) |

`beach` selects original beach-mouth art; beach-to-beach edges are suppressed like
original `PostProcessTile/TryRiver`. Zero is a genuine isolated/source shape, not
unknown neighboring geography. Geometry is a legally observable fact of the
**center tile's artwork**, even at the fog frontier, not a neighbor map query.
It carries no neighbor IDs, terrain, ownership, deposits or units.

The observed revision and the tile's observed turn/stale flag identify its last
observation. Remembered geometry is persisted and never recomputed from hidden
changes. Missing legacy geometry stays unknown until observation. `river:false`
is known absence; missing geometry must not be interpreted as zero connectivity.
Clients still choose installed art locally; no asset pixels or art-table indices
are transmitted and no hidden tiles may be requested to complete a river mouth.

## City display names

Founding accepts an optional name: omitted/empty uses neutral `City N`, considering
only the player's own city names. Found actions suggest the same safe name; AI and
ruin-created cities use the authoritative default. Custom founding/rename names
share trimmed, 128-codepoint, valid-Unicode/control-safe validation. Duplicate
display names are allowed; globally checking names would disclose unknown cities.
Stable entity IDs remain separate. Existing saved/custom/foreign names are not
silently renamed. Localized civilization naming lists are not pinned yet: neutral
fallback is explicit, not a claim of civilization-specific original names.

## Committed turn context and timings

`turn_context` contains version `committed-turn-context-v1`, matching revision/turn,
mode, status, optional `active_player_id`, and `live_progress:false`. Status is
`acting`, `planning`, `conflict`, `resolving`, `ended`, or `unavailable`.
Sequential/long-turn committed state exposes the current public scheduled actor.
Hybrid conflict participants, private planning queues and future orders are not
published; do not fabricate an acting-nation strip sequence from them.

This is **not** mid-resolution progress. A resolver may hold a transaction/world
lock and publish only its committed result. `live_resolution_progress:false` is
explicit. Likewise `unit_route_preview:false`: retained destination orders and
general tile costs are not unit-specific route permission/turn-count quotes.
The engine still rejects/stops illegal movement; this server delivery does not
fix client-drawn green straight paths or add an omniscient route endpoint.

- Command HTTP responses set `Server-Timing: gameplay;dur=<milliseconds>` for
  gateway execution, excluding request decoding/authentication and client queueing.
- Event reads set `X-Gameplay-Read-Mode: current|stream` and `Server-Timing` for
  `snapshot` or `initial_read` query/projection duration. Reads use `no-store`.
  Initial read failures return safe 503 before success headers, not empty 200.
- Streaming subscribes before its initial durable read, closing the lost-wakeup
  race between reading and subscribing. Wakeups still carry no cursor authority.
- Resync data includes `event_cursor` alongside `current_revision` and `snapshot`.
  The SSE `id` is now the **durable event sequence**, not the snapshot revision;
  those counters need not be equal. The stream adopts that cursor after resync
  instead of repeating a full snapshot on every wakeup. Use `event_cursor` for
  resume and `current_revision` for model/command reconciliation, never infer one
  from the other. Command/idempotency identities and event ordering are unchanged.
- Private `Service.ObserveCommand` telemetry separates admission/lock-load,
  engine/planning/barrier resolution, visibility, persistence, and total duration.
  Its default debug logger uses bounded fields and a command fingerprint, never
  credentials/private orders. Configure a fast concurrency-safe observer before
  serving if collecting metrics. Replay is separately labelled.

No live latency benchmark, client scheduling change, auto-retry, new event-rate
limit or measured speedup is claimed. Use these facts to distinguish server work
from client admission/read-worker/render delay. Client-only decoding/rendering,
rejection-vs-refresh presentation, strip UI and live acceptance remain outside this
server-only delivery.
