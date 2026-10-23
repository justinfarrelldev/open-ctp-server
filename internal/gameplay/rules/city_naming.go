package rules

import (
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const CityNamingVersion = "ctp2-civilization-city-names-v1"

type RuntimeCityNaming struct {
	Version       string                `json:"version"`
	Language      string                `json:"language"`
	Civilizations []RuntimeCivilization `json:"civilizations"`
}

type RuntimeCivilization struct {
	ID        string   `json:"id"`
	Aliases   []string `json:"aliases"`
	CityNames []string `json:"city_names"`
}

func (p RuntimeCityNaming) Resolve(country string) (RuntimeCivilization, bool) {
	key := strings.ToLower(strings.TrimSpace(country))
	for _, civilization := range p.Civilizations {
		for _, alias := range civilization.Aliases {
			if alias == key {
				return civilization, true
			}
		}
	}
	return RuntimeCivilization{}, false
}

func (p RuntimeCityNaming) Civilization(id string) (RuntimeCivilization, bool) {
	for _, civilization := range p.Civilizations {
		if civilization.ID == id {
			return civilization, true
		}
	}
	return RuntimeCivilization{}, false
}

func ValidateCityNaming(p *RuntimeCityNaming) error {
	if p == nil {
		return nil
	}
	if p.Version != CityNamingVersion || p.Language != "english" || len(p.Civilizations) == 0 {
		return errors.New("city naming policy is unsupported")
	}
	ids, aliases := map[string]bool{}, map[string]string{}
	for _, civ := range p.Civilizations {
		if civ.ID == "" || ids[civ.ID] || len(civ.Aliases) == 0 || len(civ.CityNames) == 0 {
			return errors.New("civilization city naming is incomplete or duplicated")
		}
		ids[civ.ID] = true
		for _, alias := range civ.Aliases {
			if alias == "" || alias != strings.ToLower(strings.TrimSpace(alias)) || aliases[alias] != "" && aliases[alias] != civ.ID {
				return errors.New("civilization aliases are invalid or ambiguous")
			}
			aliases[alias] = civ.ID
		}
		for _, name := range civ.CityNames {
			if !validCityDisplayName(name) {
				return errors.New("civilization city name is invalid")
			}
		}
	}
	return nil
}

func validCityDisplayName(name string) bool {
	if !utf8.ValidString(name) || name == "" || strings.TrimSpace(name) != name || utf8.RuneCountInString(name) > 128 {
		return false
	}
	for _, r := range name {
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func cloneCityNaming(p *RuntimeCityNaming) *RuntimeCityNaming {
	if p == nil {
		return nil
	}
	clone := *p
	clone.Civilizations = make([]RuntimeCivilization, len(p.Civilizations))
	for n, civ := range p.Civilizations {
		clone.Civilizations[n] = civ
		clone.Civilizations[n].Aliases = append([]string(nil), civ.Aliases...)
		sort.Strings(clone.Civilizations[n].Aliases)
		clone.Civilizations[n].CityNames = append([]string(nil), civ.CityNames...)
	}
	sort.Slice(clone.Civilizations, func(i, j int) bool { return clone.Civilizations[i].ID < clone.Civilizations[j].ID })
	return &clone
}
