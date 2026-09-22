package fullgame

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/visibility"
)

type Script struct {
	Version  int             `json:"version"`
	Commands []ScriptCommand `json:"commands"`
}

type ScriptCommand struct {
	Command   string          `json:"command"`
	ActorRole string          `json:"actor_role"`
	Payload   json.RawMessage `json:"payload"`
}

func ResolveScript(script Script, publicProjection []byte, actors map[string]string) ([]ScriptCommand, error) {
	if script.Version != 1 || len(script.Commands) == 0 {
		return nil, fmt.Errorf("script is invalid")
	}
	var projection visibility.Projection
	if err := json.Unmarshal(publicProjection, &projection); err != nil {
		return nil, fmt.Errorf("decode public projection: %w", err)
	}
	roles := make(map[string]string)
	for role, playerID := range actors {
		if !isHumanScriptActor(role) || playerID == "" {
			return nil, fmt.Errorf("script actor %q is not a human-controlled role", role)
		}
		units := make([]string, 0)
		for _, unit := range projection.Units {
			if unit.OwnerID == playerID {
				units = append(units, unit.ID)
			}
		}
		sort.Strings(units)
		for index, id := range units {
			roles[fmt.Sprintf("%s.unit-%d", role, index+1)] = id
		}
		roles[fmt.Sprintf("%s.army-1", role)] = publicScriptArmyID(playerID)
		tiles := make([]string, 0, len(projection.Tiles))
		for _, tile := range projection.Tiles {
			if !tile.Stale {
				tiles = append(tiles, tile.ID)
			}
		}
		sort.Strings(tiles)
		for index, id := range tiles {
			roles[fmt.Sprintf("%s.tile-%d", role, index+1)] = id
		}
		cities := make([]string, 0)
		for _, city := range projection.Cities {
			if city.OwnerID != playerID && !city.Stale {
				cities = append(cities, city.ID)
			}
		}
		sort.Strings(cities)
		for index, id := range cities {
			roles[fmt.Sprintf("%s.target-city-%d", role, index+1)] = id
		}
	}
	resolved := make([]ScriptCommand, len(script.Commands))
	for index, command := range script.Commands {
		if command.Command == "" || !isHumanScriptActor(command.ActorRole) || actors[command.ActorRole] == "" || !json.Valid(command.Payload) {
			return nil, fmt.Errorf("script command %d must target a human-controlled actor", index)
		}
		var payload any
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return nil, err
		}
		value, err := resolveRoles(payload, roles)
		if err != nil {
			return nil, err
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		resolved[index] = ScriptCommand{Command: command.Command, ActorRole: command.ActorRole, Payload: encoded}
	}
	return resolved, nil
}

func isHumanScriptActor(role string) bool {
	return strings.HasPrefix(role, "human-") && len(strings.TrimPrefix(role, "human-")) > 0
}

func publicScriptArmyID(playerID string) string {
	if suffix := strings.TrimPrefix(playerID, "account-"); suffix != "" && suffix != playerID {
		return fmt.Sprintf("fullgame-army-human-%s", suffix)
	}
	return fmt.Sprintf("fullgame-army-%s", playerID)
}


func resolveRoles(value any, roles map[string]string) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		if role, ok := typed["role"].(string); ok && len(typed) == 1 {
			id := roles[role]
			if id == "" {
				return nil, fmt.Errorf("public role %q is unavailable", role)
			}
			return id, nil
		}
		for key, child := range typed {
			resolved, err := resolveRoles(child, roles)
			if err != nil {
				return nil, err
			}
			typed[key] = resolved
		}
	case []any:
		for index, child := range typed {
			resolved, err := resolveRoles(child, roles)
			if err != nil {
				return nil, err
			}
			typed[index] = resolved
		}
	}
	return value, nil
}
