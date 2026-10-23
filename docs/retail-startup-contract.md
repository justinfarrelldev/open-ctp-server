# Retail starting discoveries and research

New retail games now initialize starting discoveries, current age and legal initial
research **before** initial visibility and transactional game-start persistence.
This fixes the reported population-one city with no discoveries/research and zero
eligible units. It does not remove production prerequisites or population checks.

## Authoritative setup and pinned policy

The existing owner-controlled `rules_json` setup array accepts these retail options:

```json
["difficulty:medium", "start-age:AGE_ONE"]
```

Difficulty IDs follow the original six DiffDB blocks, in source order:

| ID | Original index | Max human advances | Max AI advances |
| --- | --- | --- | --- |
| `beginner` | 0 | 6 | 2 |
| `easy` | 1 | 6 | 3 |
| `medium` | 2 | 4 | 4 |
| `hard` | 3 | 3 | 4 |
| `very-hard` | 4 | 3 | 6 |
| `impossible` | 5 | 3 | 8 |

These table values describe the verified original data, not hardcoded executable
caps. The importer reads actual ordered `ADVANCE_CHANCES` human/AI pairs and
`MAX_HUMAN_ADVANCES` / `MAX_AI_ADVANCES` from the configured retail catalog. It
requires six well-formed profiles and valid advance references. Missing, repeated,
negative or malformed values fail closed instead of silently starting empty.

The explicit server default is **medium**, with the pin's first age (`AGE_ONE` in
the original catalog). This is a server default, not a guess from a local mutable
`userprofile.txt`, the client's UI, or an assertion that every original player used
medium. A selected `start-age` must exist within the pinned start/end age interval.
Unknown/empty/duplicate starting selections return a safe setup rejection (400)
and cannot leave a partially persisted game. Synthetic fixtures reject these
retail-only selections rather than pretending to apply them.

New canonical retail definitions include versioned `starting_advances` policy
`ctp2-starting-advances-v1`, the selected difficulty, all six source-ordered profiles,
and `start_age_id`. They are covered by the exact rules content hash. Public
`rules.starting_difficulty` and `rules.start_age_id` expose the selected settings,
not individual players' discoveries, random rolls or seeds. Absence means no such
versioned policy in that pin; never infer grants for an older game.

## Original grant semantics

The implementation follows original release
`ctp2_source/ctp2_code/gs/gameobj/Player.cpp::InitPlayer` (around lines 505–575),
`gs/database/DiffDB.cpp::ParseAdvanceChances` (325–362), and
`gs/gameobj/Advances.cpp::InitialAdvance` (436):

1. At the first age, walk chance rows in **file order**, not alphabetical order.
2. Select human versus AI chances/cap using the persisted controller kind.
3. A chance >=100 grants unconditionally and increments the granted count, even
   if the cap is already reached. Thus the cap is not a blanket ceiling on
   guaranteed discoveries.
4. A lower chance draws only while the cumulative grant count is below the cap.
   A draw in `[0,100)` succeeds exactly when it is **less than** the chance. Zero
   chances still consume a draw while eligible, as in the source. Skipped capped
   rows consume none. Starting grants do not first require researching prerequisites.
5. A later-age start instead grants every advance from a strictly earlier age;
   it does not additionally roll the first-age difficulty list.
6. Store the selected current age and choose an initial legal research target
   after all grants. The existing deterministic canonical advance-ID order is the
   research tie-break; the original scans database order, so identical initial
   target selection is not promised for every modified/later-age catalog. If no
   legal research remains within the ending-age boundary, the target stays empty.

All six original profiles guarantee Toolmaking for humans and AI. Impossible
also guarantees Agriculture. Other starting discoveries depend on the source
chances and caps; they are not all granted merely to make production succeed.
Initial grants are not earned-research completion: they allocate no command/event
IDs, award no research-completion feats, and do not execute asset SLIC.

The server uses a dedicated, deterministic per-player bootstrap stream derived
from the generated world's private RNG seed/state. This is not a claim to replay
the original game's global RNG byte-for-byte. It does not consume terrain/spawn or
ongoing gameplay RNG streams, and one player's draws cannot consume another's.
No client-authored draw, seed, discovery or grant count is accepted.

## Persistence, restart and existing games

The isolated engine bootstrap commits nation changes only after successful
validation. Grants, legal research and a one-time initialization marker persist in
the same new-game transaction as the world, pin and initial event. A failed start
cannot leave a world with partial grants. Duplicate start, runtime restart and
immutable command receipt replay do not reroll or regrant anything.

**Existing saves, including the reported games 58 and 59, are not rewritten.**
Loading canonical state does not invoke initialization. Legacy/synthetic pins
without the policy remain unchanged. An uninitialized existing gameplay state
cannot be passed through new-game bootstrap to overwrite progress.

Restart the server and create a **new retail game** to obtain the new pin/policy.
For an existing pre-fix game, manually selecting and completing Toolmaking remains
the non-destructive way to unlock Warrior; simply waiting with an empty research
target does not discover it. Settler independently requires population above one.

## Regression coverage and boundaries

- Actual startup across all six profiles, mixed human/AI, legal research, current
  age, roster-order independence and deterministic canonical replay.
- First city at population one advertises/accepts Warrior and rejects Settler.
- The city's real pinned economy grows it to population two, without assigning
  population or food stores; Settler then appears in complete options and its
  atomic Add succeeds, including on a nonempty production queue. All six profiles
  pass this transition.
- Source-order cap behavior, guaranteed-over-cap grants, strict chance threshold,
  zero-chance draw behavior, later ages, malformed policy/RNG rollback and legacy
  no-initialization behavior.
- Authenticated in-memory public startup persists the grants before founding;
  invalid setup rolls back, corrected retry succeeds, duplicate start and runtime
  restart/receipt replay preserve state. No live service or user game is used.

This repair implements **starting-discovery difficulty policy**, not all original
difficulty systems. It does not newly implement difficulty-dependent gold/settler
bonuses, AI handicaps, science/food bonuses or script policy, and does not change
the server's calendar epoch for later-age starts. Existing supported rules govern
those mechanics. No client UI or private prerequisite implementation is added.

Evidence: `plans/checklists/retail-starting-advances.md`.
