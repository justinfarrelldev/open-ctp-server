package world

// RuinGeneratorVersion adds ruins without changing terrain, rivers, goods or
// spawns in older games. Rolls belong to canonical gameplay state, not this map.
const RuinGeneratorVersion = "ctp2-ruins-v3"

// GenerateGoodyHuts in wldgen.cpp and Const.txt use 6x6 boxes with one random
// coordinate per box and chance 1.0. Water selections stay empty (no retry).
func generateRuins(tiles []WorldTile, spawns []PlayerSpawn, profile mapProfile, rng SplitMix64RNG) {
	var stream DeterministicRNG = rng
	for x := 0; x < profile.width/6; x++ {
		for y := 0; y < profile.height/6; y++ {
			var dx, dy int
			stream, dx = stream.Intn(6)
			stream, dy = stream.Intn(6)
			tx, ty := x*6+dx, y*6+dy
			blocked := false
			for _, spawn := range spawns {
				if spawn.X == tx && spawn.Y == ty {
					blocked = true
					break
				}
			}
			if blocked {
				continue
			}
			for index, tile := range tiles {
				if tile.X == tx && tile.Y == ty && isLandTerrain(tile.Terrain) {
					tiles[index].Ruin = true
					break
				}
			}
		}
	}
}
