package world

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
)

const (
	defaultRulesJSON          = "[]"
	defaultExclusionsJSON     = "{}"
	defaultGameStatus         = "started"
	defaultFixtureCountryName = "Civilization"
	maxSpawnPlacementAttempts = 64
)

var (
	ErrWorldGenerationInvalidInput = errors.New("world generation received invalid input")
	ErrWorldGenerationRetryCap     = errors.New("world generation exceeded deterministic spawn attempt cap")
)

type mapProfile struct {
	width           int
	height          int
	wrapX           bool
	wrapY           bool
	baseWaterTarget int
	baseSetupRadius int
}

type climateProfile struct {
	temperatureShift float64
	mountainBias     float64
}

type generatedWorld struct {
	snapshot WorldSnapshot
	turn     MutableTurnState
}

type spawnCandidate struct {
	x         int
	y         int
	score     int
	continent int
}

var supportedMapProfiles = map[string]mapProfile{
	"duel":   {width: 20, height: 12, wrapX: true, wrapY: false, baseWaterTarget: 48, baseSetupRadius: 2},
	"small":  {width: 24, height: 16, wrapX: true, wrapY: false, baseWaterTarget: 50, baseSetupRadius: 2},
	"medium": {width: 32, height: 20, wrapX: true, wrapY: false, baseWaterTarget: 52, baseSetupRadius: 3},
	"large":  {width: 40, height: 24, wrapX: true, wrapY: false, baseWaterTarget: 54, baseSetupRadius: 3},
	"huge":   {width: 48, height: 28, wrapX: true, wrapY: false, baseWaterTarget: 56, baseSetupRadius: 4},
}

var supportedClimates = map[string]climateProfile{
	"arid":      {temperatureShift: 0.10, mountainBias: 0.03},
	"cold":      {temperatureShift: -0.18, mountainBias: 0.08},
	"temperate": {temperatureShift: 0.00, mountainBias: 0.05},
	"tropical":  {temperatureShift: 0.14, mountainBias: 0.02},
}

// GenerateWorld deterministically creates the immutable world snapshot and initial mutable turn state.
func GenerateWorld(input GenerationInput) (WorldSnapshot, MutableTurnState, error) {
	world, err := generateWorld(input)
	if err != nil {
		return WorldSnapshot{}, MutableTurnState{}, err
	}
	return world.snapshot.Clone(), world.turn.Clone(), nil
}

func generateWorld(input GenerationInput) (generatedWorld, error) {
	validated, profile, climate, err := validateGenerationInput(input)
	if err != nil {
		return generatedWorld{}, err
	}

	masterSeed, err := deriveGeneratorSeed(validated)
	if err != nil {
		return generatedWorld{}, fmt.Errorf("%w: %v", ErrWorldGenerationInvalidInput, err)
	}

	rng := NewDeterministicRNG(masterSeed)
	tiles, continents, nextRNG := generateTiles(validated, profile, climate, rng)
	spawns, finalRNG, err := generateSpawns(validated, profile, tiles, continents, nextRNG)
	if err != nil {
		return generatedWorld{}, err
	}

	turn := MutableTurnState{
		TurnNumber:         1,
		CurrentPlayerIndex: 0,
		ActivePlayerOrder:  append([]int64(nil), validated.PlayerOrder...),
		RecalculationCount: 0,
		RNGState:           uint64(finalRNG.Seed()),
		Ratings:            WorldRatings{},
		Pollution:          WorldPollution{},
	}

	snapshot := WorldSnapshot{
		GameID:           validated.GameID,
		GameStatus:       defaultGameStatus,
		GeneratorVersion: validated.GeneratorVersion,
		Config:           WorldConfigFromHostParams(validated.HostParams),
		PlayerOrder:      append([]int64(nil), validated.PlayerOrder...),
		Spawns:           spawns,
		Tiles:            tiles,
		Turn:             turn,
	}

	return generatedWorld{snapshot: snapshot, turn: turn}, nil
}

