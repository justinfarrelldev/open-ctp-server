package units

import "fmt"

type Server struct {
}

// TODO add args once protobuffers are compiled and available here
func (s *Server) GetUnit(unitType UnitType) (*Unit, error) {
	if unitType == UnitType_ABOLITIONIST {
		description := "The Abolitionist is your primary weapon against the insidious attacks of foreign <L:DATABASE_UNITS,UNIT_SLAVER>Slavers<e>.  Two actions are at her disposal:  the <L:DATABASE_ORDERS,ORDER_UNDERGROUND_RAILWAY>Free Slave<e> action, which frees any slaves in the target city, and <L:DATABASE_ORDERS,ORDER_INCITE_UPRISING>Aid Uprising<e>, which can help propel a restless enemy city into full-scale <L:DATABASE_CONCEPTS,CONCEPT_RIOTS>Rioting<e>.\n\nBecause the Abolitionist is a <L:DATABASE_CONCEPTS,CONCEPT_STEALTH_UNITS>Stealth Unit<e>, it is able to see other stealth units."
		return &Unit{
			Description: &description,
		}, nil
	}

	return nil, fmt.Errorf("the unit type given (%v) is not supported by the server", unitType)
}
