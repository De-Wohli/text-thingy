package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/models"
)

func (s *Store) CreateAccount(ctx context.Context, displayName string) (models.Account, error) {
	account := models.Account{
		ID:          uuid.NewString(),
		DisplayName: displayName,
		Honor:       0,
		Gold:        50,
		Coordinate:  models.Coordinate{X: 10, Y: 7},
	}
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO accounts (id, display_name, honor, gold, coord_x, coord_y)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, account.ID, account.DisplayName, account.Honor, account.Gold, account.Coordinate.X, account.Coordinate.Y)
	if err != nil {
		return models.Account{}, fmt.Errorf("create account: %w", err)
	}
	return account, nil
}

func (s *Store) GetAccount(ctx context.Context, id string) (models.Account, error) {
	var a models.Account
	var activeCharacterID *string
	var partyID *string
	err := s.Pool.QueryRow(ctx, `
		SELECT id, display_name, honor, gold, active_character_id, coord_x, coord_y, party_id
		FROM accounts WHERE id = $1
	`, id).Scan(&a.ID, &a.DisplayName, &a.Honor, &a.Gold, &activeCharacterID, &a.Coordinate.X, &a.Coordinate.Y, &partyID)
	if err != nil {
		return models.Account{}, fmt.Errorf("get account: %w", err)
	}
	a.ActiveCharacterID = activeCharacterID
	a.PartyID = partyID
	return a, nil
}

func (s *Store) UpdateCoordinate(ctx context.Context, accountID string, coord models.Coordinate) error {
	_, err := s.Pool.Exec(ctx, `UPDATE accounts SET coord_x = $2, coord_y = $3 WHERE id = $1`, accountID, coord.X, coord.Y)
	return err
}

func (s *Store) SetActiveCharacter(ctx context.Context, accountID, characterID string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE accounts SET active_character_id = $2 WHERE id = $1`, accountID, characterID)
	return err
}

// ApplyHonorDelta atomically updates an account's honor and appends an
// audit row, returning the clamped resulting score.
func (s *Store) ApplyHonorDelta(ctx context.Context, accountID string, delta int, typology models.ChoiceTypology, reason string) (int, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var newHonor int
	err = tx.QueryRow(ctx, `
		UPDATE accounts SET honor = GREATEST(-100, LEAST(100, honor + $2))
		WHERE id = $1
		RETURNING honor
	`, accountID, delta).Scan(&newHonor)
	if err != nil {
		return 0, fmt.Errorf("apply honor delta: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO honor_log (account_id, delta, typology, reason) VALUES ($1, $2, $3, $4)
	`, accountID, delta, string(typology), reason); err != nil {
		return 0, fmt.Errorf("record honor log: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return newHonor, nil
}

func (s *Store) AddGold(ctx context.Context, accountID string, amount int) error {
	_, err := s.Pool.Exec(ctx, `UPDATE accounts SET gold = gold + $2 WHERE id = $1`, accountID, amount)
	return err
}

func (s *Store) CreateCharacter(ctx context.Context, c models.Character) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO characters
			(id, account_id, name, race_id, class_id, level, status, hp_current, hp_max,
			 ability_str, ability_dex, ability_con, ability_int, ability_wis, ability_cha)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, c.ID, c.AccountID, c.Name, c.RaceID, c.ClassID, c.Level, c.Status, c.HPCurrent, c.HPMax,
		c.AbilityScores.Str, c.AbilityScores.Dex, c.AbilityScores.Con, c.AbilityScores.Int, c.AbilityScores.Wis, c.AbilityScores.Cha)
	if err != nil {
		return fmt.Errorf("create character: %w", err)
	}
	return nil
}

func (s *Store) ListCharacters(ctx context.Context, accountID string) ([]models.Character, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, account_id, name, race_id, class_id, level, status, hp_current, hp_max,
		       ability_str, ability_dex, ability_con, ability_int, ability_wis, ability_cha, created_at
		FROM characters WHERE account_id = $1 ORDER BY created_at ASC
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list characters: %w", err)
	}
	defer rows.Close()

	var characters []models.Character
	for rows.Next() {
		var c models.Character
		if err := rows.Scan(&c.ID, &c.AccountID, &c.Name, &c.RaceID, &c.ClassID, &c.Level, &c.Status,
			&c.HPCurrent, &c.HPMax, &c.AbilityScores.Str, &c.AbilityScores.Dex, &c.AbilityScores.Con,
			&c.AbilityScores.Int, &c.AbilityScores.Wis, &c.AbilityScores.Cha, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan character: %w", err)
		}
		characters = append(characters, c)
	}
	return characters, rows.Err()
}

func (s *Store) SaveDungeon(ctx context.Context, d models.Dungeon) error {
	grid, err := json.Marshal(d.Grid)
	if err != nil {
		return err
	}
	rooms, err := json.Marshal(d.Rooms)
	if err != nil {
		return err
	}
	encounters, err := json.Marshal(d.Encounters)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO dungeons (id, party_id, grid, rooms, encounters, resolved)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, d.ID, d.PartyID, grid, rooms, encounters, d.Resolved)
	if err != nil {
		return fmt.Errorf("save dungeon: %w", err)
	}
	return nil
}

func (s *Store) ResolveDungeon(ctx context.Context, id string) error {
	_, err := s.Pool.Exec(ctx, `UPDATE dungeons SET resolved = true, resolved_at = now() WHERE id = $1`, id)
	return err
}
