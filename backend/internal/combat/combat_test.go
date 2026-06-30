package combat

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func sampleFighter() models.Character {
	return models.Character{
		ID:            "char-1",
		Name:          "Brak",
		ClassID:       models.ClassFighter,
		Level:         1,
		HPCurrent:     10,
		HPMax:         10,
		AbilityScores: models.AbilityScores{Str: 16, Dex: 12, Con: 14, Int: 10, Wis: 10, Cha: 10},
	}
}

func weakMonster() models.Monster {
	// Trivially easy to hit and kill so victory is reachable in bounded time.
	return models.Monster{ID: "training-dummy", Name: "Training Dummy", ArmorClass: 1, HP: 1, AttackBonus: -10, DamageDie: "1d4"}
}

func TestResolveAlwaysTerminates(t *testing.T) {
	character := sampleFighter()
	monsters := []models.Monster{weakMonster(), weakMonster()}
	for i := 0; i < 50; i++ {
		result := Resolve(character, monsters)
		if len(result.Rounds) == 0 {
			t.Fatal("expected at least one attack roll")
		}
	}
}

func TestResolveVictoryDefeatsAllMonsters(t *testing.T) {
	character := sampleFighter()
	monsters := []models.Monster{weakMonster()}
	result := Resolve(character, monsters)
	if !result.Victory {
		t.Fatalf("expected victory against a trivial monster, got defeat: %+v", result)
	}
	if len(result.MonstersDefeated) != 1 {
		t.Fatalf("expected 1 monster defeated, got %d", len(result.MonstersDefeated))
	}
}

func TestResolveDefeatClampsHPToOneNotZero(t *testing.T) {
	character := sampleFighter()
	character.HPCurrent = 10
	character.HPMax = 10
	// Overwhelming monster: impossible to hit (AC 99), guaranteed to hit back.
	overwhelming := models.Monster{ID: "ancient-horror", Name: "Ancient Horror", ArmorClass: 99, HP: 999, AttackBonus: 99, DamageDie: "10d10+50"}
	result := Resolve(character, []models.Monster{overwhelming})
	if result.Victory {
		t.Fatal("expected defeat against an unbeatable monster")
	}
	if result.CharacterHPAfter != 1 {
		t.Fatalf("expected HP clamped to 1 on defeat, got %d", result.CharacterHPAfter)
	}
}

func TestResolveZeroHPCharacterStartsAtFullHP(t *testing.T) {
	character := sampleFighter()
	character.HPCurrent = 0 // shouldn't happen in practice, but guard against it
	monsters := []models.Monster{weakMonster()}
	result := Resolve(character, monsters)
	if result.CharacterHPBefore != 0 {
		t.Fatalf("expected CharacterHPBefore to record the input value, got %d", result.CharacterHPBefore)
	}
	if !result.Victory {
		t.Fatal("expected a 0-HP character to be treated as full HP for the fight, not an instant loss")
	}
}

func TestResolveEmptyEncounterIsVacuousVictoryWithNonNilRounds(t *testing.T) {
	character := sampleFighter()
	result := Resolve(character, []models.Monster{})
	if !result.Victory {
		t.Fatal("expected an empty encounter to resolve as an automatic victory")
	}
	if result.Rounds == nil {
		t.Fatal("expected Rounds to be an empty slice, not nil — nil serializes to JSON null and crashes the frontend's .map()")
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
