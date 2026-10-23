package rules

import (
	"os"
	"strings"
)

func loadRetailStartingAdvanceDifficulties(path string) ([]StartingAdvanceDifficulty, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, importError(ErrorInvalidDirectory, path, 1, "", err.Error())
	}
	return ParseRetailStartingAdvanceDifficulties(data)
}

// ParseRetailStartingAdvanceDifficulties reads the six unnamed original DiffDB
// blocks, preserving the human/AI chance pair and row order. It does not interpret
// unrelated difficulty bonuses or execute asset scripts.
func ParseRetailStartingAdvanceDifficulties(data []byte) ([]StartingAdvanceDifficulty, error) {
	const file = "DiffDB.txt"
	tokens, err := lexRetailCTP2(string(data), file)
	if err != nil {
		return nil, err
	}
	var result []StartingAdvanceDifficulty
	for index := 0; index < len(tokens); {
		if tokens[index].value != "{" || len(result) >= len(startingDifficultyIDs) {
			return nil, importError(ErrorMalformedValue, file, tokens[index].line, "DIFFICULTY", "expected six source-ordered difficulty blocks")
		}
		fields, next, err := parseRetailFields(tokens, index+1, file, "DIFFICULTY")
		if err != nil {
			return nil, err
		}
		profile := StartingAdvanceDifficulty{ID: startingDifficultyIDs[len(result)]}
		capValue := func(key string) (int, error) {
			var matching []RetailRawField
			for _, field := range fields {
				if field.Key == key {
					matching = append(matching, field)
				}
			}
			if len(matching) != 1 || len(matching[0].Values) != 1 || len(matching[0].Fields) != 0 {
				return 0, importError(ErrorMalformedValue, file, 1, key, "expected one cap value")
			}
			value, err := retailParseInt("DIFFICULTY", key, matching[0].Values[0])
			if err != nil {
				return 0, err
			}
			if value < 0 {
				return 0, importError(ErrorMalformedValue, file, 1, key, "must be nonnegative")
			}
			return value, nil
		}
		profile.MaxHumanAdvances, err = capValue("MAX_HUMAN_ADVANCES")
		if err != nil {
			return nil, err
		}
		profile.MaxAIAdvances, err = capValue("MAX_AI_ADVANCES")
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		blocks := 0
		for _, field := range fields {
			if field.Key == "ADVANCE_CHANCES" {
				blocks++
				if len(field.Values) != 0 || len(field.Fields) == 0 {
					return nil, importError(ErrorMalformedValue, file, 1, field.Key, "chance rows missing")
				}
				for _, row := range field.Fields {
					if !strings.HasPrefix(row.Key, "ADVANCE_") || len(row.Values) != 2 || len(row.Fields) != 0 || seen[row.Key] {
						return nil, importError(ErrorMalformedValue, file, 1, field.Key, "expected unique advance and human/AI chance pair")
					}
					human, err := retailParseInt(row.Key, "human chance", row.Values[0])
					if err != nil {
						return nil, err
					}
					ai, err := retailParseInt(row.Key, "AI chance", row.Values[1])
					if err != nil {
						return nil, err
					}
					if human < 0 || ai < 0 {
						return nil, importError(ErrorMalformedValue, file, 1, field.Key, "chances must be nonnegative")
					}
					profile.Chances = append(profile.Chances, StartingAdvanceChance{AdvanceID: row.Key, HumanChance: human, AIChance: ai})
					seen[row.Key] = true
				}
			}
		}
		if blocks != 1 {
			return nil, importError(ErrorMalformedValue, file, 1, "ADVANCE_CHANCES", "expected exactly one chance block")
		}
		result = append(result, profile)
		index = next
	}
	if len(result) != len(startingDifficultyIDs) {
		return nil, importError(ErrorMalformedValue, file, 1, "DIFFICULTY", "expected six difficulty blocks")
	}
	return result, nil
}
