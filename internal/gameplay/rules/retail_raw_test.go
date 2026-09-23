package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseRetailRawRecordsHandlesCTP2DatabaseSyntax(t *testing.T) {
	const input = `2
UNIT_EXAMPLE {
   Description DESCRIPTION_UNIT_EXAMPLE
   Attack 10
   MaxMovePoints 300
   IgnoreZOC
   CanSee: Standard
   CanSee: Stealth
   MovementType: Land
   CargoData {
      MaxCargo 2
      Load SOUND_LOAD
   }
}

ADVANCE_EXAMPLE {
   Prerequisites ADVANCE_ROOT
   Prerequisites ADVANCE_BRANCH
   Cost 1234
   Railroads
}
`

	records, err := ParseRetailRawRecords("fixture.txt", []byte(input), RetailCatalog{Filename: "fixture.txt", Kind: "unit", Prefixes: []string{"UNIT_", "ADVANCE_"}})
	if err != nil {
		t.Fatalf("ParseRetailRawRecords() error = %v", err)
	}
	if got, want := len(records), 2; got != want {
		t.Fatalf("len(records) = %d, want %d", got, want)
	}
	unit := records[0]
	if got, want := unit.ID, "UNIT_EXAMPLE"; got != want {
		t.Fatalf("record[0].ID = %q, want %q", got, want)
	}
	if got := unit.FirstValue("Attack"); got != "10" {
		t.Fatalf("Attack = %q, want 10", got)
	}
	if !unit.HasFlag("IgnoreZOC") {
		t.Fatalf("IgnoreZOC flag was not preserved")
	}
	if got, want := unit.Values("CanSee"), []string{"Standard", "Stealth"}; !sameStringSlices(got, want) {
		t.Fatalf("CanSee values = %#v, want %#v", got, want)
	}
	cargo, ok := unit.FirstBlock("CargoData")
	if !ok {
		t.Fatalf("CargoData block was not preserved")
	}
	if got := cargo.FirstValue("MaxCargo"); got != "2" {
		t.Fatalf("CargoData.MaxCargo = %q, want 2", got)
	}

	advance := records[1]
	if got, want := advance.Values("Prerequisites"), []string{"ADVANCE_ROOT", "ADVANCE_BRANCH"}; !sameStringSlices(got, want) {
		t.Fatalf("Prerequisites = %#v, want %#v", got, want)
	}
	if !advance.HasFlag("Railroads") {
		t.Fatalf("Railroads flag was not preserved")
	}
}

