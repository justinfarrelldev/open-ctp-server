package units

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAllUnitInfo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request to /units!")
	fmt.Printf("\nHttp request: %v", r)
}

func GetUnitInfo(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	unitType := parts[len(parts)-1]

	fmt.Printf("\nGot request to /units/%v!", unitType)
	fmt.Printf("\nHttp request: %v", r)
}

// Define constants for unit categories
const (
	GenericUnitType = iota
	SettlerUnitType
	AerialUnitType
	NavalUnitType
	UnderseaUnitType
	AttackUnitType
	DefenseUnitType
	RangedUnitType
	FlankerUnitType
	SpecialUnitType
)

// Custom type for unit category
type UnitCategory int

// Enumerate possible unit categories
const (
	GenericUnit  UnitCategory = GenericUnitType
	SettlerUnit  UnitCategory = SettlerUnitType
	AerialUnit   UnitCategory = AerialUnitType
	NavalUnit    UnitCategory = NavalUnitType
	UnderseaUnit UnitCategory = UnderseaUnitType
	AttackUnit   UnitCategory = AttackUnitType
	DefenseUnit  UnitCategory = DefenseUnitType
	RangedUnit   UnitCategory = RangedUnitType
	FlankerUnit  UnitCategory = FlankerUnitType
	SpecialUnit  UnitCategory = SpecialUnitType
)

type Unit struct {
	Description   *string
	Category      *UnitCategory
	Attack        *int32
	Defense       *int32
	ZbRangeAttack *int32
	Firepower     *int32
	Armor         *int32
	MaxHp         *int32
	ShieldCost    *int32
	PowerPoints   *int32
	ShieldHunger  *int32
	FoodHunger    *int32
	MaxMovePoints *int32
	VisionRange   *int32
	// This is known as "enableAdvance" in the original source. This is the advance required to craft the unit.
	RequiredAdvance *advances.Advance //ADVANCE_CLASSICAL_EDUCATION
}
