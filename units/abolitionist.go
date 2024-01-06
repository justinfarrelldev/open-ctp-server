package units

import (
	advances "github.com/justinfarrelldev/open-ctp-server/advances"
)

var (
	description     string            = "The Abolitionist is your primary weapon against the insidious attacks of foreign <L:DATABASE_UNITS,UNIT_SLAVER>Slavers<e>.  Two actions are at her disposal:  the <L:DATABASE_ORDERS,ORDER_UNDERGROUND_RAILWAY>Free Slave<e> action, which frees any slaves in the target city, and <L:DATABASE_ORDERS,ORDER_INCITE_UPRISING>Aid Uprising<e>, which can help propel a restless enemy city into full-scale <L:DATABASE_CONCEPTS,CONCEPT_RIOTS>Rioting<e>.\n\nBecause the Abolitionist is a <L:DATABASE_CONCEPTS,CONCEPT_STEALTH_UNITS>Stealth Unit<e>, it is able to see other stealth units."
	attack          int32             = 0
	defense         int32             = 10
	zbRangeAttack   int32             = 0
	firepower       int32             = 1
	armor           int32             = 1
	maxHp           int32             = 10
	shieldCost      int32             = 540
	powerPoints     int32             = 250
	shieldHunger    int32             = 5
	foodHunger      int32             = 0
	maxMovePoints   int32             = 300
	visionRange     int32             = 1
	requiredAdvance *advances.Advance = &advances.AdvancedInfantryTactics
)

var Abolitionist Unit = Unit{
	Description:     &description,
	Category:        UnitCategory.Enum(UnitCategory_SPECIAL),
	Attack:          &attack,
	Defense:         &defense,
	ZbRangeAttack:   &zbRangeAttack,
	Firepower:       &firepower,
	Armor:           &armor,
	MaxHp:           &maxHp,
	ShieldCost:      &shieldCost,
	PowerPoints:     &powerPoints,
	ShieldHunger:    &shieldHunger,
	FoodHunger:      &foodHunger,
	MaxMovePoints:   &maxMovePoints,
	VisionRange:     &visionRange,
	RequiredAdvance: requiredAdvance,
}
