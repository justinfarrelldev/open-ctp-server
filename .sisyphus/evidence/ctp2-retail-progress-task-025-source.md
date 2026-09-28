# Task 025 source check: trade-route actions require current visible cities

Retail/source context:

- CTP2 trade route execution is city-based: trade routes are managed through city/trade-pool state and must attach to live source/destination cities owned/known at execution time.
- The server engine mirrors this by rejecting route creation when the source city is unavailable or not owned by the actor.

Implementation notes:

- Public legal-action projection no longer emits `create-trade-route` actions from stale fog-memory city records.
- This keeps AI consumption projection-only while ensuring legal trade-route candidates are based on current visible city evidence, not obsolete memory.
