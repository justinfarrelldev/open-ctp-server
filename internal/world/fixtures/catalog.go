package fixtures

// Catalog captures deterministic CTP2-derived world-generation fixture metadata.
//
// These fixtures are derivation inputs for tests and future generator parity work only.
// They intentionally keep the CTP2 reference breadcrumbs in-repo so the runtime server
// never has to read from the external source tree.

type HostParams struct {
	Style            string
	RulesJSON        string
	ExclusionsJSON   string
	StartingSettlers int
	StartingGold     int
	Seed             int64
	Wetness          int
	MapSize          string
	Climate          string
}

type MapParameters struct {
	WrapX           bool
	WrapY           bool
	PlayerCount     int
	SetupRadius     int
	OceanLand       int
	IslandContinent int
	Homogeneity     int
	GoodCount       int
}

type ExpectedInvariants struct {
	SpawnCountMatchesPlayerOrder bool
	UniquePlayerStarts           bool
	RequiresLandContinents       bool
	TerrainBias                  string
	StartDistanceRule            string
	Notes                        []string
}

type SourceReference struct {
	File   string
	Lines  string
	Reason string
}

type Definition struct {
	Name             string
	GeneratorVersion string
	HostParams       HostParams
	MapParams        MapParameters
	PlayerOrder      []int64
	Expected         ExpectedInvariants
	SourceReferences []SourceReference
}

const GeneratorVersion = "ctp2-reference-fixtures-v1"

var Inventory = []Definition{
	{
		Name:             "small-default",
		GeneratorVersion: GeneratorVersion,
		HostParams: HostParams{
			Style:            "standard_ctp2",
			RulesJSON:        "[]",
			ExclusionsJSON:   "{}",
			StartingSettlers: 1,
			StartingGold:     100,
			Seed:             41041,
			Wetness:          50,
			MapSize:          "small",
			Climate:          "temperate",
		},
		MapParams: MapParameters{
			WrapX:           true,
			WrapY:           false,
			PlayerCount:     3,
			SetupRadius:     2,
			OceanLand:       50,
			IslandContinent: 50,
			Homogeneity:     50,
			GoodCount:       50,
		},
		PlayerOrder: []int64{101, 202, 303},
		Expected: ExpectedInvariants{
			SpawnCountMatchesPlayerOrder: true,
			UniquePlayerStarts:           true,
			RequiresLandContinents:       true,
			TerrainBias:                  "balanced moisture bands with mixed plains, forest, and desert thresholds",
			StartDistanceRule:            "setup radius 2 drives FindPlayerStart radius d=3 with unique land starts",
			Notes: []string{
				"Baseline small-map fixture for future byte-stable canonical world snapshots.",
				"Acts as the default parity case before wet and dry variants shift moisture thresholds.",
			},
		},
		SourceReferences: []SourceReference{
			{File: "ctp2_code/gs/utility/gameinit.cpp", Lines: "1111-1125", Reason: "custom map sliders, map size, wrap flags, and CreateTheWorld bootstrap all flow from profile settings here"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "382-605", Reason: "continent, land, homogeneity, wetness, and temperature thresholds control terrain bands"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "1777-1853", Reason: "player-count and setup-radius inputs drive start scoring and minimum distance expectations"},
			{File: "ctp2_code/gs/world/WrldCont.cpp", Lines: "94-113", Reason: "continents are numbered before spawn selection, so later generator parity must preserve continent-based start expectations"},
		},
	},
	{
		Name:             "small-wet",
		GeneratorVersion: GeneratorVersion,
		HostParams: HostParams{
			Style:            "standard_ctp2",
			RulesJSON:        "[]",
			ExclusionsJSON:   "{}",
			StartingSettlers: 1,
			StartingGold:     100,
			Seed:             52052,
			Wetness:          80,
			MapSize:          "small",
			Climate:          "temperate",
		},
		MapParams: MapParameters{
			WrapX:           true,
			WrapY:           false,
			PlayerCount:     3,
			SetupRadius:     2,
			OceanLand:       50,
			IslandContinent: 50,
			Homogeneity:     70,
			GoodCount:       50,
		},
		PlayerOrder: []int64{101, 202, 303},
		Expected: ExpectedInvariants{
			SpawnCountMatchesPlayerOrder: true,
			UniquePlayerStarts:           true,
			RequiresLandContinents:       true,
			TerrainBias:                  "wet-biased moisture histogram should expand swamp, forest, kelp, and grassland outcomes relative to the dry fixture",
			StartDistanceRule:            "setup radius 2 still requires separated land starts even after wetter terrain conversion",
			Notes: []string{
				"Wetness materially changes the second-pass wet histogram expectations in wldgen.cpp.",
				"Future generator parity should prove the wet fixture produces fewer desert-band tiles than the dry fixture for the same map size tier.",
			},
		},
		SourceReferences: []SourceReference{
			{File: "ctp2_code/gs/utility/gameinit.cpp", Lines: "1111-1119", Reason: "GetWetDry and GetMapSize are captured before world construction"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "483-566", Reason: "wetmap histogram thresholds transform plains into swamp, forest, grassland, plains, or desert bands"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "1777-1853", Reason: "spawn scoring still depends on player count and setup radius after terrain generation"},
			{File: "ctp2_code/gs/world/WrldCont.cpp", Lines: "134-199", Reason: "ocean and land growth produce continent numbering used by start validation"},
		},
	},
	{
		Name:             "small-dry",
		GeneratorVersion: GeneratorVersion,
		HostParams: HostParams{
			Style:            "standard_ctp2",
			RulesJSON:        "[]",
			ExclusionsJSON:   "{}",
			StartingSettlers: 1,
			StartingGold:     100,
			Seed:             63063,
			Wetness:          20,
			MapSize:          "small",
			Climate:          "temperate",
		},
		MapParams: MapParameters{
			WrapX:           true,
			WrapY:           false,
			PlayerCount:     3,
			SetupRadius:     2,
			OceanLand:       50,
			IslandContinent: 50,
			Homogeneity:     30,
			GoodCount:       50,
		},
		PlayerOrder: []int64{101, 202, 303},
		Expected: ExpectedInvariants{
			SpawnCountMatchesPlayerOrder: true,
			UniquePlayerStarts:           true,
			RequiresLandContinents:       true,
			TerrainBias:                  "dry-biased moisture histogram should widen desert and plains outcomes while reducing swamp and kelp conversions",
			StartDistanceRule:            "setup radius 2 still requires separated land starts even when the map trends drier",
			Notes: []string{
				"Dry fixture is the counterpart to the wet fixture and should materially diverge in moisture-sensitive terrain bands.",
				"Future generator parity should prove the dry fixture yields more desert-band tiles than the wet fixture for the same size tier.",
			},
		},
		SourceReferences: []SourceReference{
			{File: "ctp2_code/gs/utility/gameinit.cpp", Lines: "1111-1119", Reason: "GetWetDry and GetMapSize are captured before world construction"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "483-566", Reason: "wetmap histogram thresholds are the reason this fixture diverges from the wet case"},
			{File: "ctp2_code/gs/world/wldgen.cpp", Lines: "1777-1853", Reason: "spawn scoring still depends on player count and setup radius after terrain generation"},
			{File: "ctp2_code/gs/world/WrldCont.cpp", Lines: "94-113", Reason: "continent numbering remains a prerequisite for deterministic start validation"},
		},
	},
}

func ByName(name string) (Definition, bool) {
	for _, definition := range Inventory {
		if definition.Name == name {
			return definition, true
		}
	}

	return Definition{}, false
}
