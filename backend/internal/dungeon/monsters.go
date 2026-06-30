package dungeon

import "dnd5e-web/backend/internal/models"

// SRDLowCRMonsters are SRD-derived CR 1/8 - 1/4 monsters, sized for a
// level-1 party (per the CR Budget Math requirement).
var SRDLowCRMonsters = []models.Monster{
	{ID: "bandit", Name: "Bandit", ChallengeRating: 0.125, HP: 11, AttackBonus: 3, DamageDie: "1d6+1"},
	{ID: "cultist", Name: "Cultist", ChallengeRating: 0.125, HP: 9, AttackBonus: 3, DamageDie: "1d6+1"},
	{ID: "kobold", Name: "Kobold", ChallengeRating: 0.125, HP: 5, AttackBonus: 4, DamageDie: "1d4+2"},
	{ID: "giant-rat", Name: "Giant Rat", ChallengeRating: 0.125, HP: 7, AttackBonus: 4, DamageDie: "1d4+2"},
	{ID: "wolf", Name: "Wolf", ChallengeRating: 0.25, HP: 11, AttackBonus: 4, DamageDie: "2d4+2"},
	{ID: "goblin", Name: "Goblin", ChallengeRating: 0.25, HP: 7, AttackBonus: 4, DamageDie: "1d6+2"},
}

var BossMonsters = []models.Monster{
	{ID: "bandit-captain", Name: "Bandit Captain", ChallengeRating: 2, HP: 65, AttackBonus: 5, DamageDie: "1d6+3"},
	{ID: "cult-fanatic", Name: "Cult Fanatic", ChallengeRating: 2, HP: 33, AttackBonus: 4, DamageDie: "1d6+2"},
}

// xpByCR is the SRD-style XP budget per challenge rating, used to size
// encounters rather than awarding raw monster counts.
var xpByCR = map[float64]int{
	0.125: 25,
	0.25:  50,
	0.5:   100,
	1:     200,
	2:     450,
}

const (
	easyXPBudget   = 25
	mediumXPBudget = 50
)
