// Package skillcheck implements 5e-style ability checks for the non-combat
// interactions a tabletop session expects outside of a fight — searching
// for traps, investigating a room, reading a situation — as a d20 +
// ability modifier (+ proficiency bonus, if the class is proficient in
// that skill) roll against a DC.
package skillcheck

import (
	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
)

type Result struct {
	Skill            models.Skill `json:"skill"`
	D20              int          `json:"d20"`
	AbilityModifier  int          `json:"abilityModifier"`
	ProficiencyBonus int          `json:"proficiencyBonus"`
	Total            int          `json:"total"`
	DC               int          `json:"dc"`
	Proficient       bool         `json:"proficient"`
	Success          bool         `json:"success"`
}

func Roll(character models.Character, skill models.Skill, dc int) Result {
	d20 := combat.RollD20()
	abilityMod := models.AbilityModifier(models.AbilityScoreFor(character.AbilityScores, models.SkillAbility[skill]))
	proficient := models.IsProficientInSkill(character.ClassID, skill)
	profBonus := 0
	if proficient {
		level := character.Level
		if level < 1 {
			level = 1
		}
		profBonus = models.ProficiencyBonusForLevel(level)
	}
	total := d20 + abilityMod + profBonus
	return Result{
		Skill:            skill,
		D20:              d20,
		AbilityModifier:  abilityMod,
		ProficiencyBonus: profBonus,
		Total:            total,
		DC:               dc,
		Proficient:       proficient,
		Success:          total >= dc,
	}
}

// DCFor picks a difficulty class for a contextual non-combat action. Boss
// rooms are tuned a little harder — the stakes (and the foe's cunning) are
// higher right before the final fight.
func DCFor(roomType models.DungeonRoomType, context string) int {
	dc, ok := baseDC[context]
	if !ok {
		dc = 12 // a moderate SRD-typical DC as the default for an unrecognized context
	}
	if roomType == models.RoomBoss {
		dc += 3
	}
	return dc
}

var baseDC = map[string]int{
	"search-for-traps": 12,
	"investigate":      11,
	"listen":           10,
	"read-the-room":    13,
}