func TestParseRetailRawCatalogCountsLocalCTP2Data(t *testing.T) {
	root := localRetailGameDataRoot(t)
	catalogs := []struct {
		catalog RetailCatalog
		want    int
	}{
		{RetailCatalog{Filename: "Units.txt", Kind: "unit", Prefixes: []string{"UNIT_"}}, 74},
		{RetailCatalog{Filename: "Advance.txt", Kind: "advance", Prefixes: []string{"ADVANCE_"}}, 107},
		{RetailCatalog{Filename: "Wonder.txt", Kind: "wonder", Prefixes: []string{"WONDER_"}}, 30},
		{RetailCatalog{Filename: "buildings.txt", Kind: "building", Prefixes: []string{"IMPROVE_"}}, 52},
		{RetailCatalog{Filename: "tileimp.txt", Kind: "terrain_improvement", Prefixes: []string{"TILEIMP_"}}, 40},
		{RetailCatalog{Filename: "terrain.txt", Kind: "terrain", Prefixes: []string{"TERRAIN_"}}, 26},
		{RetailCatalog{Filename: "govern.txt", Kind: "government", Prefixes: []string{"GOVERNMENT_"}}, 12},
		{RetailCatalog{Filename: "pollution.txt", Kind: "pollution", Prefixes: []string{"POLLUTION_TRIGGER_LIST_"}}, 4},
	}

	for _, tc := range catalogs {
		t.Run(tc.catalog.Filename, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, tc.catalog.Filename))
			if err != nil {
				t.Fatalf("read retail catalog: %v", err)
			}
			records, err := ParseRetailRawRecords(tc.catalog.Filename, data, tc.catalog)
			if err != nil {
				t.Fatalf("ParseRetailRawRecords(%s) error = %v", tc.catalog.Filename, err)
			}
			if got := len(records); got != tc.want {
				t.Fatalf("len(records) = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestParseRetailRawCatalogPreservesSelectedRetailValues(t *testing.T) {
	root := localRetailGameDataRoot(t)

	units := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "Units.txt", Kind: "unit", Prefixes: []string{"UNIT_"}})
	carrier := mustRetailRecord(t, units, "UNIT_AIRCRAFT_CARRIER")
	if got := carrier.FirstValue("Attack"); got != "25" {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER Attack = %q, want 25", got)
	}
	if got := carrier.FirstValue("Defense"); got != "30" {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER Defense = %q, want 30", got)
	}
	if got := carrier.FirstValue("ShieldCost"); got != "3000" {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER ShieldCost = %q, want 3000", got)
	}
	if got := carrier.FirstValue("MaxMovePoints"); got != "400" {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER MaxMovePoints = %q, want 400", got)
	}
	if got, want := carrier.Values("CanCarry"), []string{"MedAir", "SmallAir"}; !sameStringSlices(got, want) {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER CanCarry = %#v, want %#v", got, want)
	}
	cargo, ok := carrier.FirstBlock("CargoData")
	if !ok {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER CargoData block missing")
	}
	if got := cargo.FirstValue("MaxCargo"); got != "5" {
		t.Fatalf("UNIT_AIRCRAFT_CARRIER CargoData.MaxCargo = %q, want 5", got)
	}

	advances := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "Advance.txt", Kind: "advance", Prefixes: []string{"ADVANCE_"}})
	railroad := mustRetailRecord(t, advances, "ADVANCE_RAILROAD")
	if got := railroad.FirstValue("Cost"); got != "10912" {
		t.Fatalf("ADVANCE_RAILROAD Cost = %q, want 10912", got)
	}
	if got, want := railroad.Values("Prerequisites"), []string{"ADVANCE_BANKING", "ADVANCE_MODERN_METALLURGY"}; !sameStringSlices(got, want) {
		t.Fatalf("ADVANCE_RAILROAD Prerequisites = %#v, want %#v", got, want)
	}
	if !railroad.HasFlag("Railroads") {
		t.Fatalf("ADVANCE_RAILROAD Railroads flag missing")
	}

	wonders := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "Wonder.txt", Kind: "wonder", Prefixes: []string{"WONDER_"}})
	zeroCrime := mustRetailRecord(t, wonders, "WONDER_ZERO_CRIME_BILL")
	if got := zeroCrime.FirstValue("ProductionCost"); got != "8600" {
		t.Fatalf("WONDER_ZERO_CRIME_BILL ProductionCost = %q, want 8600", got)
	}
	if got := zeroCrime.FirstValue("DecCrimePercent"); got != "60" {
		t.Fatalf("WONDER_ZERO_CRIME_BILL DecCrimePercent = %q, want 60", got)
	}

	buildings := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "buildings.txt", Kind: "building", Prefixes: []string{"IMPROVE_"}})
	factory := mustRetailRecord(t, buildings, "IMPROVE_FACTORY")
	if got := factory.FirstValue("ProductionCost"); got != "2025" {
		t.Fatalf("IMPROVE_FACTORY ProductionCost = %q, want 2025", got)
	}
	if got := factory.FirstValue("Upkeep"); got != "8" {
		t.Fatalf("IMPROVE_FACTORY Upkeep = %q, want 8", got)
	}
	if got := factory.FirstValue("ProductionPercent"); got != "0.15" {
		t.Fatalf("IMPROVE_FACTORY ProductionPercent = %q, want 0.15", got)
	}
	if !factory.HasFlag("AllowGrunts") {
		t.Fatalf("IMPROVE_FACTORY AllowGrunts flag missing")
	}

	tileImps := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "tileimp.txt", Kind: "terrain_improvement", Prefixes: []string{"TILEIMP_"}})
	farms := mustRetailRecord(t, tileImps, "TILEIMP_FARMS")
	if got := farms.FirstValue("Class"); got != "Farm" {
		t.Fatalf("TILEIMP_FARMS Class = %q, want Farm", got)
	}
	terrainEffect, ok := farms.FirstBlock("TerrainEffect")
	if !ok {
		t.Fatalf("TILEIMP_FARMS TerrainEffect block missing")
	}
	if got := terrainEffect.FirstValue("BonusFood"); got != "10" {
		t.Fatalf("TILEIMP_FARMS TerrainEffect.BonusFood = %q, want 10", got)
	}
	if got := terrainEffect.FirstValue("ProductionCost"); got != "200" {
		t.Fatalf("TILEIMP_FARMS TerrainEffect.ProductionCost = %q, want 200", got)
	}

	pollution := mustParseRetailCatalog(t, root, RetailCatalog{Filename: "pollution.txt", Kind: "pollution", Prefixes: []string{"POLLUTION_TRIGGER_LIST_"}})
	smallMap := mustRetailRecord(t, pollution, "POLLUTION_TRIGGER_LIST_SMALL_MAP")
	if got, want := smallMap.Values("POLLUTION_TRIGGER"), []string{"4000000", "6000000", "8000000", "10000000", "15000000", "9000000000"}; !sameStringSlices(got, want) {
		t.Fatalf("small-map pollution trigger levels = %#v, want %#v", got, want)
	}
}

func localRetailGameDataRoot(t *testing.T) string {
	t.Helper()
	candidates := []string{
		filepath.Join("..", "..", "..", "..", "open-ctp-server", "ctp2", "ctp2_released_game", "ctp2_data", "default", "gamedata"),
		"/home/justi/Projects/open-ctp-server/ctp2/ctp2_released_game/ctp2_data/default/gamedata",
	}
	for _, candidate := range candidates {
		if stat, err := os.Stat(filepath.Join(candidate, "Units.txt")); err == nil && !stat.IsDir() {
			return candidate
		}
	}
	t.Skip("local CTP2 released-game data is not present")
	return ""
}

func mustParseRetailCatalog(t *testing.T, root string, catalog RetailCatalog) []RetailRawRecord {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, catalog.Filename))
	if err != nil {
		t.Fatalf("read retail catalog %s: %v", catalog.Filename, err)
	}
	records, err := ParseRetailRawRecords(catalog.Filename, data, catalog)
	if err != nil {
		t.Fatalf("ParseRetailRawRecords(%s) error = %v", catalog.Filename, err)
	}
	return records
}

func mustRetailRecord(t *testing.T, records []RetailRawRecord, id string) RetailRawRecord {
	t.Helper()
	for _, record := range records {
		if record.ID == id {
			return record
		}
	}
	t.Fatalf("retail record %q not found", id)
	return RetailRawRecord{}
}

func sameStringSlices(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