func validateGenerationInput(input GenerationInput) (GenerationInput, mapProfile, climateProfile, error) {
	if input.GameID <= 0 {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: game_id must be greater than zero", ErrWorldGenerationInvalidInput)
	}
	if strings.TrimSpace(input.GeneratorVersion) == "" {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: generator_version must be specified", ErrWorldGenerationInvalidInput)
	}
	if len(input.PlayerOrder) == 0 {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: player_order must include at least one player", ErrWorldGenerationInvalidInput)
	}

	seen := make(map[int64]struct{}, len(input.PlayerOrder))
	for _, accountID := range input.PlayerOrder {
		if accountID <= 0 {
			return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: player_order contains invalid account_id %d", ErrWorldGenerationInvalidInput, accountID)
		}
		if _, ok := seen[accountID]; ok {
			return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: player_order contains duplicate account_id %d", ErrWorldGenerationInvalidInput, accountID)
		}
		seen[accountID] = struct{}{}
	}

	normalizedConfig, err := NormalizeWorldConfig(WorldConfigFromHostParams(input.HostParams))
	if err != nil {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: %v", ErrWorldGenerationInvalidInput, err)
	}
	if normalizedConfig.Seed == nil {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: seed must be specified", ErrWorldGenerationInvalidInput)
	}
	if normalizedConfig.Wetness == nil {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: wetness must be specified", ErrWorldGenerationInvalidInput)
	}
	if *normalizedConfig.Wetness < 0 || *normalizedConfig.Wetness > 100 {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: wetness must be between 0 and 100", ErrWorldGenerationInvalidInput)
	}
	if normalizedConfig.Style == "" {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: style must be specified", ErrWorldGenerationInvalidInput)
	}
	if normalizedConfig.MapSize == "" {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: map_size must be specified", ErrWorldGenerationInvalidInput)
	}
	if normalizedConfig.Climate == "" {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: climate must be specified", ErrWorldGenerationInvalidInput)
	}

	profile, ok := supportedMapProfiles[strings.ToLower(normalizedConfig.MapSize)]
	if !ok {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: unsupported map_size %q", ErrWorldGenerationInvalidInput, normalizedConfig.MapSize)
	}
	climate, ok := supportedClimates[strings.ToLower(normalizedConfig.Climate)]
	if !ok {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: unsupported climate %q", ErrWorldGenerationInvalidInput, normalizedConfig.Climate)
	}

	normalizedConfig.RulesJSON = defaultIfEmpty(normalizedConfig.RulesJSON, defaultRulesJSON)
	normalizedConfig.ExclusionsJSON = defaultIfEmpty(normalizedConfig.ExclusionsJSON, defaultExclusionsJSON)
	normalizedConfig.GeneratorVersion = input.GeneratorVersion

	validated := GenerationInput{
		GameID:           input.GameID,
		PlayerOrder:      append([]int64(nil), input.PlayerOrder...),
		GeneratorVersion: strings.TrimSpace(input.GeneratorVersion),
		HostParams:       normalizedConfig.ToHostParams(),
	}

	minimumArea := len(validated.PlayerOrder) * (profile.baseSetupRadius + 1) * 2
	if profile.width*profile.height < minimumArea {
		return GenerationInput{}, mapProfile{}, climateProfile{}, fmt.Errorf("%w: map size %q cannot fit %d players", ErrWorldGenerationInvalidInput, normalizedConfig.MapSize, len(validated.PlayerOrder))
	}

	return validated, profile, climate, nil
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func deriveGeneratorSeed(input GenerationInput) (int64, error) {
	key, err := CanonicalJSON(struct {
		Seed             int64   `json:"seed"`
		Wetness          int     `json:"wetness"`
		MapSize          string  `json:"map_size"`
		Climate          string  `json:"climate"`
		Style            string  `json:"style"`
		RulesJSON        string  `json:"rules_json"`
		ExclusionsJSON   string  `json:"exclusions_json"`
		StartingSettlers int     `json:"starting_settlers"`
		StartingGold     int     `json:"starting_gold"`
		GeneratorVersion string  `json:"generator_version"`
		PlayerOrder      []int64 `json:"player_order"`
	}{
		Seed:             *input.HostParams.Seed,
		Wetness:          *input.HostParams.Wetness,
		MapSize:          input.HostParams.MapSize,
		Climate:          input.HostParams.Climate,
		Style:            input.HostParams.Style,
		RulesJSON:        input.HostParams.RulesJSON,
		ExclusionsJSON:   input.HostParams.ExclusionsJSON,
		StartingSettlers: input.HostParams.StartingSettlers,
		StartingGold:     input.HostParams.StartingGold,
		GeneratorVersion: input.GeneratorVersion,
		PlayerOrder:      input.PlayerOrder,
	})
	if err != nil {
		return 0, err
	}

	hasher := fnv.New64a()
	_, _ = hasher.Write(key)
	var mixed [8]byte
	binary.LittleEndian.PutUint64(mixed[:], hasher.Sum64())
	seed := int64(binary.LittleEndian.Uint64(mixed[:]) ^ uint64(*input.HostParams.Seed))
	if seed == 0 {
		seed = 1
	}
	return seed, nil
}

func generateTiles(input GenerationInput, profile mapProfile, climate climateProfile, rng SplitMix64RNG) ([]WorldTile, [][]int, SplitMix64RNG) {
	width := profile.width
	height := profile.height
	wetness := *input.HostParams.Wetness
	waterTarget := profile.baseWaterTarget + climateWaterAdjustment(input.HostParams.Climate) - (wetness-50)/12
	if waterTarget < 35 {
		waterTarget = 35
	}
	if waterTarget > 72 {
		waterTarget = 72
	}

	tiles := make([]WorldTile, 0, width*height)
	continents := make([][]int, height)
	for y := 0; y < height; y++ {
		continents[y] = make([]int, width)
	}

	terrainRNG := rng.Split("terrain")
	continentID := 0
	for y := 0; y < height; y++ {
		latitude := normalizeLatitude(y, height)
		for x := 0; x < width; x++ {
			positionNoise := coordinateNoise(terrainRNG.Seed(), x, y)
			moisture := ClampFloat64(float64(wetness)/100 + positionNoise*0.50 - 0.25)
			heat := ClampFloat64(1 - latitude + climate.temperatureShift + positionNoise*0.20 - 0.10)
			altitude := ClampFloat64(positionNoise*0.70 + latitude*0.30)
			terrain := classifyTerrain(moisture, heat, altitude, waterTarget, climate)
			if terrain != "ocean" && terrain != "shelf" {
				continentID++
				continents[y][x] = continentID
			}
			tiles = append(tiles, WorldTile{X: x, Y: y, Terrain: terrain})
		}
	}

	mergeContinents(tiles, continents, profile)
	return tiles, continents, terrainRNG.Split("post-terrain")
}

func normalizeLatitude(y, height int) float64 {
	if height <= 1 {
		return 0
	}
	midpoint := float64(height-1) / 2
	return absFloat(float64(y)-midpoint) / midpoint
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}

func climateWaterAdjustment(climate string) int {
	switch strings.ToLower(climate) {
	case "tropical":
		return -4
	case "arid":
		return 6
	case "cold":
		return 2
	default:
		return 0
	}
}

func coordinateNoise(seed int64, x, y int) float64 {
	hasher := fnv.New64a()
	var buffer [24]byte
	binary.LittleEndian.PutUint64(buffer[0:8], uint64(seed))
	binary.LittleEndian.PutUint64(buffer[8:16], uint64(x+1))
	binary.LittleEndian.PutUint64(buffer[16:24], uint64(y+1))
	_, _ = hasher.Write(buffer[:])
	return float64(hasher.Sum64()%10000) / 10000.0
}

func classifyTerrain(moisture, heat, altitude float64, waterTarget int, climate climateProfile) string {
	waterThreshold := float64(waterTarget) / 100
	if altitude < waterThreshold*0.48 {
		if altitude < waterThreshold*0.34 {
			return "ocean"
		}
		return "shelf"
	}
	if altitude > 0.88-climate.mountainBias {
		return "mountain"
	}
	if heat < 0.18 {
		if moisture > 0.48 {
			return "tundra"
		}
		return "desert"
	}
	if moisture > 0.82 {
		if heat > 0.52 {
			return "swamp"
		}
		return "forest"
	}
	if moisture > 0.62 {
		return "forest"
	}
	if moisture < 0.20 {
		return "desert"
	}
	if heat > 0.72 && moisture > 0.46 {
		return "grassland"
	}
	return "plains"
}

func mergeContinents(tiles []WorldTile, continents [][]int, profile mapProfile) {
	width := profile.width
	height := profile.height
	terrainByCoordinate := make(map[[2]int]string, len(tiles))
	for _, tile := range tiles {
		terrainByCoordinate[[2]int{tile.X, tile.Y}] = tile.Terrain
	}

	visited := make([][]bool, height)
	for y := 0; y < height; y++ {
		visited[y] = make([]bool, width)
	}

	continentID := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if visited[y][x] || !isLandTerrain(terrainByCoordinate[[2]int{x, y}]) {
				continue
			}
			continentID++
			floodFillContinent(x, y, continentID, terrainByCoordinate, continents, visited, profile)
		}
	}
}

