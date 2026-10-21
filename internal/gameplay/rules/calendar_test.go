package rules

import (
	"math"
	"testing"
)

func TestClientCalendarUsesOneBasedPhaseEpochAndPinnedPeriods(t *testing.T) {
	start := -4000
	calendar := RuntimeCalendar{EpochTurn: 1, StartYear: &start, Periods: []RuntimeCalendarPeriod{{StartTurn: 0, YearsPerTurn: 50}, {StartTurn: 100, YearsPerTurn: 10}, {StartTurn: 200, YearsPerTurn: 1}}}
	for turn, want := range map[int]int64{1: -4000, 100: 950, 101: 1000, 102: 1010, 201: 2000, 202: 2001} {
		year, ok := calendar.YearAtTurn(turn)
		if !ok || year != want {
			t.Errorf("turn %d = %d known=%t, want %d", turn, year, ok, want)
		}
	}
	if _, ok := (RuntimeCalendar{FinalTurn: 1000}).YearAtTurn(1); ok {
		t.Fatal("legacy pin invented calendar")
	}
	calendar.Periods[0].YearsPerTurn = math.MaxInt
	if _, ok := calendar.YearAtTurn(math.MaxInt); ok {
		t.Fatal("overflowed calendar returned a fabricated date")
	}
}
