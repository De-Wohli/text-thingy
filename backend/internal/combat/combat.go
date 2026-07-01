// Package combat is a turn-based 5e encounter engine: real initiative
// order, one action per combatant per turn, monster turns auto-resolved
// (this is an automated-DM table, not a human controlling monsters), and a
// running log the gateway narrates and broadcasts to the whole party.
//
// It deliberately does not model every SRD subsystem: one attack per turn
// (no multiattack/bonus actions), no spell slot economy (a Wizard's "Attack"
// reuses its damage cantrip every turn), and no saving throws. Losing a
// fight has no permadeath — see backend/cmd/gateway/combat.go, which heals
// a defeated party after a retreat rather than ending the game.
package combat

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"

	"dnd5e-web/backend/internal/models"
)

var (
	ErrUnknownCombatant = errors.New("unknown combatant")
	ErrNotYourTurn      = errors.New("it is not that combatant's turn")
	ErrTargetDown       = errors.New("target is already down")
)

type CombatantKind string

const (
	KindPlayer  CombatantKind = "player"
	KindMonster CombatantKind = "monster"
)

// Combatant is one participant in an Encounter — a party member's active
// character or one monster from the room's encounter.
type Combatant struct {
	ID          string        `json:"id"`
	Kind        CombatantKind `json:"kind"`
	AccountID   string        `json:"accountId,omitempty"` // players only — authorizes whose turn it is
	Name        string        `json:"name"`
	Initiative  int           `json:"initiative"`
	HP          int           `json:"hp"`
	MaxHP       int           `json:"maxHp"`
	AC          int           `json:"ac"`
	AttackBonus int           `json:"attackBonus"`
	DamageDie   string        `json:"-"`
	DamageBonus int           `json:"-"`
	Dodging     bool          `json:"dodging"`
	Fled        bool          `json:"fled"`
	Defeated    bool          `json:"defeated"`
}

func (c *Combatant) Alive() bool { return !c.Defeated && !c.Fled }

// AttackRoll is one swing, recorded so the gateway can narrate it and the
// client can render a combat log.
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

