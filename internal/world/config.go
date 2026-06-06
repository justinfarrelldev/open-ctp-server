package world

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// WorldConfig is the normalized host-generation configuration used by the pure world harness.
type WorldConfig struct {
	Style            string `json:"style,omitempty"`
	RulesJSON        string `json:"rules_json,omitempty"`
	ExclusionsJSON   string `json:"exclusions_json,omitempty"`
	StartingSettlers int    `json:"starting_settlers"`
	StartingGold     int    `json:"starting_gold"`
	Seed             *int64 `json:"seed,omitempty"`
	Wetness          *int   `json:"wetness,omitempty"`
	MapSize          string `json:"map_size,omitempty"`
	Climate          string `json:"climate,omitempty"`
	GeneratorVersion string `json:"generator_version,omitempty"`
}

// Clone returns a deep copy of the config so callers can mutate the copy safely.
func (c WorldConfig) Clone() WorldConfig {
	clone := c
	if c.Seed != nil {
		seed := *c.Seed
		clone.Seed = &seed
	}
	if c.Wetness != nil {
		wetness := *c.Wetness
		clone.Wetness = &wetness
	}
	return clone
}

// NormalizeWorldConfig trims string fields and compacts embedded JSON text so identical
// inputs produce byte-identical serialized snapshots later.
func NormalizeWorldConfig(config WorldConfig) (WorldConfig, error) {
	normalized := config.Clone()
	normalized.Style = strings.TrimSpace(normalized.Style)
	normalized.MapSize = strings.TrimSpace(normalized.MapSize)
	normalized.Climate = strings.TrimSpace(normalized.Climate)
	normalized.GeneratorVersion = strings.TrimSpace(normalized.GeneratorVersion)

	if normalized.StartingSettlers < 0 {
		return WorldConfig{}, errors.New("starting_settlers must be non-negative")
	}
	if normalized.StartingGold < 0 {
		return WorldConfig{}, errors.New("starting_gold must be non-negative")
	}

	var err error
	normalized.RulesJSON, err = compactJSONString(normalized.RulesJSON)
	if err != nil {
		return WorldConfig{}, fmt.Errorf("rules_json: %w", err)
	}
	normalized.ExclusionsJSON, err = compactJSONString(normalized.ExclusionsJSON)
	if err != nil {
		return WorldConfig{}, fmt.Errorf("exclusions_json: %w", err)
	}

	return normalized, nil
}

// WorldConfigFromHostParams converts the current contract-layer host params into the pure
// normalized config shape without adding any DB or HTTP dependencies.
func WorldConfigFromHostParams(params GenerationHostParams) WorldConfig {
	return WorldConfig{
		Style:            params.Style,
		RulesJSON:        params.RulesJSON,
		ExclusionsJSON:   params.ExclusionsJSON,
		StartingSettlers: params.StartingSettlers,
		StartingGold:     params.StartingGold,
		Seed:             params.Seed,
		Wetness:          params.Wetness,
		MapSize:          params.MapSize,
		Climate:          params.Climate,
		GeneratorVersion: params.GeneratorVersion,
	}
}

// ToHostParams converts the normalized config back into the current contract shape.
func (c WorldConfig) ToHostParams() GenerationHostParams {
	clone := c.Clone()
	return GenerationHostParams{
		Style:            clone.Style,
		RulesJSON:        clone.RulesJSON,
		ExclusionsJSON:   clone.ExclusionsJSON,
		StartingSettlers: clone.StartingSettlers,
		StartingGold:     clone.StartingGold,
		Seed:             clone.Seed,
		Wetness:          clone.Wetness,
		MapSize:          clone.MapSize,
		Climate:          clone.Climate,
		GeneratorVersion: clone.GeneratorVersion,
	}
}

func compactJSONString(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	var compact bytes.Buffer
	if err := json.Compact(&compact, []byte(trimmed)); err != nil {
		return "", err
	}
	return compact.String(), nil
}
