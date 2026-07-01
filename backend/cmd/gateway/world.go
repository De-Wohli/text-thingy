package main

import (
	"context"
	"encoding/json"

	"dnd5e-web/backend/internal/chat"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/narrator"
	"dnd5e-web/backend/internal/world"
	"dnd5e-web/backend/internal/wsproto"
)

// sendLocationState pushes the current location (description, exits) and
// who else is standing there to one client.
func (s *server) sendLocationState(ctx context.Context, client *chat.Client, locationID models.LocationID, narration string) error {
	loc, ok := world.Locations[locationID]
	if !ok {
		return wsError("unknown location")
	}
	present := s.buildPresentAccounts(ctx, s.presentAccountIDs(locationID), client.AccountID)
	return client.WriteJSON(wsproto.LocationState{
		Type:      "LOCATION_STATE",
		Location:  loc,
		Present:   present,
		Narration: narration,
	})
}

// broadcastPresenceUpdate refreshes everyone else standing at a location
// when someone new arrives or leaves, so their "who's here" list stays
// accurate without them having to travel away and back.
func (s *server) broadcastPresenceUpdate(ctx context.Context, locationID models.LocationID, exceptAccountID string) {
	for _, accountID := range s.presentAccountIDs(locationID) {
		if accountID == exceptAccountID {
			continue
		}
		loc := world.Locations[locationID]
		present := s.buildPresentAccounts(ctx, s.presentAccountIDs(locationID), accountID)
		s.hub.SendTo(accountID, wsproto.LocationState{Type: "LOCATION_STATE", Location: loc, Present: present})
	}
}

// handleTravel moves the account to a connected location. Movement here is
// a graph edge, not a tile step — see internal/world's package doc for why
// the overworld stopped being a coordinate grid.
func (s *server) handleTravel(ctx context.Context, client *chat.Client, payload json.RawMessage) error {
	var p wsproto.TravelPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return err
	}
	account, err := s.store.GetAccount(ctx, client.AccountID)
	if err != nil {
		return err
	}
	if !world.CanTravel(account.LocationID, p.ToLocationID) {
		return wsError("you can't get there from here")
	}
	previousLocation := account.LocationID
	if err := s.store.UpdateLocation(ctx, client.AccountID, p.ToLocationID); err != nil {
		return err
	}
	s.setPresence(client.AccountID, p.ToLocationID)
	s.broadcastPresenceUpdate(ctx, previousLocation, client.AccountID)

	characterName := s.activeCharacterName(ctx, account)
	loc := world.Locations[p.ToLocationID]
	line := narrator.EnterLocation(characterName, loc.Name, loc.Description)
	s.sendNarration(client.AccountID, "", line)

	if err := s.sendLocationState(ctx, client, p.ToLocationID, line); err != nil {
		return err
	}
	s.broadcastPresenceUpdate(ctx, p.ToLocationID, client.AccountID)

	sync, err := s.stateSync(ctx, client.AccountID)
	if err != nil {
		return err
	}
	return client.WriteJSON(sync)
}
