// Package skillcheck implements 5e-style ability checks and the mechanical
// outcomes they produce. Each successful or failed check carries a typed
// Outcome that the gateway applies — either immediately (trap damage, temp
// HP) or when the room's encounter is built (initiative/attack/damage mods).
package skillcheck

import (
	"dnd5e-web/backend/internal/combat"
	"dnd5e-web/backend/internal/models"
)

// Outcome is the mechanical consequence of a skill check result.
type Outcome string

const (
	// Success outcomes
	OutcomeMonsterRemoved Outcome = "monster_removed" // Investigation: strip weakest foe
	OutcomePlayerFirst    Outcome = "player_first"    // Perception: players always go first
	OutcomeSneakAttack    Outcome = "sneak_attack"    // Stealth: free attack before initiative
	OutcomeAttackBonus    Outcome = "attack_bonus"    // Insight: +2 attack in the fight
	OutcomeDamageBonus    Outcome = "damage_bonus"    // Arcana: +1 damage throughout the fight
	OutcomeTempHP         Outcome = "temp_hp"         // Athletics: +3 HP before the fight

	// Failure outcomes
	OutcomeTrapDamage   Outcome = "trap_damage"   // Investigation: 1d4 trap springs
	OutcomeMonsterReady Outcome = "monster_ready" // Perception/Stealth: monsters alert (+2 atk)

	// No extra effect
	OutcomeNone Outcome = "none"
)

// CooldownSeconds is how long a failed check locks out the same skill+room
// combination. In a live session this prevents retry-spam while feeling
// proportionate (one minute ~ "you used up your attempt this scene").
const CooldownSeconds = 60

type Result struct {
	Skill            models.Skill `json:"skill"`
	D20              int          `json:"d20"`
	AbilityModifier  int          `json:"abilityModifier"`
	ProficiencyBonus int          `json:"proficiencyBonus"`
	Total            int          `json:"total"`
	DC               int          `json:"dc"`
	Proficient       bool         `json:"proficient"`
	Success          bool         `json:"success"`
	Outcome          Outcome      `json:"outcome"`
	OutcomeValue     int          `json:"outcomeValue"`    // magnitude: damage dealt, bonus, HP granted
	CooldownSeconds  int          `json:"cooldownSeconds"` // 0 on success, CooldownSeconds on failure
}

// outcomeFor maps each skill×success combination to its mechanical consequence
// and the numeric magnitude (0 if the effect is qualitative / applied in engine).
func outcomeFor(skill models.Skill, success bool) (Outcome, int) {
	if success {
		switch skill {
		case models.SkillInvestigation:
			return OutcomeMonsterRemoved, 0
		case models.SkillPerception:
			return OutcomePlayerFirst, 20 // +20 to all player initiatives
		case models.SkillStealth:
			return OutcomeSneakAttack, 0
		case models.SkillInsight:
			return OutcomeAttackBonus, 2 // +2 to every player attack roll
		case models.SkillArcana:
			return OutcomeDamageBonus, 1 // +1 to every player damage roll
		case models.SkillAthletics:
			return OutcomeTempHP, 3 // +3 HP added before the fight
		}
	} else {
		switch skill {
		case models.SkillInvestigation:
			// Trap springs — deal 1d4 damage to the active character.
			return OutcomeTrapDamage, combat.RollDice("1d4")
		case models.SkillPerception, models.SkillStealth:
			// Monsters heard/spotted you first — they get +2 on attacks.
			return OutcomeMonsterReady, 2
		}
	}
	return OutcomeNone, 0
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
	success := total >= dc

	outcome, outcomeValue := outcomeFor(skill, success)
	cooldown := 0
	if !success {
		cooldown = CooldownSeconds
	}

	return Result{
		Skill:            skill,
		D20:              d20,
		AbilityModifier:  abilityMod,
		ProficiencyBonus: profBonus,
		Total:            total,
		DC:               dc,
		Proficient:       proficient,
		Success:          success,
		Outcome:          outcome,
		OutcomeValue:     outcomeValue,
		CooldownSeconds:  cooldown,
	}
}

// DCFor picks a difficulty class for a contextual non-combat action.
func DCFor(roomType models.DungeonRoomType, context string) int {
	dc, ok := baseDC[context]
	if !ok {
		dc = 12
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
