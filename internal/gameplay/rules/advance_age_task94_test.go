package rules

import "testing"

func TestRetailAdvanceAgesAreProjected(t *testing.T) {
	db, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(db.Ages) < 2 {
		t.Fatalf("ages = %d, want retail age records", len(db.Ages))
	}
	ageTwo, ok := db.Age("AGE_TWO")
	if !ok || ageTwo.Index != 1 || ageTwo.NeedAdvances != 5 {
		t.Fatalf("AGE_TWO = %+v ok=%v, want zero-based index 1 and five required advances", ageTwo, ok)
	}
	advance, ok := db.Advance("ADVANCE_AGRICULTURE")
	if !ok || advance.AgeID == "" {
		t.Fatalf("ADVANCE_AGRICULTURE = %+v ok=%v, want source age", advance, ok)
	}

	runtime := ProjectRetailRuntime(db)
	if runtime.StartAgeID != "AGE_ONE" || runtime.EndAgeID != "AGE_TEN" {
		t.Fatalf("runtime start/end ages = %q/%q, want AGE_ONE/AGE_TEN", runtime.StartAgeID, runtime.EndAgeID)
	}
	projected, ok := runtime.Advance("ADVANCE_AGRICULTURE")
	if !ok || projected.AgeID != advance.AgeID {
		t.Fatalf("projected agriculture = %+v ok=%v, want age %q", projected, ok, advance.AgeID)
	}
}
