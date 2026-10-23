package rules

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOriginalCivilizationCityNamesAreImportedAndPinned(t *testing.T) {
	root := localRetailGameDataRoot(t)
	if client := os.Getenv("CTPCLIENT"); client != "" {
		candidate := filepath.Join(client, "ctp2", "ctp2_data", "default", "gamedata")
		if _, err := os.Stat(filepath.Join(candidate, "civilisation.txt")); err == nil {
			root = candidate
		}
	}
	db, err := LoadRetailRulesFromDirectory(root)
	if err != nil {
		t.Fatal(err)
	}
	runtime := ProjectRetailRuntime(db)
	if runtime.CityNaming == nil {
		t.Fatal("original civilization city names were not pinned")
	}
	for _, tc := range []struct{ country, id, first string }{{"Greeks", "GREEK", "Athens"}, {"roman", "ROMAN", "Rome"}, {"america", "AMERICANS", "Washington D.C."}} {
		civ, ok := runtime.CityNaming.Resolve(tc.country)
		if !ok || civ.ID != tc.id || civ.CityNames[0] != tc.first {
			t.Fatalf("country %s resolved to %+v", tc.country, civ)
		}
	}
	definitions, hash, err := CanonicalRuntimeDefinitions(runtime)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := CompileRuntime(definitions, hash)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.CityNaming, runtime.CityNaming) {
		t.Fatal("names changed through canonical pin")
	}
	before, _ := json.Marshal(runtime.CityNaming)
	db.CityNaming.Civilizations[0].CityNames[0] = "mutation"
	after, _ := json.Marshal(runtime.CityNaming)
	if string(before) != string(after) {
		t.Fatal("projection shares mutable imported names")
	}
}

func TestExtendedCatalogPreservesExistingAmericaSetupAlias(t *testing.T) {
	db, err := LoadRetailRulesFromDirectory(localRetailGameDataRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	civ, ok := db.CityNaming.Resolve("america")
	if !ok || civ.ID != "AMERICANS" {
		t.Fatalf("public setup alias lost against renamed CountryName: %+v", civ)
	}
}

func TestCivilizationNameImporterRejectsMissingAndMalformedReferences(t *testing.T) {
	civ := []byte("ROMAN {\n CountryName ROMAN_COUNTRY\n SingularCivName ROMAN_SINGULAR\n PluralCivName ROMAN_PLURAL\n CityName ROME\n CityName ANTIUM\n}\n")
	strings := []byte("ROMAN_COUNTRY \"Rome's country\"\nROMAN_SINGULAR \"Roman\"\nROMAN_PLURAL \"Romans\"\nROME \"Rome\"\nANTIUM \"Antium\"\n")
	valid, err := ParseRetailCityNaming(civ, strings)
	if err != nil {
		t.Fatal(err)
	}
	if got := valid.Civilizations[0].CityNames; !reflect.DeepEqual(got, []string{"Rome", "Antium"}) {
		t.Fatalf("source order lost: %v", got)
	}
	for _, bad := range [][]byte{nil, []byte("ROME \"Rome\"\n"), append(append([]byte(nil), strings...), []byte("ROME \"duplicate\"\n")...), []byte("ROME \"unterminated")} {
		if _, err := ParseRetailCityNaming(civ, bad); err == nil {
			t.Fatalf("bad localized names accepted: %q", bad)
		}
	}
	latin := bytes.ReplaceAll(strings, []byte("Antium"), []byte("M\xe9xico"))
	decoded, err := ParseRetailCityNaming(civ, latin)
	if err != nil {
		t.Fatalf("original Windows-1252 string file rejected: %v", err)
	}
	if decoded.Civilizations[0].CityNames[1] != "México" {
		t.Fatalf("localized city name encoding lost: %q", decoded.Civilizations[0].CityNames[1])
	}
}
