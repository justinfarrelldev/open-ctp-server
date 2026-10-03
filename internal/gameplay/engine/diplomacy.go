package engine

import (
	"encoding/json"
	"errors"
	"fmt"
)

type contactPayload struct {
	TargetPlayerID string `json:"target_player_id"`
}

type proposeAgreementPayload struct {
	ProposalID     string `json:"proposal_id"`
	TargetPlayerID string `json:"target_player_id"`
	Kind           string `json:"kind"`
}

type respondAgreementPayload struct {
	ProposalID string `json:"proposal_id"`
	Accept     bool   `json:"accept"`
}

func applyDiplomacyCommand(state *State, command Command) (string, json.RawMessage, error) {
	switch command.Type {
	case CommandSetWar:
		return applyWarStateCommand(state, command)
	case CommandEstablishContact:
		var payload contactPayload
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return "", nil, errors.New("diplomacy payload is invalid")
		}
		return establishContact(state, command.ActorPlayerID, payload)
	case CommandProposeAgreement:
		var payload proposeAgreementPayload
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return "", nil, errors.New("diplomacy payload is invalid")
		}
		return proposeAgreement(state, command.ActorPlayerID, payload)
	case CommandRespondAgreement:
		var payload respondAgreementPayload
		if err := json.Unmarshal(command.Payload, &payload); err != nil {
			return "", nil, errors.New("diplomacy payload is invalid")
		}
		return respondAgreement(state, command.ActorPlayerID, payload)
	default:
		return "", nil, fmt.Errorf("unsupported diplomacy command %q", command.Type)
	}
}

func establishContact(state *State, actorID string, payload contactPayload) (string, json.RawMessage, error) {
	if !playersCanRelate(*state, actorID, payload.TargetPlayerID) {
		return "", nil, errors.New("diplomacy target is invalid")
	}
	relation := ensureRelation(state, actorID, payload.TargetPlayerID)
	relation.Contact = true
	if relation.EstablishedAt == 0 {
		relation.EstablishedAt = state.Phase.Turn
	}
	encoded, _ := json.Marshal(contactPayload{TargetPlayerID: payload.TargetPlayerID})
	return EventContactEstablished, encoded, nil
}

func proposeAgreement(state *State, actorID string, payload proposeAgreementPayload) (string, json.RawMessage, error) {
	if payload.ProposalID == "" || !validAgreementKind(payload.Kind) || !playersCanRelate(*state, actorID, payload.TargetPlayerID) {
		return "", nil, errors.New("diplomacy proposal is invalid")
	}
	relation := relationForPlayers(*state, actorID, payload.TargetPlayerID)
	if relation == nil || !relation.Contact {
		return "", nil, errors.New("diplomacy contact is required")
	}
	for _, proposal := range state.Proposals {
		if proposal.ID == payload.ProposalID || (proposal.FromPlayerID == actorID && proposal.ToPlayerID == payload.TargetPlayerID && proposal.Kind == payload.Kind) {
			return "", nil, errors.New("diplomacy proposal already exists")
		}
	}
	state.Proposals = append(state.Proposals, DiplomaticProposal{ID: payload.ProposalID, FromPlayerID: actorID, ToPlayerID: payload.TargetPlayerID, Kind: payload.Kind, CreatedTurn: state.Phase.Turn})
	encoded, _ := json.Marshal(payload)
	return EventAgreementProposed, encoded, nil
}

func respondAgreement(state *State, actorID string, payload respondAgreementPayload) (string, json.RawMessage, error) {
	if payload.ProposalID == "" {
		return "", nil, errors.New("diplomacy response is invalid")
	}
	index := -1
	var proposal DiplomaticProposal
	for i, candidate := range state.Proposals {
		if candidate.ID == payload.ProposalID {
			index, proposal = i, candidate
			break
		}
	}
	if index < 0 || proposal.ToPlayerID != actorID {
		return "", nil, errors.New("diplomacy response is invalid")
	}
	state.Proposals = append(state.Proposals[:index], state.Proposals[index+1:]...)
	if payload.Accept {
		relation := ensureRelation(state, proposal.FromPlayerID, proposal.ToPlayerID)
		relation.Contact = true
		applyAcceptedAgreement(state, relation, proposal.Kind)
	}
	encoded, _ := json.Marshal(struct {
		ProposalID string `json:"proposal_id"`
		Accept     bool   `json:"accept"`
		Kind       string `json:"kind"`
	}{ProposalID: payload.ProposalID, Accept: payload.Accept, Kind: proposal.Kind})
	return EventAgreementResponded, encoded, nil
}

