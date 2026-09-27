# Task 005 source check: trade routes carry resources and gold

- `ctp2/ctp2_source_code/ctp2_code/gs/gameobj/CityData.cpp` `CityData::AddTradeRoute` records the route's source resource; source cities mark it as sold and destination cities mark it as bought.
- `CityData::CalculateGoldFromResources` adds each non-pirated source trade route's value to city trade/gold.
- `TradePool::GetGoldValue` computes resource trade value from the traded resource's gold value.

Open server implementation now gives created routes a source good and positive route gold when source goods exist, applies route gold to the source city's gold yield, and allows destination city production prerequisites to see imported route goods.
