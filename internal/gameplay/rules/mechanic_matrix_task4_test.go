package rules

import (
	"strings"
	"testing"
)

func TestTask4ManifestMechanicMatrixAndContractsAreComplete(t *testing.T) {
	t.Parallel()

	manifest := SupportedManifest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if got, want := len(manifest.MechanicMatrix), len(requiredMechanicFamilies())+1; got != want {
		t.Fatalf("mechanic matrix rows = %d, want %d", got, want)
	}
	if got, want := len(manifest.SyntheticContracts), 10; got != want {
		t.Fatalf("synthetic contract rows = %d, want %d", got, want)
	}
	if got, want := strings.Join(mechanicFamilies(manifest.MechanicMatrix), ","), strings.Join(append(requiredMechanicFamilies(), "client-presentation-assets"), ","); got != want {
		t.Fatalf("mechanic families = %q, want %q", got, want)
	}
	if got, want := strings.Join(contractFamilies(manifest.SyntheticContracts), ","), strings.Join(requiredMechanicFamilies(), ","); got != want {
		t.Fatalf("contract families = %q, want %q", got, want)
	}
	for _, row := range manifest.MechanicMatrix {
		if row.Status != mechanicStatusExplicitlyExcluded && row.ContractID == "" {
			t.Fatalf("mechanic row %q is missing contract id", row.Family)
		}
	}
	for _, contract := range manifest.SyntheticContracts {
		if contract.Source.Repository != upstreamRepository || contract.Source.Commit != upstreamCommit || contract.Source.Path == "" || contract.Source.Symbol == "" {
			t.Fatalf("contract %q has invalid source provenance: %+v", contract.ID, contract.Source)
		}
		if len(contract.RequiredTags) == 0 {
			t.Fatalf("contract %q is missing required tags", contract.ID)
		}
	}
}

func TestTask4ManifestRejectsIncludedMechanicMissingRuntimeLayer(t *testing.T) {
	t.Parallel()

	manifest := SupportedManifest()
	setMechanicField(&manifest, "units", "runtime", "")
	if err := manifest.Validate(); err == nil || !strings.Contains(err.Error(), "missing runtime") {
		t.Fatalf("Validate() error = %v, want missing runtime rejection", err)
	}
}

func TestTask4ManifestRejectsIncludedMechanicMissingAPIStatus(t *testing.T) {
	t.Parallel()

	manifest := SupportedManifest()
	setMechanicField(&manifest, "special-actions", "api", "")
	if err := manifest.Validate(); err == nil || !strings.Contains(err.Error(), "missing api") {
		t.Fatalf("Validate() error = %v, want missing api rejection", err)
	}
}

func TestTask4ManifestRejectsUnknownMechanicStatus(t *testing.T) {
	t.Parallel()

	manifest := SupportedManifest()
	manifest.MechanicMatrix[0].Status = "future-maybe"
	if err := manifest.Validate(); err == nil || !strings.Contains(err.Error(), "unknown mechanic status") {
		t.Fatalf("Validate() error = %v, want unknown status rejection", err)
	}
}

func setMechanicField(manifest *SupportedRulesManifest, family, field, value string) {
	for i := range manifest.MechanicMatrix {
		if manifest.MechanicMatrix[i].Family != family {
			continue
		}
		switch field {
		case "runtime":
			manifest.MechanicMatrix[i].Runtime = value
		case "api":
			manifest.MechanicMatrix[i].API = value
		default:
			panic("unsupported field mutation")
		}
		return
	}
	panic("family not found: " + family)
}

func mechanicFamilies(rows []MechanicReadiness) []string {
	families := make([]string, 0, len(rows))
	for _, row := range rows {
		families = append(families, row.Family)
	}
	return families
}

func contractFamilies(contracts []SyntheticRecordContract) []string {
	families := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		families = append(families, contract.Family)
	}
	return families
}
