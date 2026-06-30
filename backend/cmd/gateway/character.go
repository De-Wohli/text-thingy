package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"dnd5e-web/backend/internal/models"
)

// standardArrayByClass assigns the 5e "standard array" (15, 14, 13, 12, 10,
// 8) to abilities by class priority, before racial bonuses — this is what
// makes a level-1 Fighter actually combat-viable (internal/combat rolls
// against these numbers) rather than the flat 10-across-the-board a
// placeholder chargen would produce.
var standardArrayByClass = map[models.ClassID]models.AbilityScores{
	models.ClassFighter: {Str: 15, Dex: 13, Con: 14, Int: 8, Wis: 12, Cha: 10},
	models.ClassWizard:  {Str: 8, Dex: 13, Con: 14, Int: 15, Wis: 12, Cha: 10},
}

// buildCharacter mirrors frontend/src/engine/account.ts createCharacter:
// standard-array ability scores by class, race bonuses applied on top, HP
// derived from the class hit die plus the Constitution modifier. Keep the
// two in sync.
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

	scores, ok := standardArrayByClass[classID]
	if !ok {
		scores = models.AbilityScores{Str: 10, Dex: 10, Con: 10, Int: 10, Wis: 10, Cha: 10}
	}
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

	maxHP := class.HitDie + models.AbilityModifier(scores.Con)

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
