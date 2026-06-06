package world

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	fixtures "github.com/justinfarrelldev/open-ctp-server/internal/world/fixtures"
)

func TestFixtureInventoryIncludesAllHostParams(t *testing.T) {
	t.Helper()

	requiredNames := []string{"small-default", "small-wet", "small-dry"}
	seen := make(map[string]fixtures.Definition, len(fixtures.Inventory))
	for _, definition := range fixtures.Inventory {
		seen[definition.Name] = definition
	}

	for _, name := range requiredNames {
		definition, ok := seen[name]
		if !ok {
			t.Fatalf("missing required fixture %q", name)
		}
		assertFixtureDefinition(t, definition)
	}

	assertGenerationHostParamContract(t)

	for _, definition := range fixtures.Inventory {
		if _, err := GenerateWorldFromFixture(definition); err != nil {
			t.Fatalf("fixture inventory is ready but generator consumption is still red for %q: %v", definition.Name, err)
		}
	}
}

func TestFixtureWetnessChangesExpectations(t *testing.T) {
	wet, ok := fixtures.ByName("small-wet")
	if !ok {
		t.Fatal("missing small-wet fixture")
	}
	dry, ok := fixtures.ByName("small-dry")
	if !ok {
		t.Fatal("missing small-dry fixture")
	}

	if wet.HostParams.Wetness <= dry.HostParams.Wetness {
		t.Fatalf("wet fixture wetness=%d must be greater than dry wetness=%d", wet.HostParams.Wetness, dry.HostParams.Wetness)
	}
	if wet.MapParams.Homogeneity <= dry.MapParams.Homogeneity {
		t.Fatalf("wet fixture homogeneity=%d must differ materially from dry homogeneity=%d", wet.MapParams.Homogeneity, dry.MapParams.Homogeneity)
	}
	if wet.Expected.TerrainBias == dry.Expected.TerrainBias {
		t.Fatalf("wet/dry fixtures must encode different terrain expectations, both were %q", wet.Expected.TerrainBias)
	}
	if reflect.DeepEqual(wet.Expected.Notes, dry.Expected.Notes) {
		t.Fatal("wet/dry fixtures must explain different expected invariants")
	}

	if _, err := GenerateWorldFromFixture(wet); err != nil {
		t.Fatalf("wetness expectation remains intentionally red until generator parity exists: %v", err)
	}
	if _, err := GenerateWorldFromFixture(dry); err != nil {
		t.Fatalf("dry expectation remains intentionally red until generator parity exists: %v", err)
	}
}

func assertFixtureDefinition(t *testing.T, definition fixtures.Definition) {
	t.Helper()

	if strings.TrimSpace(definition.Name) == "" {
		t.Fatal("fixture name is required")
	}
	if strings.TrimSpace(definition.GeneratorVersion) == "" {
		t.Fatalf("fixture %q is missing generator version", definition.Name)
	}
	if definition.HostParams.Style == "" || definition.HostParams.RulesJSON == "" || definition.HostParams.ExclusionsJSON == "" {
		t.Fatalf("fixture %q is missing setup defaults in host params: %+v", definition.Name, definition.HostParams)
	}
	if definition.HostParams.StartingSettlers < 0 || definition.HostParams.StartingGold < 0 {
		t.Fatalf("fixture %q has invalid starting values: %+v", definition.Name, definition.HostParams)
	}
	if definition.HostParams.Seed == 0 {
		t.Fatalf("fixture %q must declare a deterministic seed", definition.Name)
	}
	if definition.HostParams.Wetness <= 0 {
		t.Fatalf("fixture %q must declare wetness", definition.Name)
	}
	if definition.HostParams.MapSize == "" || definition.HostParams.Climate == "" {
		t.Fatalf("fixture %q must declare map size and climate", definition.Name)
	}
	if len(definition.PlayerOrder) == 0 {
		t.Fatalf("fixture %q must declare player order", definition.Name)
	}
	if definition.MapParams.PlayerCount != len(definition.PlayerOrder) {
		t.Fatalf("fixture %q player count=%d want %d", definition.Name, definition.MapParams.PlayerCount, len(definition.PlayerOrder))
	}
	if !definition.Expected.SpawnCountMatchesPlayerOrder || !definition.Expected.UniquePlayerStarts || !definition.Expected.RequiresLandContinents {
		t.Fatalf("fixture %q must encode spawn and continent invariants: %+v", definition.Name, definition.Expected)
	}
	if strings.TrimSpace(definition.Expected.TerrainBias) == "" || strings.TrimSpace(definition.Expected.StartDistanceRule) == "" {
		t.Fatalf("fixture %q must declare terrain and start-distance expectations", definition.Name)
	}
	if len(definition.Expected.Notes) == 0 {
		t.Fatalf("fixture %q must explain expected invariants", definition.Name)
	}
	if len(definition.SourceReferences) < 3 {
		t.Fatalf("fixture %q must cite CTP2 source references, got %d", definition.Name, len(definition.SourceReferences))
	}

	requiredReferenceFiles := []string{
		"ctp2_code/gs/utility/gameinit.cpp",
		"ctp2_code/gs/world/wldgen.cpp",
		"ctp2_code/gs/world/WrldCont.cpp",
	}
	missingRefs := make([]string, 0)
	for _, file := range requiredReferenceFiles {
		if !fixtureHasReference(definition, file) {
			missingRefs = append(missingRefs, file)
		}
	}
	if len(missingRefs) > 0 {
		t.Fatalf("fixture %q missing references: %s", definition.Name, strings.Join(missingRefs, ", "))
	}
}

func assertGenerationHostParamContract(t *testing.T) {
	t.Helper()

	type fieldExpectation struct {
		Name string
		Kind reflect.Kind
	}

	expected := []fieldExpectation{
		{Name: "Style", Kind: reflect.String},
		{Name: "RulesJSON", Kind: reflect.String},
		{Name: "ExclusionsJSON", Kind: reflect.String},
		{Name: "StartingSettlers", Kind: reflect.Int},
		{Name: "StartingGold", Kind: reflect.Int},
		{Name: "Seed", Kind: reflect.Pointer},
		{Name: "Wetness", Kind: reflect.Pointer},
		{Name: "MapSize", Kind: reflect.String},
		{Name: "Climate", Kind: reflect.String},
		{Name: "GeneratorVersion", Kind: reflect.String},
	}

	hostParamType := reflect.TypeOf(GenerationHostParams{})
	for _, expectation := range expected {
		field, ok := hostParamType.FieldByName(expectation.Name)
		if !ok {
			t.Fatalf("GenerationHostParams missing field %q", expectation.Name)
		}
		if field.Type.Kind() != expectation.Kind {
			t.Fatalf("GenerationHostParams field %q kind=%s want %s", expectation.Name, field.Type.Kind(), expectation.Kind)
		}
	}
}

func fixtureHasReference(definition fixtures.Definition, file string) bool {
	for _, reference := range definition.SourceReferences {
		if reference.File == file && strings.TrimSpace(reference.Lines) != "" && strings.TrimSpace(reference.Reason) != "" {
			return true
		}
	}

	return false
}

func ExampleGenerateWorldFromFixture_redContract() {
	definition, _ := fixtures.ByName("small-default")
	_, err := GenerateWorldFromFixture(definition)
	fmt.Println(err)
	// Output: world fixture generation is not implemented yet
}