// Encounter is a single room's fight in progress.
type Encounter struct {
	Combatants []*Combatant `json:"combatants"`
	TurnIndex  int          `json:"turnIndex"`
	Round      int          `json:"round"`
	Log        []AttackRoll `json:"log"`
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

// WeaponName returns the flavor name of a class's weapon/cantrip, for narration.
func WeaponName(classID models.ClassID) string {
	if p, ok := profiles[classID]; ok {
		return p.weaponName
	}
	return "weapon"
}

func characterCombatant(c models.Character) *Combatant {
	level := c.Level
	if level < 1 {
		level = 1
	}
	prof := models.ProficiencyBonusForLevel(level)
	abilityMod := attackAbilityModifier(c.ClassID, c.AbilityScores)
	cp := profiles[c.ClassID]
	damageBonus := 0
	if cp.addAbilityToDamage {
		damageBonus = abilityMod
	}
	hp := c.HPCurrent
	if hp <= 0 {
		hp = c.HPMax
	}
	return &Combatant{
		ID:          c.ID,
		Kind:        KindPlayer,
		AccountID:   c.AccountID,
		Name:        c.Name,
		Initiative:  RollD20() + models.AbilityModifier(c.AbilityScores.Dex),
		HP:          hp,
		MaxHP:       c.HPMax,
		AC:          models.ArmorClassFor(c.ClassID, models.AbilityModifier(c.AbilityScores.Dex)),
		AttackBonus: prof + abilityMod,
		DamageDie:   cp.damageDie,
		DamageBonus: damageBonus,
	}
}

func monsterCombatant(idx int, m models.Monster) *Combatant {
	return &Combatant{
		ID:          fmt.Sprintf("monster-%d", idx),
		Kind:        KindMonster,
		Name:        m.Name,
		Initiative:  RollD20(), // monsters don't track a separate Dex score in this simplified model
		HP:          m.HP,
		MaxHP:       m.HP,
		AC:          m.ArmorClass,
		AttackBonus: m.AttackBonus,
		DamageDie:   m.DamageDie,
	}
}

// NewEncounter rolls initiative for every character and monster, sorts
// descending, and resolves any monster turns that land before the first
// player in the order (this is an automated DM — nothing waits on a human
// to control a monster).
func NewEncounter(characters []models.Character, monsters []models.Monster) *Encounter {
	// Combatants/Log start as non-nil empty slices, not the zero-value nil
	// — a nil slice serializes to JSON `null`, and the frontend calls
	// .map()/.filter() on both without a null guard. This exact bug class
	// has bitten this codebase twice before (ListCharacters, combat.Resolve's
	// Rounds); third time's the rule, not the exception.
	e := &Encounter{Round: 1, Combatants: []*Combatant{}, Log: []AttackRoll{}}
	for _, c := range characters {
		e.Combatants = append(e.Combatants, characterCombatant(c))
	}
	for i, m := range monsters {
		e.Combatants = append(e.Combatants, monsterCombatant(i, m))
	}
	sort.SliceStable(e.Combatants, func(i, j int) bool {
		return e.Combatants[i].Initiative > e.Combatants[j].Initiative
	})
	e.settleAutomaticTurns()
	return e
}

func (e *Encounter) find(id string) *Combatant {
	for _, c := range e.Combatants {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// Current returns whoever's turn it is, or nil if the encounter has no
// combatants left to act (should only happen once Outcome reports over).
func (e *Encounter) Current() *Combatant {
	if len(e.Combatants) == 0 || e.TurnIndex < 0 || e.TurnIndex >= len(e.Combatants) {
		return nil
	}
	return e.Combatants[e.TurnIndex]
}

func (e *Encounter) rollAttack(actor, target *Combatant) AttackRoll {
	d20 := RollD20()
	if target.Dodging {
		// SRD Dodge: attacks against the dodger have disadvantage.
		if second := RollD20(); second < d20 {
			d20 = second
		}
	}
	total := d20 + actor.AttackBonus
	crit := d20 == 20
	hit := crit || total >= target.AC
	damage := 0
	if hit {
		damage = RollDice(actor.DamageDie) + actor.DamageBonus
		if crit {
			damage += RollDice(actor.DamageDie)
		}
		if damage < 0 {
			damage = 0
		}
		target.HP -= damage
		if target.HP <= 0 {
			target.HP = 0
			target.Defeated = true
		}
	}
	return AttackRoll{
		Attacker: actor.Name, Target: target.Name,
		D20: d20, AttackBonus: actor.AttackBonus, Total: total, TargetAC: target.AC,
		Hit: hit, Critical: crit, Damage: damage,
	}
}

// Attack is a player action: actorID must be the current turn's combatant.
func (e *Encounter) Attack(actorID, targetID string) (AttackRoll, error) {
	actor := e.find(actorID)
	if actor == nil {
		return AttackRoll{}, ErrUnknownCombatant
	}
	if e.Current() != actor {
		return AttackRoll{}, ErrNotYourTurn
	}
	target := e.find(targetID)
	if target == nil || !target.Alive() {
		return AttackRoll{}, ErrTargetDown
	}
	roll := e.rollAttack(actor, target)
	e.Log = append(e.Log, roll)
	return roll, nil
}

// Dodge is the SRD Dodge action: until the start of this combatant's next
// turn, attacks against them have disadvantage.
func (e *Encounter) Dodge(actorID string) error {
	actor := e.find(actorID)
	if actor == nil {
		return ErrUnknownCombatant
	}
	if e.Current() != actor {
		return ErrNotYourTurn
	}
	actor.Dodging = true
	return nil
}

// Flee removes the combatant from the turn order for the rest of this
// fight — it doesn't end the encounter for the rest of the party.
func (e *Encounter) Flee(actorID string) error {
	actor := e.find(actorID)
	if actor == nil {
		return ErrUnknownCombatant
	}
	if e.Current() != actor {
		return ErrNotYourTurn
	}
	actor.Fled = true
	return nil
}

func (e *Encounter) randomAlivePlayer() *Combatant {
	var alive []*Combatant
	for _, c := range e.Combatants {
		if c.Kind == KindPlayer && c.Alive() {
			alive = append(alive, c)
		}
	}
	if len(alive) == 0 {
		return nil
	}
	return alive[rand.Intn(len(alive))]
}

func (e *Encounter) stepIndex() {
	if len(e.Combatants) == 0 {
		return
	}
	e.TurnIndex++
	if e.TurnIndex >= len(e.Combatants) {
		e.TurnIndex = 0
		e.Round++
	}
}

// settleAutomaticTurns resolves monster turns starting at the current
// index (without advancing first) until it's an alive player's turn or
// the fight is over. Used both right after NewEncounter (a monster may
// have rolled top initiative) and at the tail of AdvanceTurn.
func (e *Encounter) settleAutomaticTurns() {
	for i := 0; i < len(e.Combatants)+1; i++ {
		if over, _ := e.Outcome(); over {
			return
		}
		current := e.Current()
		if current == nil {
			return
		}
		if !current.Alive() {
			e.stepIndex()
			continue
		}
		if current.Kind == KindPlayer {
			current.Dodging = false // "until the start of your next turn" — that's now
			return
		}
		if target := e.randomAlivePlayer(); target != nil {
			roll := e.rollAttack(current, target)
			e.Log = append(e.Log, roll)
		}
		e.stepIndex()
	}
}

// AdvanceTurn ends the current combatant's turn and settles the table up
// to the next player decision point (auto-resolving any monster turns in
// between).
func (e *Encounter) AdvanceTurn() {
	e.stepIndex()
	e.settleAutomaticTurns()
}

// Outcome reports whether the fight is over, and if so, whether the party
// won (every monster defeated/fled) or lost (every player defeated/fled).
// A fight with no combatants on one side from the start (e.g. an empty
// room) is a vacuous victory.
func (e *Encounter) Outcome() (over, victory bool) {
	anyPlayerAlive, anyMonsterAlive := false, false
	for _, c := range e.Combatants {
		if c.Kind == KindPlayer && c.Alive() {
			anyPlayerAlive = true
		}
		if c.Kind == KindMonster && c.Alive() {
			anyMonsterAlive = true
		}
	}
	if !anyMonsterAlive {
		return true, true
	}
	if !anyPlayerAlive {
		return true, false
	}
	return false, false
}
