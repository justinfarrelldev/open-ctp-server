# AI production, original city names, territory and influence

Server-only repairs. Clients render authoritative results; they must not derive
territory/influence from nearby cities, infer human identity from surviving units,
or substitute guessed civilization names. No client code or live saves are edited.

## AI production scheduling

An idle city's ordinary production candidate now scores **300**, above Pass /
round-complete (100), routine exploration (150) and army formation (180). Urgent
research/government/policy/combat work retains its higher priority. Closing the
round is still a fallback, not a reason to leave an otherwise serviceable city idle.
The actual runtime remains bounded at **one action followed by closeout**; no
unbounded AI loop, extra permission or artificial unit-spawning path is introduced.

For current projections, preference ranking is restricted to that city's complete
`production_options`. Unavailable or complete-empty options produce no ordinary
production candidate. A stronger locally ineligible unit cannot hide a weaker
eligible one, and submitted item metadata comes from the exact advertised item.
Foreign/stale cities and already-filled queues are excluded. Existing compatibility
projection fixtures without options retain their older bounded catalog preference
path; production runtime projections explicitly supply the options/status.

Original source: `ai/CityManagement/governor.cpp::FillEmptyBuildQueues` computes
next items and actually queues units/buildings/wonders for eligible AI cities.
These server scores are not claimed to reproduce original goal scores byte-for-byte.

## Civilization-specific city names

New retail definitions pin `city_naming`, version
`ctp2-civilization-city-names-v1`, including:

- The ordered `CIV_CITY_NAME` references from original `civilisation.txt`, resolved
  through `english/gamedata/civ_str.txt` (also supports modern `CityName` records).
- Original spelling/punctuation, decoded from Windows-1252 when necessary. Incidental
  shipped padding such as `"Mataiva "` / `"Kosice "` is trimmed, as founding already
  normalizes user-provided names. No invented universal/Roman list is used.
- Explicit English language policy; no mutable local userprofile/locale inference.
- Source record IDs and normalized ID/country/singular/plural aliases, so `Romans`
  resolves to `ROMAN` and `Greeks` to `GREEK`. The pre-existing public setup slug
  `america` is explicitly retained for `AMERICANS`, even in extended tables whose
  CountryName is `United States`; no generic singularization/suffix guessing.

Each new nation persists its resolved `civilization_id`. A supplied country not
present in the selected retail catalog rejects startup safely (400), rather than
silently borrowing another civilization's names. Catalogs differ: the original
release has 42 civilization records; the configured extended installation can add
more. For example, Aztec is an extension in the inspected selected catalog, not an
original release civilization record. The server uses the selected pinned catalog,
not a hardcoded client list or a guessed singularization.

Founding with omitted/empty name, AI founding and ruin-granted cities all use the
same pinned chooser. Original Roman cities begin Rome, Pompeii; Greek cities begin Athens,
Sparta; the original American capital spelling is **Washington D.C.**. Only that
owner's current city names are collision inputs. Exhausted lists cycle with the
original numeric suffix style (`Rome1`, `Pompeii1`, then `Rome2`, ...), checking every
candidate even against custom names. This fixes original counter/collision edge
cases without a global hidden-city-name oracle or mutable allocation counter.
Custom valid names are preserved. Capture never silently renames an existing city.

The owner-only `city_naming` projection contains `civilization_id`, `language` and
`suggested_name`. Legal founding/rename suggestions use the same chooser as
execution. Merely reading a projection does not allocate a name or mutate state.

Names are in exact canonical rules definitions/hash; restarting/loading an existing
save does not read installed string files. Legacy pins without a catalog retain
neutral own-only `City N` defaults. Existing cities are not automatically renamed;
create a new retail game to get the newly pinned lists.

Source: `CivilisationData::GetAnyCityName`, `GetCityName`, `UseCityName`,
`cityNameIsUsedByPlayer`; `CityData` construction. The chooser preserves source list
order and suffix style, but intentionally uses safe current-own-name collision
checks rather than reproducing unsafe/stale original counter behavior.

## National territory versus city influence

New retail rules pin `city_territory`, version `ctp2-city-territory-v1`:

- `Const.txt` `BORDER_INT_RADIUS` / `BORDER_SQUARED_RADIUS` (original **4 / 20**).
- Ordered `citysize0.txt` population/int-radius/squared-radius records. The default
  first city uses population threshold 6, radius **1 / 2**; population 7 selects
  threshold 18, radius **2 / 5**. The last record clamps higher populations.
