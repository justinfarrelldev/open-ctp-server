package world

// OriginalGoodsGeneratorVersion repairs resource catalog lineage while retaining
// V1 terrain/spawns, V2 rivers and V3 ruins. Persisted older versions are unchanged.
const OriginalGoodsGeneratorVersion = "ctp2-original-goods-v4"
const OriginalResourceCatalogVersion = "ctp2-original-resources-v1"

// This bounded semantic registry records original terrain.txt Resources, checked
// against original goods.txt IDs. No sprite IDs/art or installed GOG extensions
// define game placement. In particular shelf/glacier have no resources and
// swamp/tundra only have ONE. Terrain variants need explicit original aliases.
var originalResources = map[string][]string{
	"forest":    {"TERRAIN_FOREST_GOOD_ONE", "TERRAIN_FOREST_GOOD_TWO"},
	"plains":    {"TERRAIN_PLAINS_GOOD_ONE", "TERRAIN_PLAINS_GOOD_TWO"},
	"grassland": {"TERRAIN_GRASSLAND_GOOD_ONE", "TERRAIN_GRASSLAND_GOOD_TWO"},
	"desert":    {"TERRAIN_DESERT_GOOD_ONE", "TERRAIN_DESERT_GOOD_TWO"},
	"mountain":  {"TERRAIN_MOUNTAIN_GOOD_ONE", "TERRAIN_MOUNTAIN_GOOD_TWO"},
	"swamp":     {"TERRAIN_SWAMP_GOOD_ONE"},
	"tundra":    {"TERRAIN_TUNDRA_GOOD_ONE"},
	"ocean":     {"TERRAIN_WATER_DEEP_GOOD_ONE", "TERRAIN_WATER_DEEP_GOOD_TWO"},
}

// ResourceCatalogForGenerator is a pinned generation policy, not a claim about
// the viewer's installed art catalog. Legacy procedural IDs have no such pin.
func ResourceCatalogForGenerator(version string) string {
	if version == OriginalGoodsGeneratorVersion {
		return OriginalResourceCatalogVersion
	}
	return ""
}

// OriginalResourceAllowed validates initial placement only. Terrain changes may
// legitimately retain a deposit (original Cell::SetTerrain doesn't remove goods).
func OriginalResourceAllowed(terrain, good string) bool {
	if good == "" {
		return true
	}
	for _, id := range originalResources[terrain] {
		if id == good {
			return true
		}
	}
	return false
}

func generateOriginalGoods(tiles []WorldTile, rng SplitMix64RNG) {
	for index, tile := range tiles {
		tiles[index].GoodID = ""
		goods := originalResources[tile.Terrain]
		if len(goods) == 0 || coordinateNoise(rng.Split("goods").Seed(), tile.X, tile.Y) >= 1.0/49.0 {
			continue
		}
		variant := int(coordinateNoise(rng.Split("good-kind").Seed(), tile.X, tile.Y) * float64(len(goods)))
		tiles[index].GoodID = goods[variant]
	}
}
