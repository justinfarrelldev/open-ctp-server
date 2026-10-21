package rules

import "math"

// YearAtTurn implements DiffDB::GetYearFromTurn with an explicit phase epoch.
// Unknown legacy calendars, malformed periods and overflow return unavailable;
// they must not become a client hardcoded starting year or wraparound date.
func (calendar RuntimeCalendar) YearAtTurn(turn int) (int64, bool) {
	if calendar.StartYear == nil || calendar.EpochTurn < 1 || turn < calendar.EpochTurn || len(calendar.Periods) == 0 || calendar.Periods[0].StartTurn != 0 {
		return 0, false
	}
	for index, period := range calendar.Periods {
		if period.YearsPerTurn <= 0 || period.StartTurn < 0 || index > 0 && period.StartTurn <= calendar.Periods[index-1].StartTurn {
			return 0, false
		}
	}
	sourceTurn := turn - calendar.EpochTurn
	year := int64(*calendar.StartYear)
	for index, period := range calendar.Periods {
		if sourceTurn <= period.StartTurn {
			break
		}
		until := sourceTurn
		if index+1 < len(calendar.Periods) && until > calendar.Periods[index+1].StartTurn {
			until = calendar.Periods[index+1].StartTurn
		}
		span, increment := int64(until-period.StartTurn), int64(period.YearsPerTurn)
		if span > math.MaxInt64/increment {
			return 0, false
		}
		delta := span * increment
		if year > math.MaxInt64-delta {
			return 0, false
		}
		year += delta
	}
	return year, true
}

func projectRetailCalendar(calendar RetailCalendar) RuntimeCalendar {
	result := RuntimeCalendar{FinalTurn: retailRuntimeFinalTurn(calendar)}
	if len(calendar.Periods) == 0 {
		return result
	}
	start := calendar.StartYear
	result.StartYear = &start
	result.EpochTurn = 1
	for _, period := range calendar.Periods {
		result.Periods = append(result.Periods, RuntimeCalendarPeriod{StartTurn: period.StartTurn, YearsPerTurn: period.YearsPerTurn})
	}
	// Retail's round zero is the server's initial turn one. Correct this only in
	// newly generated pins; old persisted FinalTurn values are never rewritten.
	if result.FinalTurn < math.MaxInt {
		result.FinalTurn++
	}
	return result
}
