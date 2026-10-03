# Task 122 source evidence: government/readiness/social/rush coefficient closure

## Retail source checks

- Government coefficients are parsed by `GovernmentRecord` and exposed via getters for food/production/gold/knowledge, pollution, crime, readiness, support, war discontent, social sliders, martial law, empire-distance, city-count, rush-buy, science caps, trade caps, and `ProfessionalUnits` (`GovernmentRecord.cpp`, `GovernmentRecord.h`).
- City/economy source uses government coefficients in city yield, happiness/crime, pollution, and social slider calculations (`CityData.cpp`, `Happy.cpp`, `TaxRate.cpp`, `Readiness.cpp`).
- Readiness affects support cost/production and HP modifiers (`Readiness.cpp`, combat paths). Rush-buy uses government rush modifiers by production kind (`bldque.cpp`/queue buy paths). Government trade caps constrain route counts.
- `ProfessionalUnits` is parsed/exposed to AI/outcom (`Player::GetProfessionalUnits`, `C3Government::GetProfessionalUnits`, robot `scienceagent.cpp`/`wotp.cpp`) but no released gameplay execution call site beyond AI access was found; server preserves professional-unit support-free execution through explicit unit state markers and classifies the government scalar as AI/projection metadata.

## Server coverage/classification

- Runtime imports/projection cover government economy coefficients (`FoodCoef`, `ProductionCoef`, `GoldCoef`, `KnowledgeCoef`), pollution/unhappiness, crime, social slider expectations and positive/negative coefficients, readiness support/HP/transition metadata, war discontent, conquest distress/decay, peace-movement home/overseas coefficients/decays, empire-distance bounds, too-many-cities penalty, martial-law caps/effects/threshold, rush modifiers, capitalization/infrastructure coefficients, science caps, and trade caps.
- Engine coverage includes government effect upsert/removal on government change; city yield/research/gold/food/production effects; science-rate cap; public-works/social-policy commands; readiness HP combat modifier; readiness support production tax/disband behavior; crime, slider happiness, martial law, empire-distance, and city-count happiness penalties; rush-buy modifiers; and trade-route caps.
- `ProfessionalUnits` is source-classified as AI/outcom metadata; server support-free units use canonical `Unit.IsProfessional` instead of inventing a hidden government-count exemption.

## Validation

- Engine GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-122-engine-green.txt`
- Rules/ledger GREEN: `.sisyphus/evidence/ctp2-retail-progress-task-122-rules-green.txt`