func generateSpawns(input GenerationInput, profile mapProfile, tiles []WorldTile, continents [][]int, rng SplitMix64RNG) ([]PlayerSpawn, SplitMix64RNG, error) {
	candidates := make([]spawnCandidate, 0, len(tiles))
	for _, tile := range tiles {
		if !isSpawnableTerrain(tile.Terrain) {
			continue
		}
		continent := continents[tile.Y][tile.X]
		if continent == 0 {
			continue
		}
		candidates = append(candidates, spawnCandidate{
			x:         tile.X,
			y:         tile.Y,
			continent: continent,
			score:     scoreSpawnTile(tile),
		})
	}
	if len(candidates) < len(input.PlayerOrder) {
		return nil, rng, fmt.Errorf("%w: placed 0 of %d spawns after %d attempts", ErrWorldGenerationRetryCap, len(input.PlayerOrder), maxSpawnPlacementAttempts)
	}

	selectionSeed := rng.Split("spawns").Seed()
	sort.SliceStable(candidates, func(i, j int) bool {
		iNoise := coordinateNoise(selectionSeed, candidates[i].x, candidates[i].y)
		jNoise := coordinateNoise(selectionSeed, candidates[j].x, candidates[j].y)
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		if iNoise != jNoise {
			return iNoise > jNoise
		}
		if candidates[i].y != candidates[j].y {
			return candidates[i].y < candidates[j].y
		}
		return candidates[i].x < candidates[j].x
	})

	minDistance := profile.baseSetupRadius + 1
	stride := deterministicSpawnStride(len(candidates), selectionSeed)
	bestSelected := make([]spawnCandidate, 0)
	for attempt := 0; attempt < maxSpawnPlacementAttempts; attempt++ {
		startIndex := 0
		if len(candidates) > 0 {
			startIndex = (attempt * stride) % len(candidates)
		}
		selected := greedySelectSpawns(candidates, startIndex, minDistance, profile, len(input.PlayerOrder))
		if len(selected) > len(bestSelected) {
			bestSelected = append(bestSelected[:0], selected...)
		}
		if len(selected) == len(input.PlayerOrder) {
			spawns := make([]PlayerSpawn, 0, len(selected))
			for i, selectedCandidate := range selected {
				spawns = append(spawns, PlayerSpawn{
					AccountID: input.PlayerOrder[i],
					Country:   fmt.Sprintf("%s %d", defaultFixtureCountryName, i+1),
					X:         selectedCandidate.x,
					Y:         selectedCandidate.y,
				})
			}
			return spawns, rng.Split("spawn-state"), nil
		}
	}

	return nil, rng, fmt.Errorf("%w: placed %d of %d spawns after %d attempts", ErrWorldGenerationRetryCap, len(bestSelected), len(input.PlayerOrder), maxSpawnPlacementAttempts)
}

