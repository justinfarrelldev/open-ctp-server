package engine

// Forecast is a constant-current-economy display estimate, not a promised future
// outcome. Growth, losses, future orders and other players' decisions may change
// it. A nil turns value is NOT zero: inspect Status for stalled/unavailable cases.
type Forecast struct {
	Revision   int64  `json:"revision"`
	Status     string `json:"status"`
	Rate       *int   `json:"rate,omitempty"`
	Turns      *int   `json:"turns,omitempty"`
	Assumption string `json:"assumption"`
}

func newForecast(revision int64) *Forecast {
	return &Forecast{Revision: revision, Status: "unavailable", Assumption: "constant-current-economy"}
}

func finishForecast(result *Forecast, remaining, rate int) *Forecast {
	result.Rate = &rate
	if remaining <= 0 {
		turns := 0
		result.Turns = &turns
		result.Status = "completed"
		return result
	}
	if rate <= 0 {
		result.Status = "stalled"
		return result
	}
	turns := remaining / rate
	if remaining%rate != 0 {
		turns++
	}
	result.Turns = &turns
	result.Status = "estimated"
	return result
}

// allocateProduction is shared with execution so integer public-works rounding
// cannot drift between the displayed production rate and the next engine step.
func allocateProduction(production, percent int) (net, works int) {
	if production <= 0 {
		return production, 0
	}
	works = production * percent / 100
	if works > production {
		works = production
	}
	return production - works, works
}

// Production estimates only the owned queue head using the effective engine
// economy, completed improvements, injunction and public-works allocation.
func (i *Inspector) Production(cityID, actor string) *Forecast {
	city, ok := cityByID(i.state, cityID)
	if !ok || city.OwnerID != actor {
		return nil
	}
	result := newForecast(i.state.Revision)
	if len(city.Queue) == 0 {
		result.Status = "no-target"
		return result
	}
	item := city.Queue[0]
	if item.Kind == ProductionCapitalization || item.Kind == ProductionInfrastructure {
		result.Status = "continuous"
		return result
	}
	rate := 0
	if !hasEffect(city.Effects, EffectInjunction) {
		yield, err := cityYieldWithRuntime(i.state, *city, i.runtime)
		if err != nil {
			return result
		}
		rate = yield.Production
		if nation := nationForID(&i.state, nationForPlayer(i.state, actor)); nation != nil {
			rate, _ = allocateProduction(rate, publicWorksPercentForNation(*nation))
		}
	}
	return finishForecast(result, item.Cost-city.StoredProduction, rate)
}

// Research recalculates only the owner's current city yields in an isolated
// slice, then calls the execution research-rate helper. It never advances turns,
// reassigns specialists, mutates policy or predicts foreign decisions.
func (i *Inspector) Research(actor string) *Forecast {
	nation := nationForID(&i.state, nationForPlayer(i.state, actor))
	if nation == nil {
		return nil
	}
	result := newForecast(i.state.Revision)
	if nation.ResearchTargetID == "" {
		result.Status = "no-target"
		return result
	}
	if i.runtime == nil {
		return result
	}
	advance, ok := i.runtime.Advance(nation.ResearchTargetID)
	if !ok {
		return result
	}
	evaluated := i.state
	evaluated.Cities = append([]City(nil), i.state.Cities...)
	for index, city := range evaluated.Cities {
		if nationForPlayer(evaluated, city.OwnerID) != nation.ID {
			continue
		}
		yield, err := cityYieldWithRuntime(evaluated, city, i.runtime)
		if err != nil {
			return result
		}
		evaluated.Cities[index].Yields = yield
	}
	return finishForecast(result, advance.Cost-nation.ResearchProgress, researchRate(evaluated, *nation))
}
