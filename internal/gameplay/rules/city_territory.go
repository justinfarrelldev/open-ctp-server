package rules

import (
	"errors"
	"os"
	"path/filepath"
)

const CityTerritoryVersion = "ctp2-city-territory-v1"

type RuntimeCityTerritory struct {
	Version             string                     `json:"version"`
	BorderIntRadius     int                        `json:"border_int_radius"`
	BorderSquaredRadius int                        `json:"border_squared_radius"`
	Sizes               []RuntimeCityInfluenceSize `json:"sizes"`
}

type RuntimeCityInfluenceSize struct {
	ID            string `json:"id"`
	Population    int    `json:"population"`
	IntRadius     int    `json:"int_radius"`
	SquaredRadius int    `json:"squared_radius"`
}

func (p RuntimeCityTerritory) SizeForPopulation(population int) RuntimeCityInfluenceSize {
	for _, size := range p.Sizes {
		if population <= size.Population {
			return size
		}
	}
	if len(p.Sizes) > 0 {
		return p.Sizes[len(p.Sizes)-1]
	}
	return RuntimeCityInfluenceSize{}
}

func ValidateCityTerritory(p *RuntimeCityTerritory) error {
	if p == nil {
		return nil
	}
	validRadius := func(i, s int) bool { return i >= 0 && i <= 64 && s >= 0 && s <= 2*i*i }
	if p.Version != CityTerritoryVersion || p.BorderIntRadius < 1 || !validRadius(p.BorderIntRadius, p.BorderSquaredRadius) || len(p.Sizes) == 0 {
		return errors.New("city territory policy is unsupported or malformed")
	}
	ids := map[string]bool{}
	population, radius, squared := -1, -1, -1
	for _, size := range p.Sizes {
		if size.ID == "" || ids[size.ID] || size.Population < 0 || size.Population < population || !validRadius(size.IntRadius, size.SquaredRadius) || size.IntRadius < radius || size.SquaredRadius < squared {
			return errors.New("city influence sizes are invalid or unordered")
		}
		ids[size.ID] = true
		population, radius, squared = size.Population, size.IntRadius, size.SquaredRadius
	}
	return nil
}

func loadRetailCityTerritory(directory string) (*RuntimeCityTerritory, error) {
	constFile := filepath.Join(directory, "Const.txt")
	data, err := os.ReadFile(constFile)
	if err != nil {
		return nil, err
	}
	tokens, err := lexRetailCTP2(string(data), constFile)
	if err != nil {
		return nil, err
	}
	values := map[string]int{}
	for n, token := range tokens {
		if token.value != "BORDER_INT_RADIUS" && token.value != "BORDER_SQUARED_RADIUS" {
			continue
		}
		if _, seen := values[token.value]; seen || n+1 >= len(tokens) || tokens[n+1].line != token.line {
			return nil, errors.New("city border radius is missing or duplicated")
		}
		value, err := retailParseInt("CONST", token.value, tokens[n+1].value)
		if err != nil {
			return nil, err
		}
		values[token.value] = value
	}
	if len(values) != 2 {
		return nil, errors.New("city border radii are missing")
	}
	sizesFile := filepath.Join(directory, "citysize0.txt")
	data, err = os.ReadFile(sizesFile)
	if err != nil {
		return nil, err
	}
	records, err := ParseRetailRawRecords(sizesFile, data, RetailCatalog{Kind: "city-size", Prefixes: []string{"CITYSIZE_"}})
	if err != nil {
		return nil, err
	}
	p := &RuntimeCityTerritory{Version: CityTerritoryVersion, BorderIntRadius: values["BORDER_INT_RADIUS"], BorderSquaredRadius: values["BORDER_SQUARED_RADIUS"]}
	for _, record := range records {
		values := map[string]int{}
		for _, field := range record.Fields {
			switch field.Key {
			case "Population", "IntRadius", "SquaredRadius":
				if _, seen := values[field.Key]; seen || len(field.Values) != 1 || len(field.Fields) != 0 {
					return nil, errors.New("city influence radius is malformed or duplicated")
				}
				value, err := retailParseInt(record.ID, field.Key, field.Values[0])
				if err != nil {
					return nil, err
				}
				values[field.Key] = value
			}
		}
		if len(values) != 3 {
			return nil, errors.New("city influence radius/threshold is missing")
		}
		p.Sizes = append(p.Sizes, RuntimeCityInfluenceSize{ID: record.ID, Population: values["Population"], IntRadius: values["IntRadius"], SquaredRadius: values["SquaredRadius"]})
	}
	if err := ValidateCityTerritory(p); err != nil {
		return nil, err
	}
	return p, nil
}

func cloneCityTerritory(p *RuntimeCityTerritory) *RuntimeCityTerritory {
	if p == nil {
		return nil
	}
	clone := *p
	clone.Sizes = append([]RuntimeCityInfluenceSize(nil), p.Sizes...)
	return &clone
}
