package engine

import "strconv"

// SuggestedCityName is a neutral, nontechnical fallback until localized
// civilization name lists are pinned. It does not pretend every civilization is
// Roman. Only OWN city names may be supplied: original CivilisationData chooses
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

func defaultCityName(state State, actor string) string {
	names := make([]string, 0)
	for _, city := range state.Cities {
		if city.OwnerID == actor {
			names = append(names, city.Name)
		}
	}
	return SuggestedCityName(names)
}