func isSpawnableTerrain(terrain string) bool {
	switch terrain {
	case "plains", "grassland", "forest", "swamp", "tundra":
		return true
	default:
		return false
	}
}

func isLandTerrain(terrain string) bool {
	switch terrain {
	case "ocean", "shelf", "":
		return false
	default:
		return true
	}
}

func scoreSpawnTile(tile WorldTile) int {
	switch tile.Terrain {
	case "grassland":
		return 5
	case "forest":
		return 4
	case "plains":
		return 3
	case "swamp":
		return 2
	case "tundra":
		return 1
	default:
		return 0
	}
}

func isSpawnFarEnough(candidate spawnCandidate, selected []spawnCandidate, minDistance int, profile mapProfile) bool {
	for _, existing := range selected {
		if distance(candidate.x, candidate.y, existing.x, existing.y, profile) < minDistance {
			return false
		}
	}
	return true
}

func floodFillContinent(startX, startY, continentID int, terrainByCoordinate map[[2]int]string, continents [][]int, visited [][]bool, profile mapProfile) {
	queue := [][2]int{{startX, startY}}
	visited[startY][startX] = true
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		x, y := current[0], current[1]
		continents[y][x] = continentID
		for _, neighbor := range continentNeighbors(x, y, profile) {
			nx, ny := neighbor[0], neighbor[1]
			if visited[ny][nx] || !isLandTerrain(terrainByCoordinate[[2]int{nx, ny}]) {
				continue
			}
			visited[ny][nx] = true
			queue = append(queue, [2]int{nx, ny})
		}
	}
}

