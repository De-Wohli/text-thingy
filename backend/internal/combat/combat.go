// Package combat resolves a dungeon room encounter using simplified-but-real
// 5e mechanics: d20 attack rolls against Armor Class, SRD damage dice,
// proficiency bonus, and ability modifiers. It deliberately does not model
// every SRD subsystem (no spell slot expenditure, conditions, or initiative
// order beyond "character swings, then every surviving monster swings back")
// — see the package-level docs in README.md for the full list of
// simplifications.
package combat

import "dnd5e-web/backend/internal/models"

// AttackRoll is one swing of a weapon or spell attack, recorded so the
// gateway can narrate it and the client can render a combat log.
type AttackRoll struct {
	Attacker    string `json:"attacker"`
	Target      string `json:"target"`
	D20         int    `json:"d20"`
	AttackBonus int    `json:"attackBonus"`
	Total       int    `json:"total"`
	TargetAC    int    `json:"targetAc"`
	Hit         bool   `json:"hit"`
	Critical    bool   `json:"critical"`
	Damage      int    `json:"damage"`
}

// Result is the outcome of a fully-resolved encounter (the room's monsters
// fight to the death against a single round-robin, alternating with the
// character, until one side is defeated).
type Result struct {
	Rounds            []AttackRoll `json:"rounds"`
	Victory           bool         `json:"victory"`
	CharacterHPBefore int          `json:"characterHpBefore"`
	CharacterHPAfter  int          `json:"characterHpAfter"`
	MonstersDefeated  []string     `json:"monstersDefeated"`
}

// classProfile is a simplified SRD-equipment assumption: Fighters swing a
// melee weapon (Strength), Wizards cast a damage cantrip (Intelligence,
// no ability modifier added to cantrip damage per SRD rules).
type classProfile struct {
	weaponName         string
	damageDie          string
	addAbilityToDamage bool
}

var profiles = map[models.ClassID]classProfile{
	models.ClassFighter: {weaponName: "longsword", damageDie: "1d8", addAbilityToDamage: true},
	models.ClassWizard:  {weaponName: "fire bolt", damageDie: "1d10", addAbilityToDamage: false},
}

func attackAbilityModifier(classID models.ClassID, scores models.AbilityScores) int {
	if classID == models.ClassWizard {
		return models.AbilityModifier(scores.Int)
	}
	return models.AbilityModifier(scores.Str)
}

// WeaponName returns the flavor name of the character's class-appropriate
// weapon/cantrip, for narration.
func WeaponName(classID models.ClassID) string {
	if p, ok := profiles[classID]; ok {
		return p.weaponName
	}
	return "weapon"
}

// Resolve fights a character against every monster in a room. Encounters
// can be lost — if the character's HP would drop to 0 or below, the fight
// ends in defeat — but there's no permadeath in this prototype: a losing
// result still leaves the character stabilized (HP clamped to 1) rather
// than dead, and the caller is expected to offer a retreat-and-heal path
// rather than a game over screen.
func Resolve(character models.Character, monsters []models.Monster) Result {
	level := character.Level
	if level < 1 {
		level = 1
	}
	prof := models.ProficiencyBonusForLevel(level)
	abilityMod := attackAbilityModifier(character.ClassID, character.AbilityScores)
	attackBonus := prof + abilityMod
	cp := profiles[character.ClassID]
	damageBonus := 0
	if cp.addAbilityToDamage {
		damageBonus = abilityMod
	}

	charAC := models.ArmorClassFor(character.ClassID, models.AbilityModifier(character.AbilityScores.Dex))
	charHP := character.HPCurrent
	if charHP <= 0 {
		charHP = character.HPMax
	}

	monsterHP := make([]int, len(monsters))
	for i, m := range monsters {
		monsterHP[i] = m.HP
	}

	// Initialized (not nil) so json.Marshal produces `[]` instead of
	// `null` for a vacuous (zero-monster) encounter — see
	// dungeon.pickEncounterForLevel1's doc comment for why that matters.
	rounds := []AttackRoll{}
	defeated := make([]string, 0, len(monsters))
	aliveCount := len(monsters)

	for aliveCount > 0 && charHP > 0 {
		targetIdx := -1
		for i, hp := range monsterHP {
			if hp > 0 {
				targetIdx = i
				break
			}
		}
		if targetIdx == -1 {
			break
		}

		d20 := RollD20()
		total := d20 + attackBonus
		crit := d20 == 20
		hit := crit || total >= monsters[targetIdx].ArmorClass
		damage := 0
		if hit {
			damage = RollDice(cp.damageDie) + damageBonus
			if crit {
				damage += RollDice(cp.damageDie)
			}
			if damage < 0 {
				damage = 0
			}
			monsterHP[targetIdx] -= damage
			if monsterHP[targetIdx] <= 0 {
				aliveCount--
				defeated = append(defeated, monsters[targetIdx].Name)
			}
		}
		rounds = append(rounds, AttackRoll{
			Attacker: character.Name, Target: monsters[targetIdx].Name,
			D20: d20, AttackBonus: attackBonus, Total: total, TargetAC: monsters[targetIdx].ArmorClass,
			Hit: hit, Critical: crit, Damage: damage,
		})

		if aliveCount == 0 {
			break
		}

		for i, m := range monsters {
			if monsterHP[i] <= 0 || charHP <= 0 {
				continue
			}
			md20 := RollD20()
			mtotal := md20 + m.AttackBonus
			mcrit := md20 == 20
			mhit := mcrit || mtotal >= charAC
			mdamage := 0
			if mhit {
				mdamage = RollDice(m.DamageDie)
				if mcrit {
					mdamage += RollDice(m.DamageDie)
				}
				charHP -= mdamage
			}
			rounds = append(rounds, AttackRoll{
				Attacker: m.Name, Target: character.Name,
				D20: md20, AttackBonus: m.AttackBonus, Total: mtotal, TargetAC: charAC,
				Hit: mhit, Critical: mcrit, Damage: mdamage,
			})
		}
	}

	victory := aliveCount == 0 && charHP > 0
	finalHP := charHP
	if finalHP < 1 {
		finalHP = 1
	}

	return Result{
		Rounds:            rounds,
		Victory:           victory,
		CharacterHPBefore: character.HPCurrent,
		CharacterHPAfter:  finalHP,
		MonstersDefeated:  defeated,
	}
}
