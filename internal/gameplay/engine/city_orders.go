package engine

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// RenameCityPayload targets a city identity, never the currently selected unit.
type RenameCityPayload struct {
	CityID string `json:"city_id"`
	Name   string `json:"name"`
}

type CityIdentityPayload struct {
	CityID string `json:"city_id"`
}

func renameCity(state *State, actor string, payload RenameCityPayload) error {
	city, ok := cityByID(*state, payload.CityID)
	if !ok || city.OwnerID != actor {
		return errors.New("city is not owned by actor")
	}
	name := strings.TrimSpace(payload.Name)
	if name == "" || !utf8.ValidString(name) || utf8.RuneCountInString(name) > 128 {
		return errors.New("city name is invalid")
	}
	for _, letter := range name {
		if unicode.IsControl(letter) {
			return errors.New("city name is invalid")
		}
	}
	for _, other := range state.Cities {
		if other.ID != city.ID && other.Name == name {
			return errors.New("city name is unavailable")
		}
	}
	city.Name = name
	return nil
}

// disbandCitySettler source-checks CityData::Disband: size cap, not the last
// owned city, and an actual settler compatible with the city's terrain domain.
// Unlike the original assertion path, absent settlers reject without destruction.
func disbandCitySettler(state State, actor, cityID string, runtime *rules.RuntimeRules) (string, bool) {
	city, ok := cityByID(state, cityID)
	if !ok || city.OwnerID != actor || city.Population < 1 || runtime == nil {
		return "", false
	}
	limit := 3 // released MAX_DISBAND_SIZE; legacy pins predate the optional field.
	if runtime.Constants.MaxDisbandSize != nil {
		limit = *runtime.Constants.MaxDisbandSize
	}
	if city.Population > limit || ownerCityCount(state, actor) < 2 {
		return "", false
	}
	for _, unit := range runtime.Units {
		if runtimeUnitHasEffect(unit, EffectSettler) && domainCanEnter(runtime, []string{unit.Domain}, EffectiveTerrain(state, city.TileID)) {
			return unit.ID, true
		}
	}
	return "", false
}

// CanDisband is an owner-only eligibility check reusing the inspection runtime.
func (i *Inspector) CanDisband(cityID, actor string) bool {
	_, ok := disbandCitySettler(i.state, actor, cityID, i.runtime)
	return ok
}

func disbandCity(state *State, actor string, payload CityIdentityPayload, runtime *rules.RuntimeRules) error {
	settler, ok := disbandCitySettler(*state, actor, payload.CityID, runtime)
	if !ok {
		return errors.New("city cannot be disbanded")
	}
	city, _ := cityByID(*state, payload.CityID)
	unitCount := len(state.Units)
	if err := completeUnitProduction(state, city, ProductionItem{ID: settler, Kind: ProductionUnit}, runtime); err != nil {
		return err
	}
	if len(state.Units) != unitCount+1 {
		return errors.New("city disband could not create a settler")
	}
	state.Units[len(state.Units)-1].Movement = 0
	for index := range state.Wonders {
		if state.Wonders[index].CityID == city.ID {
			state.Wonders[index].Destroyed = true
		}
	}
	removeCityByID(state, city.ID)
	return nil
}