func continentNeighbors(x, y int, profile mapProfile) [][2]int {
	neighbors := make([][2]int, 0, 4)
	candidateOffsets := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, offset := range candidateOffsets {
		nx := x + offset[0]
		ny := y + offset[1]
		if profile.wrapX {
			for nx < 0 {
				nx += profile.width
			}
			nx %= profile.width
		}
		if profile.wrapY {
			for ny < 0 {
				ny += profile.height
			}
			ny %= profile.height
		}
		if nx < 0 || nx >= profile.width || ny < 0 || ny >= profile.height {
			continue
		}
		neighbors = append(neighbors, [2]int{nx, ny})
	}
	return neighbors
}

func greedySelectSpawns(candidates []spawnCandidate, startIndex, minDistance int, profile mapProfile, targetCount int) []spawnCandidate {
	selected := make([]spawnCandidate, 0, targetCount)
	for i := 0; i < len(candidates); i++ {
		candidate := candidates[(startIndex+i)%len(candidates)]
		if isSpawnFarEnough(candidate, selected, minDistance, profile) {
			selected = append(selected, candidate)
			if len(selected) == targetCount {
				return selected
			}
		}
	}
	return selected
}

func deterministicSpawnStride(candidateCount int, selectionSeed int64) int {
	if candidateCount <= 1 {
		return 1
	}
	stride := int(selectionSeed%int64(candidateCount-1)) + 1
	for gcd(stride, candidateCount) != 1 {
		stride++
		if stride >= candidateCount {
			stride = 1
		}
		if stride == 1 {
			break
		}
	}
	return stride
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

func distance(ax, ay, bx, by int, profile mapProfile) int {
	dx := absInt(ax - bx)
	if profile.wrapX {
		wrapped := profile.width - dx
		if wrapped < dx {
			dx = wrapped
		}
	}
	dy := absInt(ay - by)
	return dx + dy
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func snapshotToReadResponse(snapshot WorldSnapshot) WorldReadResponse {
	return WorldReadResponse{
		GameID:           snapshot.GameID,
		GameStatus:       snapshot.GameStatus,
		GeneratorVersion: snapshot.GeneratorVersion,
		HostParams:       snapshot.Config.ToHostParams(),
		PlayerOrder:      append([]int64(nil), snapshot.PlayerOrder...),
		Spawns:           append([]PlayerSpawn(nil), snapshot.Spawns...),
		Tiles:            append([]WorldTile(nil), snapshot.Tiles...),
	}
}
