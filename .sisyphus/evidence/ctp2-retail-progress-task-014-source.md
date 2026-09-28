# Task 014 source check: AI peace proposal actions

Retail CTP2 diplomacy models peace treaties separately from trade pacts:

- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/diplomattypes.h` defines treaty proposal types including peace and trade-pact treaty types.
- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/proposalresponseevent.cpp` registers treaty response handlers such as `PeaceTreaty_ProposalResponseEvent` and the default treaty response path.
- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/proposalanalysis.cpp` extracts treaty types from proposals and evaluates/accepts treaty conditions, including trade pact and other treaty proposal cases.

Implementation notes:

- Visibility already projects legal `propose-agreement` actions for peace when a relation is at hot war.
- The AI planner now consumes projected legal peace proposals as bounded public actions, alongside trade-pact proposals, so combat-created wars have a legal server-side recovery path before trade can resume.
