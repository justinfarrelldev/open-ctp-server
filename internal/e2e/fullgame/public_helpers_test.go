package fullgame

import (
	"testing"

	"github.com/justinfarrelldev/open-ctp-server/internal/game"
	"github.com/justinfarrelldev/open-ctp-server/internal/gameplay/engine"
)

func TestResolvePublicRoleRejectsAbsentAndAmbiguousPlayers(t *testing.T) {
	valid := game.Game{Players: []game.GamePlayer{{AccountID: 9}}}
	if got, err := ResolvePublicRole(valid, 9); err != nil || got != "account-9" {
		t.Fatalf("valid role = %q, %v", got, err)
	}
	if _, err := ResolvePublicRole(game.Game{}, 9); err == nil {
		t.Fatal("absent role accepted")
	}
	if _, err := ResolvePublicRole(game.Game{Players: []game.GamePlayer{{AccountID: 9}, {AccountID: 9}}}, 9); err == nil {
		t.Fatal("ambiguous role accepted")
	}
}

func TestResolveVisibleArmyRequiresOneProjectedArmy(t *testing.T) {
	projection := []byte(`{"units":[{"id":"unit-a","army_id":"army-a"}]}`)
	if got, err := ResolveVisibleArmy(projection, "unit-a"); err != nil || got != "army-a" {
		t.Fatalf("visible army = %q, %v", got, err)
	}
	if _, err := ResolveVisibleArmy([]byte(`{"units":[{"id":"unit-a"}]}`), "unit-a"); err == nil {
		t.Fatal("unit without a projected army was accepted")
	}
}

func TestResolveOwnedCapabilityAndVisibleForeignCity(t *testing.T) {
	projection := []byte(`{"units":[{"id":"unit-z","owner_id":"player-a","effects":[{"kind":"special_embassy"}]},{"id":"unit-a","owner_id":"player-a","effects":[{"kind":"special_embassy"}]},{"id":"foreign","owner_id":"player-b","effects":[{"kind":"special_embassy"}]}],"cities":[{"id":"city-z","owner_id":"player-b"},{"id":"city-a","owner_id":"player-b"}]}`)
	unit, city, err := ResolveOwnedCapabilityAndVisibleForeignCity(projection, "player-a", engine.EffectSpecialEmbassy)
	if err != nil || unit != "unit-a" || city != "city-a" {
		t.Fatalf("roles = %q, %q, %v", unit, city, err)
	}
	if _, _, err := ResolveOwnedCapabilityAndVisibleForeignCity([]byte(`{"units":[{"id":"foreign","owner_id":"player-b","effects":[{"kind":"special_embassy"}]}],"cities":[{"id":"city","owner_id":"player-b"}]}`), "player-a", engine.EffectSpecialEmbassy); err == nil {
		t.Fatal("foreign effect accepted")
	}
	if _, _, err := ResolveOwnedCapabilityAndVisibleForeignCity([]byte(`{"units":[{"id":"unit","owner_id":"player-a","effects":[{"kind":"special_embassy"}]}],"cities":[{"id":"old","owner_id":"player-b","stale":true}]}`), "player-a", engine.EffectSpecialEmbassy); err == nil {
		t.Fatal("stale target accepted")
	}
	if _, _, err := ResolveOwnedCapabilityAndVisibleForeignCity([]byte(`not-json`), "player-a", engine.EffectSpecialEmbassy); err == nil {
		t.Fatal("malformed projection accepted")
	}
}
