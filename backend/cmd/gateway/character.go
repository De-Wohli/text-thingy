package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/models"
)

// buildCharacter mirrors frontend/src/engine/account.ts createCharacter: base
// ability scores of 10, race bonuses applied, HP derived from the class hit
// die plus the Constitution modifier. Keep the two in sync.
func buildCharacter(accountID, name string, raceID models.RaceID, classID models.ClassID) (models.Character, error) {
	race, ok := models.Races[raceID]
	if !ok {
		return models.Character{}, fmt.Errorf("unknown race: %s", raceID)
	}
	class, ok := models.Classes[classID]
	if !ok {
		return models.Character{}, fmt.Errorf("unknown class: %s", classID)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return models.Character{}, fmt.Errorf("character name is required")
	}

	scores := models.AbilityScores{Str: 10, Dex: 10, Con: 10, Int: 10, Wis: 10, Cha: 10}
	for ability, bonus := range race.AbilityBonuses {
		switch ability {
		case "str":
			scores.Str += bonus
		case "dex":
			scores.Dex += bonus
		case "con":
			scores.Con += bonus
		case "int":
			scores.Int += bonus
		case "wis":
			scores.Wis += bonus
		case "cha":
			scores.Cha += bonus
		}
	}

	conModifier := floorDiv(scores.Con-10, 2)
	maxHP := class.HitDie + conModifier

	return models.Character{
		ID:            uuid.NewString(),
		AccountID:     accountID,
		Name:          name,
		RaceID:        raceID,
		ClassID:       classID,
		Level:         1,
		Status:        models.StatusIdle,
		HPCurrent:     maxHP,
		HPMax:         maxHP,
		AbilityScores: scores,
		CreatedAt:     time.Now(),
	}, nil
}

// floorDiv divides toward negative infinity, matching JS's Math.floor(a/b)
// semantics used on the frontend for ability modifiers.
func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}
