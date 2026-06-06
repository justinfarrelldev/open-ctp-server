package world

import (
	fixtures "github.com/justinfarrelldev/open-ctp-server/internal/world/fixtures"
)

// GenerateWorldFromFixture turns deterministic T3 fixture inventory into canonical world output.
func GenerateWorldFromFixture(definition fixtures.Definition) (WorldReadResponse, error) {
	seed := definition.HostParams.Seed
	wetness := definition.HostParams.Wetness
	snapshot, _, err := GenerateWorld(GenerationInput{
		GameID:           1,
		PlayerOrder:      append([]int64(nil), definition.PlayerOrder...),
		GeneratorVersion: definition.GeneratorVersion,
		HostParams: GenerationHostParams{
			Style:            definition.HostParams.Style,
			RulesJSON:        definition.HostParams.RulesJSON,
			ExclusionsJSON:   definition.HostParams.ExclusionsJSON,
			StartingSettlers: definition.HostParams.StartingSettlers,
			StartingGold:     definition.HostParams.StartingGold,
			Seed:             &seed,
			Wetness:          &wetness,
			MapSize:          definition.HostParams.MapSize,
			Climate:          definition.HostParams.Climate,
			GeneratorVersion: definition.GeneratorVersion,
		},
	})
	if err != nil {
		return WorldReadResponse{}, err
	}

	return snapshotToReadResponse(snapshot), nil
}
