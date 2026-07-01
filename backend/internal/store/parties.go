package store

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/models"
)

// CreateParty makes the given account the leader of a brand new party and
// returns the new party's ID. Callers are responsible for then assigning
// both the leader and the invitee to that party ID via SetPartyID.
func (s *Store) CreateParty(ctx context.Context, leaderAccountID string) (string, error) {
	id := uuid.NewString()
	_, err := s.Pool.Exec(ctx, `INSERT INTO parties (id, leader_account_id) VALUES ($1, $2)`, id, leaderAccountID)
	if err != nil {
		return "", fmt.Errorf("create party: %w", err)
	}
	return id, nil
}

// ListAccountsByPartyID returns every account currently in the given
// party, used to build the party roster and broadcast PARTY_STATE.
func (s *Store) ListAccountsByPartyID(ctx context.Context, partyID string) ([]models.Account, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, display_name, honor, gold, active_character_id, location_id, party_id
		FROM accounts WHERE party_id = $1
	`, partyID)
	if err != nil {
		return nil, fmt.Errorf("list accounts by party: %w", err)
	}
	defer rows.Close()

	accounts := []models.Account{}
	for rows.Next() {
		var a models.Account
		var activeCharacterID *string
		var pID *string
		if err := rows.Scan(&a.ID, &a.DisplayName, &a.Honor, &a.Gold, &activeCharacterID, &a.LocationID, &pID); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		a.ActiveCharacterID = activeCharacterID
		a.PartyID = pID
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