- Population, not specialist worker allocation, selects the influence size.

Founding/ruin cities claim their center and connected national radius. National
flood-fill accepts unowned/same-owner tiles and stops at foreign ownership. Missing
map cells are barriers too. The server uses its existing **bounded cardinal grid**
and squared coordinate distance, not the original wrapped/skew MapPoint geometry;
this repair does not claim to implement wrapped gameplay or silently change topology.

City influence is a distinct canonical `tiles[].influence_city_id`. It flood-fills
the selected city-size radius and stops at another city's assigned tiles. Existing
assignments win overlap; all actual centers are reserved. Remaining simultaneous
unassigned claims use stable city-ID order, never slice/map iteration order or a
client nearest-city/Voronoi guess. Growth expands and shrink releases influence.
Capture/recapture transfers that city's influence/ownership and regenerates allowed
national expansion. Destroy/disband/environmental removal clears missing-city
assignments. Disconnect/AI takeover retains the civilization's cities and influence;
it is not a city-destruction event.

**National ownership is not the same as influence.** Like the inspected original
`PrepareToRemove` / disabled `RADIUS_OP_RESET_OWNER` path, shrink/city removal clears
influence without blanket-erasing historical national claims. Capture transfers the
city's assigned area; outer national tiles can retain their previous owner if foreign
ownership blocks the subsequent flood-fill. No dangling live influence is allowed.
A larger influence area can claim unassigned national tiles, as original
`ExpandInfluence` sets both CityOwner and Owner when assigning them.

Legacy pins do not acquire guessed radius tables or retroactive city regions on
load/read. New founding on a legacy pin at least claims its explicit center; the
full source-driven radius behavior requires a new retail pin. No database migration
or live repair of pre-fix cities is performed.

Sources: `CityInfluenceIterator.cpp::{GenerateBorders,ExpandBorders,
GenerateCityInfluence,ExpandInfluence}`; `CityData::{ComputeSizeIndexes,
AdjustSizeIndices,ResetCityOwner,PrepareToRemove}`. Flood fill is iterative rather
than recursive, preventing source-style stack exhaustion on large regions.

### Wire/fog semantics

For each observed tile:

- `owner_id` is national ownership; `""` explicitly means observed unowned.
- `influence_city_id` is the assigned city ID; `""` explicitly means observed no
  city influence. It is not the tile's city-center occupant field.
- `null`/omission of optional remembered facts means unavailable/older observation,
  not an inferred empty assignment. Unexplored tiles are not synthesized.
- Fog remembers the **last observed** assignment and observation turn. Hidden
  capture, growth, destruction or reassignment is not read through fog. Reobserving
  updates the fact, including an explicit empty value after influence loss.

An observed influence ID is not permission to retrieve that hidden city's current
name, owner, population, queue or position. Canonical reload preserves fog memory;
HTTP, receipts, SSE and AI use the common viewer-safe projection.

## Stable viewer, player and palette identities

Projection now explicitly supplies:

```json
{
  "viewer_player_id": "account-7",
  "players": [
    {"id":"account-7","nation_id":"nation-account-7","civilization_id":"ROMAN","color_index":0},
    {"id":"account-8","nation_id":"nation-account-8","civilization_id":"GREEK","color_index":1}
  ]
}
```

Participant IDs/selected civilizations are public setup metadata. Palette slots are
persisted at new-game initialization in stable account-ID order, independent of
roster ordering, controller takeover, surviving units or current city ownership.
Older saves without a color slot omit it rather than pretending every player has
slot zero. This is a zero-based participant palette index, not an RGBA or a claim
to import every original game's palette. Client color selection and political /
dashed influence border rendering remain client work.

## Verification

The authenticated isolated public-runtime regression uses actual retail assets,
real server city growth/production, persisted in-memory SQL and the unchanged
one-action worker. It verifies Rome/Athens, claimed territory and influence after
founding, queued **and completed city-produced** AI units (not ruin rewards), runtime
reconstruction, stable identities, and immutable founding receipt replay with no
canonical state change. Unit regressions cover source parsing, canonical pinning,
human/AI suggestions, own-only collisions, exhaustion, foreign barriers, overlap,
growth/shrink/capture/removal, explicit empty observations and hidden-fact isolation.

Checklists: `plans/checklists/ai-production-civilization-names.md` and
`plans/checklists/city-territory-influence.md`.
