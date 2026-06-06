package world

import (
	"errors"

	fixtures "github.com/justinfarrelldev/open-ctp-server/internal/world/fixtures"
)

var ErrWorldFixtureGenerationUnimplemented = errors.New("world fixture generation is not implemented yet")

// GenerateWorldFromFixture is the future generator contract that will turn the
// deterministic T3 fixture inventory into canonical world output.
func GenerateWorldFromFixture(definition fixtures.Definition) (WorldReadResponse, error) {
	_ = definition
	return WorldReadResponse{}, ErrWorldFixtureGenerationUnimplemented
}
