# City-founded national territory and city influence

Server only. New retail pins include original Const border radii and ordered
citysize0 population/radius records. No live database repair, save migration,
client renderer changes or inferred client-side ownership.

- [x] Inspect original CityInfluenceIterator::GenerateBorders/ExpandBorders,
  GenerateCityInfluence/ExpandInfluence, CityData::ComputeSizeIndexes,
  AdjustSizeIndices, ResetCityOwner and PrepareToRemove.
- [x] RED → GREEN/refactor: import/pin/validate original border and city-size policy;
  preserve legacy pins and bounded server topology; document coordinate differences.
- [x] RED → GREEN/refactor: founding and ruins claim the center/national radius and
  distinct city influence; foreign claims block national expansion; city overlap
  is deterministic and never replaces another city's existing influence.
- [x] RED → GREEN/refactor: growth/shrink, capture/recapture, disband/razing/removal
  update assignments without dangling references; disconnect/takeover retains cities
  and influence; original national claims persist
  when influence is cleared by city loss/shrink, unlike city influence.
- [x] RED → GREEN/refactor: visible influence, explicit observed-empty assignments,
  fog memory/reobservation and canonical restart; no hidden city/live assignment
  leakage or projection-side state mutations.
- [x] Authoritative viewer/player/civilization identity and stable color slots, not
  inferred from surviving units; no client palette or rendering implementation.
- [x] Public founding receipt/current-view integration plus AI worker/unit-production
  regressions under actual one-action/closeout configuration.
- [x] Thorough source/privacy/determinism/legacy review; updated contracts/checklists;
  full suite/race/vet/build/freeze/format/diff checks and normal atomic commits.

Source facts: national ownership and city assignment are separate Cell fields.
National flood fill starts at center, accepts unowned/same-owner tiles and stops
at foreign ownership. Influence stops at another city's assignment. Population
chooses the first city-size record whose Population is >= current population,
clamped to the last record, independent of specialist worker utilization. National
claims are not blanket-erased on shrink/removal (the source reset-owner case is
explicitly disabled); city-influence assignments must be cleared safely.

## Evidence and review

- RED importer/engine regressions established missing policy, center/national claims
  and city assignments. GREEN pins 4/20 national radii plus ordered source city-size
  thresholds (first positive population <=6 has influence 1/2, >6 selects 2/5).
  Version/radius/population/order/reference checks fail closed; no installed file is
  read during canonical load. Legacy pins do not infer newly imported radii.
- Iterative cardinal flood fill is bounded by pinned radii/map cells. Foreign-owned
  national walls and missing map cells cannot be jumped. Existing city influence
  blocks competitors; actual centers are reserved. Source-owned assignments are
  retained, with stable city-ID order only for remaining unassigned simultaneous
  claims. Tile/center indexes avoid per-tile world scans during reconciliation.
- Source AdjustSizeIndices and PrepareToRemove clear city influence, not all national
  claims. Capture transfers assigned cells then runs national expansion, which stops
  at foreign historical outer claims. Tests cover center/ring separation, overlap,
  population growth/shrink, capture/recapture, removal and authoritative disband.
- Departure is a scheduled controller takeover, not city destruction: an explicit
  regression preserves its city/influence. Removing a city clears dangling live
  assignments; validation rejects references to missing cities or ownership mismatch.
- RED fog test → pointer-valued observed influence. Explicit empty, stale memory,
  hidden capture/reassignment/name changes, canonical reload, reobservation and
  read-only projection are tested. Fog never dereferences remembered IDs into live
  hidden city details. All transports use the same projection.
- Explicit viewer ID, public participant/civilization IDs and persisted zero-based
  color slots remain usable after the last settler disappears and after restart.
  Slots are stable under roster reordering; legacy missing slots remain unknown.
  This does not claim client RGBA/dashed-border drawing or wrapped/skew gameplay.
- The public original-assets runtime fixture checks Rome's founding receipt/current
  view assignments, Greek AI development, restart and immutable receipt replay.
  It uses in-memory SQL, not a live game/database or asset mutation.

Contract and scope: `docs/city-development-contract.md`. No client inference,
renderer changes, automatic renaming, legacy save migration or database repair.

Final verification: complete default suite and core/onboarding race suite passed;
additional fullgame race deadline retry passed in isolation without changing the
per-scenario budget. Vet/build, 87-file freeze, gofmt and diff checks passed. Coverage
81.2% overall / 86.6% engine (existing 90% target not weakened). Full commands/logs
and replay-pin evidence are in `ai-production-civilization-names.md`. All changes
committed through unchanged normal hooks; no live/user/client/asset mutations.
