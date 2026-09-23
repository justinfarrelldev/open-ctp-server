package rules

const RetailPollutionHistoryTurns = 5

const (
	RetailPollutionTrendDownward = "downward"
	RetailPollutionTrendLevel    = "level"
	RetailPollutionTrendUpward   = "upward"
)

// RetailPollutionState mirrors the small global-pollution state managed by
// CTP2's Pollution object: a trigger phase, global-warming phase, short global
// pollution history, and trend.
type RetailPollutionState struct {
	Phase   int     `json:"phase"`
	GWPhase int     `json:"gw_phase"`
	History []int64 `json:"history"`
	Trend   string  `json:"trend,omitempty"`
}

type RetailPollutionEvent struct {
	Triggered      bool   `json:"triggered"`
	Phase          int    `json:"phase"`
	TriggerLevel   int64  `json:"trigger_level,omitempty"`
	PollutionLevel int64  `json:"pollution_level,omitempty"`
	DisasterType   string `json:"disaster_type,omitempty"`
}

func RetailPollutionBeginTurn(state RetailPollutionState, triggers RetailPollutionTriggerList, playerPollution []int64, worldPollutionReductionActive bool) (RetailPollutionState, RetailPollutionEvent) {
	next := normalizeRetailPollutionState(state)
	level := RetailGlobalPollutionLevel(next, playerPollution)
	event := RetailPollutionEvent{Phase: next.Phase, PollutionLevel: level}
	if worldPollutionReductionActive || next.Phase < 0 || next.Phase >= len(triggers.TriggerLevels) {
		return next, event
	}
	trigger := triggers.TriggerLevels[next.Phase]
	if level <= trigger {
		return next, event
	}
	event.Triggered = true
	event.TriggerLevel = trigger
	if next.Phase < len(triggers.DisasterTypes) {
		event.DisasterType = triggers.DisasterTypes[next.Phase]
	}
	if next.Phase < len(triggers.TriggerLevels)-1 {
		next.Phase++
	}
	return next, event
}

func RetailPollutionEndRound(state RetailPollutionState, playerPollution []int64) RetailPollutionState {
	next := normalizeRetailPollutionState(state)
	pollution := RetailGlobalPollutionLevel(next, playerPollution)
	copy(next.History[1:], next.History[:RetailPollutionHistoryTurns-1])
	next.History[0] = pollution
	next.Trend = RetailPollutionTrend(next.History)
	return next
}

func RetailGlobalPollutionLevel(state RetailPollutionState, playerPollution []int64) int64 {
	history := normalizeRetailPollutionHistory(state.History)
	pollution := history[0]
	for _, value := range playerPollution {
		pollution += value
	}
	if pollution < 0 {
		return 0
	}
	return pollution
}

func RetailPollutionRoundsToNextDisaster(state RetailPollutionState, triggers RetailPollutionTriggerList, worldPollutionReductionActive bool) int64 {
	if worldPollutionReductionActive || state.Phase < 0 || state.Phase >= len(triggers.TriggerLevels) {
		return 0
	}
	history := normalizeRetailPollutionHistory(state.History)
	if history[0] <= 0 {
		return 0
	}
	delta := history[0] - history[1]
	if delta <= 0 {
		return 0
	}
	return (triggers.TriggerLevels[state.Phase] - history[0]) / delta
}

func RetailPollutionTrend(history []int64) string {
	normalized := normalizeRetailPollutionHistory(history)
	var sx, sy float64
	for i, level := range normalized {
		x := float64(i + 1)
		sx += x
		sy += float64(level)
	}
	count := float64(len(normalized))
	sxoss := sx / count
	var st2, slope float64
	for i, level := range normalized {
		t := float64(i+1) - sxoss
		st2 += t * t
		slope += t * float64(level)
	}
	if st2 != 0 {
		slope /= st2
	}
	if slope < 0 {
		return RetailPollutionTrendDownward
	}
	if slope > 0 {
		return RetailPollutionTrendUpward
	}
	return RetailPollutionTrendLevel
}

func normalizeRetailPollutionState(state RetailPollutionState) RetailPollutionState {
	state.History = normalizeRetailPollutionHistory(state.History)
	if state.Trend == "" {
		state.Trend = RetailPollutionTrend(state.History)
	}
	return state
}

func normalizeRetailPollutionHistory(history []int64) []int64 {
	normalized := make([]int64, RetailPollutionHistoryTurns)
	copy(normalized, history)
	return normalized
}
