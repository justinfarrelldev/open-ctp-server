package engine

import (
	"sort"

	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/rules"
)

// Inspector reuses the game's pinned rules for read-only display calculations.
// It is canonical-side only: callers MUST gate Tile reads by current visibility.
// Remembered display facts belong in FogObservation, never in a new hidden read.
type Inspector struct {
	state       State
	runtime     *rules.RuntimeRules
	coordinates map[[2]int]Tile
}

// NewInspector compiles the same pinned runtime used by command execution once.
func NewInspector(state State) (*Inspector, error) {
	runtime, err := runtimeRulesForState(state)
	if err != nil {
		return nil, err
	}
	coordinates := make(map[[2]int]Tile, len(state.Tiles))
	for _, tile := range state.Tiles {
		coordinates[[2]int{tile.X, tile.Y}] = tile
	}
	return &Inspector{state: state, runtime: runtime, coordinates: coordinates}, nil
}

// TileYield is an inspectable tile's pre-city-policy output. All three values are
// explicit, including zero; nil TileInspection.Yields denotes unavailable data.
type TileYield struct {
	Food       int `json:"food"`
	Production int `json:"production"`
	Gold       int `json:"gold"`
}

// ObservedOverlay contains public art/inspection facts, not hidden effect rules.
type ObservedOverlay struct {
	Kind          string `json:"kind"`
	ImprovementID string `json:"improvement_id,omitempty"`
	Remaining     int    `json:"remaining,omitempty"`
}

// TileInspection is a last-observable snapshot. MovementCost is general terrain
// cost, not unit access, occupancy, domain legality or current enemy ZOC.
type TileInspection struct {
	RiverGeometry *RiverGeometry    `json:"river_geometry,omitempty"`
	Yields        *TileYield        `json:"yields,omitempty"`
	YieldContext  string            `json:"yield_context,omitempty"`
	MovementCost  *int              `json:"movement_cost,omitempty"`
	Pollution     int               `json:"pollution"`
	Overlays      []ObservedOverlay `json:"overlays,omitempty"`
}

// Tile returns only display facts derived from the actual effective engine rules.
// Unsupported yields/costs are absent, never invented zeroes or fresh fog values.
func (i *Inspector) Tile(tileID string) *TileInspection {
	tile, ok := tileByID(i.state, tileID)
	if !ok {
		return nil
	}
	result := &TileInspection{Pollution: tile.Pollution, YieldContext: "terrain", RiverGeometry: i.riverGeometry(tile)}
	yield, err := terrainYieldWithRuntime(i.state, tileID, i.runtime)
	if _, city := cityByTile(i.state, tileID); city {
		result.YieldContext = "city-center"
		yield, err = cityTileYieldWithRuntime(i.state, tileID, i.runtime)
	}
	if err == nil {
		result.Yields = &TileYield{Food: yield.Food, Production: yield.Production, Gold: yield.Gold}
	} else {
		result.YieldContext = ""
	}
	result.MovementCost = i.MovementCost(tileID)
	for _, overlay := range i.state.Overlays {
		if overlay.TileID == tileID {
			result.Overlays = append(result.Overlays, ObservedOverlay{Kind: overlay.Kind, ImprovementID: overlay.ImprovementID, Remaining: overlay.Remaining})
		}
	}
	sort.Slice(result.Overlays, func(a, b int) bool {
		left, right := result.Overlays[a], result.Overlays[b]
		if left.Kind != right.Kind {
			return left.Kind < right.Kind
		}
		if left.ImprovementID != right.ImprovementID {
			return left.ImprovementID < right.ImprovementID
		}
		return left.Remaining < right.Remaining
	})
	return result
}

// MovementCost uses execution's terrain/river/city/improvement calculation. It
// does not certify legality or reveal unit-specific obstacles.
func (i *Inspector) MovementCost(tileID string) *int {
	if cost, ok := terrainCost(i.state, i.runtime, tileID); ok {
		return &cost
	}
	return nil
}

// CityArt contains semantic observed facts. AgeID is the current owner's
// canonical age, not a guessed founder culture or client sprite recipe.
type CityArt struct {
	AgeID   string `json:"age_id,omitempty"`
	Capital bool   `json:"capital"`
}

// CityArt requires the same current-visibility gate as Tile. Stale cities must
// use their stored CityArt observation, including a nil legacy unknown.
func (i *Inspector) CityArt(cityID string) *CityArt {
	city, ok := cityByID(i.state, cityID)
	if !ok {
		return nil
	}
	art := &CityArt{Capital: hasEffect(cityEffects(i.state, *city), EffectCapitol)}
	if nation := nationForID(&i.state, nationForPlayer(i.state, city.OwnerID)); nation != nil {
		art.AgeID = nation.CurrentAgeID
	}
	return art
}

// CalendarDisplay is an authoritative year; localization/BC-AD formatting remain
// client presentation. Absent displays indicate no versioned timescale in the pin.
type CalendarDisplay struct {
	Version string `json:"version"`
	Turn    int    `json:"turn"`
	Year    int64  `json:"year"`
}

func (i *Inspector) Calendar() *CalendarDisplay {
	if i.runtime == nil {
		return nil
	}
	year, ok := i.runtime.Calendar.YearAtTurn(i.state.Phase.Turn)
	if !ok {
		return nil
	}
	return &CalendarDisplay{Version: "ctp2-timescale-v1", Turn: i.state.Phase.Turn, Year: year}
}

// RushBuyQuote is display-only. A cost is not an executable command field, and
// purchasing must still satisfy the service's expected revision and engine checks.
type RushBuyQuote struct {
	Revision int64  `json:"revision"`
	Status   string `json:"status"`
	Cost     *int   `json:"cost,omitempty"`
}

// RushBuy returns owner-only quotes using the exact execution calculation.
func (i *Inspector) RushBuy(cityID, actor string) *RushBuyQuote {
	city, ok := cityByID(i.state, cityID)
	if !ok || city.OwnerID != actor {
		return nil
	}
	nation := nationForID(&i.state, nationForPlayer(i.state, actor))
	if nation == nil {
		return nil
	}
	quote := &RushBuyQuote{Revision: i.state.Revision, Status: "empty-queue"}
	if len(city.Queue) == 0 {
		return quote
	}
	item := city.Queue[0]
	if item.Kind == ProductionCapitalization || item.Kind == ProductionInfrastructure {
		quote.Status = "not-rushable"
		return quote
	}
	cost := rushBuyCost(*city, item, rushBuyTypeCoefficient(i.state, actor, item, i.runtime))
	if cost < 1 && city.StoredProduction < item.Cost {
		// Execution rejects nonpositive prices for unfinished production, e.g.
		// overflow or an unrepresentably small coefficient. Never display a
		// negative purchase price or claim that this item is already funded.
		quote.Status = "unavailable"
		return quote
	}
	quote.Cost = &cost
	quote.Status = "available"
	if cost == 0 {
		quote.Status = "completed"
	} else if nation.Gold < cost {
		quote.Status = "insufficient-gold"
	}
	return quote
}
