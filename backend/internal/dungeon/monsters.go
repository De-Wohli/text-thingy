package dungeon

import "dnd5e-web/backend/internal/models"

// SRDLowCRMonsters are SRD-derived CR 1/8 - 1/4 monsters, sized for a
// level-1 party (per the CR Budget Math requirement). AC/HP/AttackBonus/
// DamageDie are taken from the SRD statblocks so combat resolution
// (internal/combat) rolls against real numbers.
var SRDLowCRMonsters = []models.Monster{
	{ID: "bandit", Name: "Bandit", ChallengeRating: 0.125, ArmorClass: 12, HP: 11, AttackBonus: 3, DamageDie: "1d6+1"},
	{ID: "cultist", Name: "Cultist", ChallengeRating: 0.125, ArmorClass: 11, HP: 9, AttackBonus: 3, DamageDie: "1d6+1"},
	{ID: "kobold", Name: "Kobold", ChallengeRating: 0.125, ArmorClass: 12, HP: 5, AttackBonus: 4, DamageDie: "1d4+2"},
	{ID: "giant-rat", Name: "Giant Rat", ChallengeRating: 0.125, ArmorClass: 12, HP: 7, AttackBonus: 4, DamageDie: "1d4+2"},
	{ID: "wolf", Name: "Wolf", ChallengeRating: 0.25, ArmorClass: 13, HP: 11, AttackBonus: 4, DamageDie: "2d4+2"},
	{ID: "goblin", Name: "Goblin", ChallengeRating: 0.25, ArmorClass: 15, HP: 7, AttackBonus: 4, DamageDie: "1d6+2"},
}

// BossMonsters are deliberately tuned well below their real SRD CR (the
// Bandit Captain and Cult Fanatic are CR 2 in the book). 5e's CR math
// assumes a 4-character party splitting the action economy; this prototype
// has no party-formation flow yet (see README "Known simplifications"), so
// every dungeon is fought solo. A true CR 2 boss is mathematically
// unwinnable for one level-1 character (65 HP at a ~50% hit chance vs. a
// character who goes down in 2-3 hits) — these numbers are sized instead
// for "dangerous but winnable solo fight", roughly CR 1/2 toughness.
var BossMonsters = []models.Monster{
	{ID: "bandit-captain", Name: "Bandit Captain", ChallengeRating: 0.5, ArmorClass: 13, HP: 18, AttackBonus: 3, DamageDie: "1d6+1"},
	{ID: "cult-fanatic", Name: "Cult Fanatic", ChallengeRating: 0.5, ArmorClass: 12, HP: 15, AttackBonus: 3, DamageDie: "1d6+1"},
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
