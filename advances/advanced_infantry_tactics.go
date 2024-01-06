package advances

import (
	ages "github.com/justinfarrelldev/open-ctp-server/ages"
)

var (
	cost int32 = 43080
	// TODO hook up to Age 3
	age ages.Age = ages.Age{}
)

var AdvancedInfantryTactics Advance = Advance{
	Prerequisites: []*Advance{
		// TODO
		// Prerequisites ADVANCE_NAVAL_AVIATION
		// Prerequisites ADVANCE_VERTICAL_FLIGHT_AIRCRAFT
	},
	Cost: &cost,
	Age:  &age,
}
