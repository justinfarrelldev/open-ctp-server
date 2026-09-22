# Task 3 original-source review

- Reviewed: 2026-08-11 UTC
- Pinned source: `civctp2/civctp2@0668dd199c2519ff8ab7d2871e7c7a89e21c8e52`
- Scope reviewed: `ctp2_code/gs/utility/gstypes.h`, `ctp2_code/ui/netshell/ns_gamesetup.h` (public network setup slots and separate UI human cap), `ctp2_code/gs/utility/gameinit.cpp:2432-2454` (independent per-hotseat-slot human/robot and civilization application), and the civilization database record/parser schema.

## Observed behavior

- The engine-level capacity is 32 players, but the public network setup surface defines eight ordered slots.
- The original network UI's four-human limit is separate from the slot model and is not retained as this server's controller policy.
- Initialization applies each slot's controller and civilization choice independently; no named composition or half-human derivation is inherent in the setup data.
- Civilization selection is record-based. Selectability must be derived from validated imported metadata rather than a repository-owned retail roster.

## Implementation consequence

Use one ordered one-based `{slot, countryRecordID, controllerKind}` setup contract for counts 2 through 8 inclusive. Permit each slot to independently select `human` or `ai`; reject missing/duplicate slots, duplicate/unselectable civilization IDs, and unsupported controllers. Resolve selectable IDs from imported playable, non-special, non-barbarian civilization records. No protected civilization records, names, translations, or snapshots are copied into this repository.
