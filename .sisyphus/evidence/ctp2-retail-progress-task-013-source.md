# Task 013 source check: AI trade pact proposals before trade routes

Retail CTP2 diplomacy has explicit trade-pact proposal handling and AI trade-route management:

- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/diplomattypes.h` defines `PROPOSAL_TREATY_TRADE_PACT`.
- `ctp2/ctp2_source_code/ctp2_code/ai/diplomacy/nproposalevent.cpp` contains `TradePact_NewProposalEvent`, checks whether a trade pact already exists, computes proposal priority, and creates a trade-pact proposal (`new_proposal.detail.first_type = PROPOSAL_TREATY_TRADE_PACT`).
- `ctp2/ctp2_source_code/ctp2_code/ai/ctpai.cpp` calls `Governor::GetGovernor(player).ManageGoodsTradeRoutes()` in AI turn processing, so trade agreements are a prerequisite for the later goods-trade route loop.

Implementation notes:

- Visibility already exposes legal `propose-agreement` actions for trade pacts.
- The AI planner now consumes those projected legal trade-pact actions, rather than waiting for a union proposal, so server-side AI can establish trade pacts through public projections before creating trade routes.