func applyAcceptedAgreement(state *State, relation *DiplomaticRelation, kind string) {
	switch kind {
	case "trade":
		relation.Trade = true
		relation.Treaty = "trade"
	case "peace":
		relation.Treaty = "peace"
		removeWar(state, relation.PlayerAID, relation.PlayerBID)
	case "union":
		relation.Treaty = "union"
		relation.Trade = true
		addReciprocalEmbassies(state, relation.PlayerAID, relation.PlayerBID)
		removeWar(state, relation.PlayerAID, relation.PlayerBID)
	}
}

func validateDiplomacyState(state State) error {
	players := makeIDSet(playerIDs(state.Players))
	for _, relation := range state.Diplomacy {
		if relation.ID == "" || relation.PlayerAID == "" || relation.PlayerBID == "" || relation.PlayerAID >= relation.PlayerBID || !players[relation.PlayerAID] || !players[relation.PlayerBID] || !validTreaty(relation.Treaty) || relation.EstablishedAt < 0 || relation.RegardBonusAToB < 0 || relation.RegardTurnsAToB < 0 || relation.RegardBonusBToA < 0 || relation.RegardTurnsBToA < 0 {
			return errors.New("diplomacy relation is invalid")
		}
		if relation.ID != relationID(relation.PlayerAID, relation.PlayerBID) {
			return errors.New("diplomacy relation id is invalid")
		}
	}
	for _, proposal := range state.Proposals {
		if proposal.ID == "" || proposal.FromPlayerID == "" || proposal.ToPlayerID == "" || proposal.FromPlayerID == proposal.ToPlayerID || !players[proposal.FromPlayerID] || !players[proposal.ToPlayerID] || !validAgreementKind(proposal.Kind) || proposal.CreatedTurn < 0 {
			return errors.New("diplomacy proposal is invalid")
		}
	}
	return nil
}

func playersCanRelate(state State, left, right string) bool {
	if left == "" || right == "" || left == right {
		return false
	}
	players := makeIDSet(playerIDs(state.Players))
	return players[left] && players[right]
}

func relationForPlayers(state State, left, right string) *DiplomaticRelation {
	id := relationID(left, right)
	for i := range state.Diplomacy {
		if state.Diplomacy[i].ID == id {
			return &state.Diplomacy[i]
		}
	}
	return nil
}

func ensureRelation(state *State, left, right string) *DiplomaticRelation {
	id := relationID(left, right)
	for i := range state.Diplomacy {
		if state.Diplomacy[i].ID == id {
			return &state.Diplomacy[i]
		}
	}
	a, b := sortedPlayers(left, right)
	state.Diplomacy = append(state.Diplomacy, DiplomaticRelation{ID: id, PlayerAID: a, PlayerBID: b})
	return &state.Diplomacy[len(state.Diplomacy)-1]
}

func relationID(left, right string) string {
	a, b := sortedPlayers(left, right)
	return a + ":" + b
}

func sortedPlayers(left, right string) (string, string) {
	if right < left {
		return right, left
	}
	return left, right
}

func validAgreementKind(kind string) bool {
	return kind == "trade" || kind == "peace" || kind == "union"
}
func validTreaty(treaty string) bool {
	return treaty == "" || treaty == "trade" || treaty == "peace" || treaty == "union"
}

func removeWar(state *State, left, right string) {
	a, b := sortedPlayers(left, right)
	filtered := state.Policy.Wars[:0]
	for _, war := range state.Policy.Wars {
		wa, wb := sortedPlayers(war.A, war.B)
		if wa == a && wb == b {
			continue
		}
		filtered = append(filtered, TurnPolicyWar{A: wa, B: wb})
	}
	state.Policy.Wars = filtered
}

func addReciprocalEmbassies(state *State, leftPlayerID, rightPlayerID string) {
	leftNationID, rightNationID := nationForPlayer(*state, leftPlayerID), nationForPlayer(*state, rightPlayerID)
	leftNation, rightNation := nationForID(state, leftNationID), nationForID(state, rightNationID)
	if leftNation == nil || rightNation == nil {
		return
	}
	if !containsString(leftNation.Embassies, rightNation.ID) {
		leftNation.Embassies = append(leftNation.Embassies, rightNation.ID)
	}
	if !containsString(rightNation.Embassies, leftNation.ID) {
		rightNation.Embassies = append(rightNation.Embassies, leftNation.ID)
	}
}
