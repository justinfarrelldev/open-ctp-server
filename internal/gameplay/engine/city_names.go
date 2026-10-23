package engine

import (
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
	"strconv"
)

// SuggestedCityName is the neutral legacy/unmapped fallback. New retail games
// instead use their pinned civilization's original ordered list.
// Only OWN city names may be supplied: original CivilisationData chooses
// names used by that player, not global names (which would be a fog oracle).
func SuggestedCityName(ownedNames []string) string {
	used := make(map[string]bool, len(ownedNames))
	for _, name := range ownedNames {
		used[name] = true
	}
	for n := 1; ; n++ {
		name := "City " + strconv.Itoa(n)
		if !used[name] {
			return name
		}
	}
}

// SuggestedCivilizationCityName walks the source list before recycling it with
// the original numeric suffix style. Check every candidate against OWN existing
// names, fixing original counter/custom-name collision cases without a fog oracle.
func SuggestedCivilizationCityName(ownedNames, cityNames []string) string {
	if len(cityNames) == 0 {
		return SuggestedCityName(ownedNames)
	}
	used := map[string]bool{}
	for _, name := range ownedNames {
		used[name] = true
	}
	for round := 0; ; round++ {
		suffix := ""
		if round > 0 {
			suffix = strconv.Itoa(round)
		}
		for _, base := range cityNames {
			runes := []rune(base)
			if len(runes) > 128-len(suffix) {
				runes = runes[:128-len(suffix)]
			}
			name := string(runes) + suffix
			if !used[name] {
				return name
			}
		}
	}
}

type CityNamingSuggestion struct {
	CivilizationID string `json:"civilization_id,omitempty"`
	Language       string `json:"language"`
	SuggestedName  string `json:"suggested_name"`
}

func (i *Inspector) CityNaming(actor string) *CityNamingSuggestion {
	if i.runtime == nil || i.runtime.CityNaming == nil {
		return nil
	}
	result := &CityNamingSuggestion{Language: i.runtime.CityNaming.Language, SuggestedName: i.SuggestedCityName(actor)}
	for _, player := range i.state.Players {
		if player.ID == actor {
			if nation := nationForID(&i.state, player.NationID); nation != nil {
				result.CivilizationID = nation.CivilizationID
			}
		}
	}
	return result
}

func (i *Inspector) SuggestedCityName(actor string) string {
	return defaultCityName(i.state, actor, i.runtime)
}

func defaultCityName(state State, actor string, runtimes ...*rules.RuntimeRules) string {
	names := make([]string, 0)
	for _, city := range state.Cities {
		if city.OwnerID == actor {
			names = append(names, city.Name)
		}
	}
	var runtime *rules.RuntimeRules
	if len(runtimes) > 0 {
		runtime = runtimes[0]
	}
	if runtime == nil {
		runtime, _ = runtimeRulesForState(state)
	}
	if runtime != nil && runtime.CityNaming != nil {
		for _, player := range state.Players {
			if player.ID == actor {
				if nation := nationForID(&state, player.NationID); nation != nil {
					if civ, ok := runtime.CityNaming.Civilization(nation.CivilizationID); ok {
						return SuggestedCivilizationCityName(names, civ.CityNames)
					}
				}
			}
		}
	}
	return SuggestedCityName(names)
}
