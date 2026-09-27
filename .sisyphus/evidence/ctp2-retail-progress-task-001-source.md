# Task 001 Source Notes

- Original CTP2 visibility/ZOC code checks current player visibility before considering enemy units: `ctp2/ctp2_source_code/ctp2_code/gs/world/WrlUnit.cpp` lines 134-159 call `g_player[owner]->IsVisible(pos)` and then require visible enemy units for visible-only ZOC checks.
- Original player visibility delegates to the current vision map: `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/Player.cpp` lines 4224-4227 return `m_vision->IsVisible(pos)`.
- Server implication for this slice: combat/bombard legality must accept targets visible in the current unit/city sight map, not only tiles already persisted to fog memory by previous moves. Fog memory remains a fallback for previously observed legal targets where current projections allow them.
