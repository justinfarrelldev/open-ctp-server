package engine

import (
	"strings"
	"testing"
)

func TestTask9DiplomacyTradeGovernmentLifecycle(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Nations[0].Advances = []string{"ADVANCE_SYNTHETIC_ROOT"}
	state = runtimeApply(t, state, "player-a", CommandEstablishContact, contactPayload{TargetPlayerID: "player-b"})
	if len(state.Diplomacy) != 1 || !state.Diplomacy[0].Contact {
		t.Fatalf("contact state = %#v", state.Diplomacy)
	}
	state = runtimeApply(t, state, "player-a", CommandProposeAgreement, proposeAgreementPayload{ProposalID: "trade-1", TargetPlayerID: "player-b", Kind: "trade"})
	if len(state.Proposals) != 1 {
		t.Fatalf("proposal state = %#v", state.Proposals)
	}
	state = runtimeApply(t, state, "player-b", CommandRespondAgreement, respondAgreementPayload{ProposalID: "trade-1", Accept: true})
	if len(state.Proposals) != 0 || len(state.Diplomacy) != 1 || !state.Diplomacy[0].Trade || state.Diplomacy[0].Treaty != "trade" {
		t.Fatalf("accepted trade state = %#v proposals=%#v", state.Diplomacy, state.Proposals)
	}
	state = runtimeApply(t, state, "player-a", CommandSetGovernment, setGovernmentPayload{GovernmentID: "GOVERNMENT_SYNTHETIC_REPUBLIC"})
	if state.Nations[0].GovernmentID != "GOVERNMENT_SYNTHETIC_REPUBLIC" || !hasEffect(state.Nations[0].Effects, EffectResearch) {
		t.Fatalf("government state = %#v", state.Nations[0])
	}
}

func TestTask9UnionTreatyCanDriveDiplomaticVictory(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	state.Phase.ActivePlayerOrder = []string{"player-a"}
	state.Cities = []City{{ID: "city-a", OwnerID: "player-a", TileID: "tile-a"}, {ID: "city-b", OwnerID: "player-b", TileID: "tile-b"}}
	state = runtimeApply(t, state, "player-a", CommandEstablishContact, contactPayload{TargetPlayerID: "player-b"})
	state = runtimeApply(t, state, "player-a", CommandProposeAgreement, proposeAgreementPayload{ProposalID: "union-1", TargetPlayerID: "player-b", Kind: "union"})
	state = runtimeApply(t, state, "player-b", CommandRespondAgreement, respondAgreementPayload{ProposalID: "union-1", Accept: true})
	state = runtimeApply(t, state, "player-a", CommandRoundComplete, struct{}{})
	if !state.Terminal.Ended || state.Terminal.Reason != "diplomatic_world_union" {
		t.Fatalf("union treaty did not end game: %#v", state.Terminal)
	}
}

func TestTask9InvalidDiplomacyAndGovernmentFailClosed(t *testing.T) {
	state := runtimeStartedState(t)
	state.Rules = serverOwnedRulesPin(t)
	cases := []struct {
		name    string
		actor   string
		command string
		payload any
		want    string
	}{
		{name: "proposal without contact", actor: "player-a", command: CommandProposeAgreement, payload: proposeAgreementPayload{ProposalID: "p", TargetPlayerID: "player-b", Kind: "trade"}, want: "contact"},
		{name: "bad agreement kind", actor: "player-a", command: CommandProposeAgreement, payload: proposeAgreementPayload{ProposalID: "p", TargetPlayerID: "player-b", Kind: "script"}, want: "proposal"},
		{name: "government missing advance", actor: "player-a", command: CommandSetGovernment, payload: setGovernmentPayload{GovernmentID: "GOVERNMENT_SYNTHETIC_REPUBLIC"}, want: "prerequisite"},
		{name: "unknown government", actor: "player-a", command: CommandSetGovernment, payload: setGovernmentPayload{GovernmentID: "GOVERNMENT_MISSING"}, want: "government"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := runtimeApplyExpectError(t, state, tc.actor, tc.command, tc.payload)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v want %q", err, tc.want)
			}
		})
	}
	contact := runtimeApply(t, state, "player-a", CommandEstablishContact, contactPayload{TargetPlayerID: "player-b"})
	proposed := runtimeApply(t, contact, "player-a", CommandProposeAgreement, proposeAgreementPayload{ProposalID: "p", TargetPlayerID: "player-b", Kind: "trade"})
	if err := runtimeApplyExpectError(t, proposed, "player-a", CommandRespondAgreement, respondAgreementPayload{ProposalID: "p", Accept: true}); err == nil || !strings.Contains(err.Error(), "response") {
		t.Fatalf("wrong actor response error = %v", err)
	}
}
