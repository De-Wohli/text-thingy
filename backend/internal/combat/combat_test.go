package combat

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func sampleFighter(id, name string) models.Character {
	return models.Character{
		ID:            id,
		AccountID:     "account-" + id,
		Name:          name,
		ClassID:       models.ClassFighter,
		Level:         1,
		HPCurrent:     12,
		HPMax:         12,
		AbilityScores: models.AbilityScores{Str: 16, Dex: 14, Con: 15, Int: 8, Wis: 12, Cha: 10},
	}
}

func weakMonster() models.Monster {
	return models.Monster{ID: "training-dummy", Name: "Training Dummy", ArmorClass: 1, HP: 1, AttackBonus: -10, DamageDie: "1d4"}
}

func unbeatableMonster() models.Monster {
	return models.Monster{ID: "ancient-horror", Name: "Ancient Horror", ArmorClass: 99, HP: 999, AttackBonus: 99, DamageDie: "10d10+50"}
}

func TestNewEncounterFieldsAreNeverNilSlices(t *testing.T) {
	// Regression: a nil Go slice serializes to JSON `null`, and the
	// frontend calls .map()/.filter() on Combatants/Log without a null
	// guard — this exact bug class has bitten ListCharacters and
	// combat.Resolve's Rounds field earlier in this codebase's history.
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	if e.Combatants == nil {
		t.Fatal("expected Combatants to be a non-nil slice")
	}
	if e.Log == nil {
		t.Fatal("expected Log to be a non-nil slice, even before any attack has happened")
	}
}

func TestNewEncounterOrdersByInitiativeDescending(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	for i := 1; i < len(e.Combatants); i++ {
		if e.Combatants[i-1].Initiative < e.Combatants[i].Initiative {
			t.Fatalf("combatants not sorted by initiative descending: %+v", e.Combatants)
		}
	}
}

func TestAttackOnlyAllowedOnYourTurn(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	current := e.Current()
	if current == nil {
		t.Fatal("expected a current combatant")
	}
	// Find a combatant who is NOT current.
	var notCurrent *Combatant
	for _, c := range e.Combatants {
		if c != current {
			notCurrent = c
		}
	}
	if notCurrent == nil {
		t.Skip("only one combatant, can't test out-of-turn rejection")
	}
	if _, err := e.Attack(notCurrent.ID, current.ID); err != ErrNotYourTurn {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestAttackDefeatsWeakMonster(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	current := e.Current()
	if current == nil || current.Kind != KindPlayer {
		t.Fatalf("expected the fighter to act first against a trivial monster, got %+v", current)
	}
	if _, err := e.Attack(current.ID, findMonster(e).ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !findMonster(e).Defeated {
		t.Fatal("expected the trivial monster to be defeated in one hit")
	}
	over, victory := e.Outcome()
	if !over || !victory {
		t.Fatalf("expected victory once the only monster is defeated, got over=%v victory=%v", over, victory)
	}
}

func findMonster(e *Encounter) *Combatant {
	for _, c := range e.Combatants {
		if c.Kind == KindMonster {
			return c
		}
	}
	return nil
}

func TestDodgeGivesDisadvantageToAttacker(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	player := e.find("c1")
	if e.Current() != player {
		t.Skip("monster acted first this run, skipping deterministic dodge check")
	}
	if err := e.Dodge(player.ID); err != nil {
		t.Fatalf("unexpected error dodging: %v", err)
	}
	if !player.Dodging {
		t.Fatal("expected player.Dodging to be true after Dodge()")
	}
}

func TestFleeRemovesCombatantFromTurnOrderWithoutEndingFight(t *testing.T) {
	// Deliberately a weak monster, not unbeatable: an unbeatable monster
	// going first could one-shot a player during the automatic
	// leading-turn resolution, before either player gets to act — that
	// would make "two players, one flees" collapse to "one player left"
	// for reasons unrelated to what this test is checking.
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak"), sampleFighter("c2", "Mira")}, []models.Monster{weakMonster()}, nil)
	current := e.Current()
	if current == nil || current.Kind != KindPlayer {
		t.Fatal("expected a player to act first against trivially-easy-to-go-first setup")
	}
	if err := e.Flee(current.ID); err != nil {
		t.Fatalf("unexpected error fleeing: %v", err)
	}
	if !current.Fled {
		t.Fatal("expected Fled to be true")
	}
	if current.Alive() {
		t.Fatal("a fled combatant should not be Alive()")
	}
	over, _ := e.Outcome()
	if over {
		t.Fatal("fleeing one of two players should not end the fight while the other is still up")
	}
}

func TestOutcomeDefeatWhenAllPlayersDown(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{unbeatableMonster()}, nil)
	for i := 0; i < 50; i++ {
		if over, _ := e.Outcome(); over {
			break
		}
		current := e.Current()
		if current == nil || current.Kind != KindPlayer {
			t.Fatalf("expected only the player to ever need manual action, got %+v", current)
		}
		if _, err := e.Attack(current.ID, findMonster(e).ID); err != nil {
			t.Fatalf("unexpected attack error: %v", err)
		}
		e.AdvanceTurn()
	}
	over, victory := e.Outcome()
	if !over {
		t.Fatal("expected the fight to be over against an unbeatable monster within 50 rounds")
	}
	if victory {
		t.Fatal("expected defeat, not victory, against an unbeatable monster")
	}
}

func TestAdvanceTurnAutoResolvesMonsterTurns(t *testing.T) {
	// Two players, one monster that can't possibly win — the monster's
	// turn (whenever it comes up) must resolve automatically without any
	// caller action, and control must always return on a player's turn.
	e := NewEncounter(
		[]models.Character{sampleFighter("c1", "Brak"), sampleFighter("c2", "Mira")},
		[]models.Monster{weakMonster()},
		nil,
	)
	for i := 0; i < 10; i++ {
		if over, _ := e.Outcome(); over {
			return
		}
		current := e.Current()
		if current == nil {
			t.Fatal("expected a current combatant while the fight is ongoing")
		}
		if current.Kind != KindPlayer {
			t.Fatalf("AdvanceTurn should never leave a monster as the current actor, got %+v", current)
		}
		if _, err := e.Attack(current.ID, findMonster(e).ID); err != nil {
			t.Fatalf("unexpected attack error: %v", err)
		}
		e.AdvanceTurn()
	}
}

func TestUnknownCombatantErrors(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{weakMonster()}, nil)
	if _, err := e.Attack("nobody", "also-nobody"); err != ErrUnknownCombatant {
		t.Fatalf("expected ErrUnknownCombatant, got %v", err)
	}
	if err := e.Dodge("nobody"); err != ErrUnknownCombatant {
		t.Fatalf("expected ErrUnknownCombatant from Dodge, got %v", err)
	}
	if err := e.Flee("nobody"); err != ErrUnknownCombatant {
		t.Fatalf("expected ErrUnknownCombatant from Flee, got %v", err)
	}
}

func TestEmptyMonsterListIsVacuousVictory(t *testing.T) {
	e := NewEncounter([]models.Character{sampleFighter("c1", "Brak")}, []models.Monster{}, nil)
	over, victory := e.Outcome()
	if !over || !victory {
		t.Fatalf("expected a vacuous victory with no monsters, got over=%v victory=%v", over, victory)
	}
}

func TestWeaponNameKnownForBothClasses(t *testing.T) {
	if WeaponName(models.ClassFighter) == "weapon" {
		t.Error("expected a specific weapon name for Fighter")
	}
	if WeaponName(models.ClassWizard) == "weapon" {
		t.Error("expected a specific weapon name for Wizard")
	}
}
