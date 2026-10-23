package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func loadRetailCityNaming(directory string) (*RuntimeCityNaming, error) {
	civ, err := os.ReadFile(filepath.Join(directory, "civilisation.txt"))
	if err != nil {
		return nil, err
	}
	localized, err := os.ReadFile(filepath.Join(directory, "..", "..", "english", "gamedata", "civ_str.txt"))
	if err != nil {
		return nil, err
	}
	return ParseRetailCityNaming(civ, localized)
}

// ParseRetailCityNaming resolves ordered CityName references from the original
// civilisation DB through its English string table. Text is pinned, never loaded
// on save decode, and legacy Windows-1252 is decoded before canonical JSON.
func ParseRetailCityNaming(civilizations, localized []byte) (*RuntimeCityNaming, error) {
	if !utf8.Valid(localized) {
		var err error
		localized, err = charmap.Windows1252.NewDecoder().Bytes(localized)
		if err != nil {
			return nil, err
		}
	}
	localized = []byte(strings.TrimPrefix(string(localized), "\ufeff"))
	tokens, err := lexRetailCTP2(string(localized), "civ_str.txt")
	if err != nil {
		return nil, err
	}
	texts := map[string]string{}
	for n := 0; n < len(tokens); n += 2 {
		if n+1 >= len(tokens) || tokens[n].line != tokens[n+1].line {
			return nil, fmt.Errorf("civ_str.txt: expected key and localized string at line %d", tokens[n].line)
		}
		key, value := tokens[n].value, tokens[n+1].value
		if _, exists := texts[key]; exists {
			return nil, fmt.Errorf("civ_str.txt: duplicate string %s", key)
		}
		texts[key] = value
	}
	records, err := ParseRetailRawRecords("civilisation.txt", civilizations, RetailCatalog{Kind: "civilization", Prefixes: []string{""}})
	if err != nil {
		return nil, err
	}
	policy := &RuntimeCityNaming{Version: CityNamingVersion, Language: "english"}
	for _, record := range records {
		civ := RuntimeCivilization{ID: record.ID}
		aliases := map[string]bool{strings.ToLower(record.ID): true}
		// Existing public setup uses "america". Modern tables rename the
		// CountryName to "United States", while the source identity remains
		// AMERICANS. Preserve that explicit setup alias, never guess suffixes.
		if record.ID == "AMERICANS" {
			aliases["america"] = true
		}
		if label := texts[record.ID]; label != "" {
			aliases[strings.ToLower(strings.TrimSpace(label))] = true
		}
		fields := map[string]bool{}
		for _, field := range record.Fields {
			key := field.Key
			switch key {
			case "CIV_COUNTRY":
				key = "CountryName"
			case "CIV_SINGULAR":
				key = "SingularCivName"
			case "CIV_PLURAL":
				key = "PluralCivName"
			case "CIV_CITY_NAME":
				key = "CityName"
			}
			switch key {
			case "CountryName", "SingularCivName", "PluralCivName", "CityName":
				capital := field.Key == "CIV_CITY_NAME" && len(field.Values) == 2 && field.Values[1] == "CITY_IS_CAPITAL"
				if len(field.Values) != 1 && !capital || len(field.Fields) != 0 || key != "CityName" && fields[key] {
					return nil, fmt.Errorf("civilisation.txt: malformed %s in %s", field.Key, record.ID)
				}
				text, ok := texts[field.Values[0]]
				// Shipped tables contain incidental trailing spaces (Mataiva,
				// Kosice); founding normalizes the same padding for custom names.
				text = strings.TrimSpace(text)
				if !ok || !validCityDisplayName(text) {
					return nil, fmt.Errorf("civilisation.txt: invalid/missing localized reference %s", field.Values[0])
				}
				fields[key] = true
				if key == "CityName" {
					civ.CityNames = append(civ.CityNames, text)
				} else {
					aliases[strings.ToLower(text)] = true
				}
			}
		}
		if !fields["CountryName"] || !fields["SingularCivName"] || !fields["PluralCivName"] {
			return nil, fmt.Errorf("civilisation.txt: missing identity in %s", record.ID)
		}
		for alias := range aliases {
			civ.Aliases = append(civ.Aliases, alias)
		}
		sort.Strings(civ.Aliases)
		policy.Civilizations = append(policy.Civilizations, civ)
	}
	if err := ValidateCityNaming(policy); err != nil {
		return nil, err
	}
	return cloneCityNaming(policy), nil
}
