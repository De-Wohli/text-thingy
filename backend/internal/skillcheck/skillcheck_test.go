package skillcheck

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func sampleFighter() models.Character {
	return models.Character{
		ClassID:       models.ClassFighter,
		Level:         1,
		AbilityScores: models.AbilityScores{Str: 15, Dex: 13, Con: 14, Int: 8, Wis: 12, Cha: 10},
	}
}

func TestRollUsesGoverningAbilityModifier(t *testing.T) {
	character := sampleFighter() // Wis 12 -> modifier +1
	for i := 0; i < 50; i++ {
		result := Roll(character, models.SkillPerception, 12)
		if result.AbilityModifier != 1 {
			t.Fatalf("expected Wis-based modifier +1 for Perception, got %d", result.AbilityModifier)
		}
		if result.Total != result.D20+result.AbilityModifier+result.ProficiencyBonus {
			t.Fatalf("Total should equal D20+AbilityModifier+ProficiencyBonus, got %+v", result)
		}
	}
}

func TestRollAppliesProficiencyBonusOnlyWhenProficient(t *testing.T) {
	character := sampleFighter() // Fighter is proficient in Athletics and Perception, not Arcana
	prof := Roll(character, models.SkillAthletics, 10)
	if !prof.Proficient || prof.ProficiencyBonus == 0 {
		t.Fatalf("expected Fighter to be proficient in Athletics with a nonzero bonus, got %+v", prof)
	}
	notProf := Roll(character, models.SkillArcana, 10)
	if notProf.Proficient || notProf.ProficiencyBonus != 0 {
		t.Fatalf("expected Fighter to not be proficient in Arcana, got %+v", notProf)
	}
}

func TestRollSuccessMatchesTotalVsDC(t *testing.T) {
	character := sampleFighter()
	for i := 0; i < 100; i++ {
		result := Roll(character, models.SkillAthletics, 12)
		if result.Success != (result.Total >= result.DC) {
			t.Fatalf("Success should be Total >= DC, got %+v", result)
		}
	}
}

func TestDCForUnknownContextHasASensibleDefault(t *testing.T) {
	if dc := DCFor(models.RoomHallway, "something-made-up"); dc <= 0 {
		t.Fatalf("expected a positive default DC, got %d", dc)
	}
}

func TestDCForBossRoomIsHarder(t *testing.T) {
	hallwayDC := DCFor(models.RoomHallway, "search-for-traps")
	bossDC := DCFor(models.RoomBoss, "search-for-traps")
	if bossDC <= hallwayDC {
		t.Fatalf("expected boss room DC (%d) to be harder than hallway DC (%d)", bossDC, hallwayDC)
	}
}
