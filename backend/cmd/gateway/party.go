package main

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/narrator"
	"dnd5e-web/backend/internal/wsproto"
)

// broadcastPartyState sends every current member the up-to-date roster
// (display name, active character name, HP) for their party.
func (s *server) broadcastPartyState(ctx context.Context, partyID string) {
	accounts, err := s.store.ListAccountsByPartyID(ctx, partyID)
	if err != nil {
		return
	}
	members := make([]wsproto.PartyMember, 0, len(accounts))
	for _, a := range accounts {
		m := wsproto.PartyMember{AccountID: a.ID, DisplayName: a.DisplayName}
		if a.ActiveCharacterID != nil {
			if c, err := s.store.GetCharacter(ctx, *a.ActiveCharacterID); err == nil {
				m.CharacterName = c.Name
				m.HPCurrent = c.HPCurrent
				m.HPMax = c.HPMax
			}
		}
		members = append(members, m)
	}
	for _, a := range accounts {
		s.hub.SendTo(a.ID, wsproto.PartyState{Type: "PARTY_STATE", PartyID: partyID, Members: members})
	}
}

// handleInviteToParty looks up the target by display name and pushes them
// a PARTY_INVITE_RECEIVED — purely an in-memory handshake (mirrors how
// voteRooms works), nothing is persisted until the invite is accepted.
func (s *server) handleInviteToParty(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.InviteToPartyPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	target, err := s.store.FindAccountByDisplayName(ctx, p.TargetDisplayName)
	if err != nil {
		return wsError("no adventurer found with that name")
	}
	if target.ID == client.AccountID {
		return wsError("you can't invite yourself")
	}
	inviter, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}

	invite := &partyInvite{
		ID:              uuid.NewString(),
		FromAccountID:   client.AccountID,
		FromDisplayName: inviter.DisplayName,
		ToAccountID:     target.ID,
	}
	s.partyInvitesMu.Lock()
	s.partyInvites[invite.ID] = invite
	s.partyInvitesMu.Unlock()

	s.hub.SendTo(target.ID, wsproto.PartyInviteReceived{
		Type:            "PARTY_INVITE_RECEIVED",
		InviteID:        invite.ID,
		FromAccountID:   invite.FromAccountID,
		FromDisplayName: invite.FromDisplayName,
	})
	return nil
}

func (s *server) takeInvite(inviteID string) (*partyInvite, bool) {
	s.partyInvitesMu.Lock()
	defer s.partyInvitesMu.Unlock()
	invite, ok := s.partyInvites[inviteID]
	if ok {
		delete(s.partyInvites, inviteID)
	}
	return invite, ok
}

func (s *server) handleAcceptPartyInvite(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.RespondToPartyInvitePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	invite, ok := s.takeInvite(p.InviteID)
	if !ok || invite.ToAccountID != client.AccountID {
		return wsError("that invite is no longer valid")
	}

	target, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if target.PartyID != nil {
		return wsError("you're already in a party — leave it first")
	}

	inviter, err := s.store.GetAccount(ctx, invite.FromAccountID)
	if err != nil {
		return err
	}

	partyID := ""
	if inviter.PartyID != nil {
		partyID = *inviter.PartyID
	} else {
		newID, err := s.store.CreateParty(ctx, inviter.ID)
		if err != nil {
			return err
		}
		partyID = newID
		if err := s.store.SetPartyID(ctx, inviter.ID, &partyID); err != nil {
			return err
		}
		s.hub.SetPartyID(inviter.ID, partyID)
	}

	if err := s.store.SetPartyID(ctx, client.AccountID, &partyID); err != nil {
		return err
	}
	s.hub.SetPartyID(client.AccountID, partyID)

	line := narrator.PartyFormed(inviter.DisplayName, target.DisplayName)
	s.sendNarration(inviter.ID, "", line)
	s.sendNarration(client.AccountID, "", line)

	s.broadcastPartyState(ctx, partyID)

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}

func (s *server) handleDeclinePartyInvite(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.RespondToPartyInvitePayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	invite, ok := s.takeInvite(p.InviteID)
	if ok && invite.ToAccountID == client.AccountID {
		declinerName := client.AccountID
		if account, err := s.store.GetAccount(ctx, client.AccountID); err == nil {
			declinerName = account.DisplayName
		}
		s.hub.SendTo(invite.FromAccountID, wsproto.NewError(declinerName+" declined the party invite"))
	}
	return nil
}

func (s *server) handleLeaveParty(ctx context.Context, client *chat.Client) error {
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if account.PartyID == nil {
		return nil
	}
	oldPartyID := *account.PartyID
	if err := s.store.SetPartyID(ctx, client.AccountID, nil); err != nil {
		return err
	}
	s.hub.SetPartyID(client.AccountID, "")

	s.broadcastPartyState(ctx, oldPartyID)
	s.hub.SendTo(client.AccountID, wsproto.PartyState{Type: "PARTY_STATE", Members: []wsproto.PartyMember{}})

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}
