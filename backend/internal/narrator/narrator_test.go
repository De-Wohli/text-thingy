package narrator

import "testing"

func TestAllGeneratorsProduceNonEmptyText(t *testing.T) {
	checks := []func() string{
		func() string { return DungeonEntry("Fuyune") },
		func() string { return AttackSwing("Fuyune", "Bandit", true, false, 5) },
		func() string { return AttackSwing("Fuyune", "Bandit", true, true, 12) },
		func() string { return AttackSwing("Fuyune", "Bandit", false, false, 0) },
		func() string { return RoomVictory("Fuyune", "corridor", []string{"Bandit"}) },
		func() string { return RoomDefeat("Fuyune", "corridor") },
		func() string { return DungeonResolved("Fuyune", 25) },
		func() string { return ChoiceResolution("Let them go", "merciful") },
		func() string { return ChoiceResolution("Silence them", "ruthless") },
		func() string { return ChoiceResolution("Take their gear", "pragmatic") },
		func() string { return VoteResolution("Let them go", "merciful", true) },
		func() string { return VoteResolution("Let them go", "merciful", false) },
	}
	for i, check := range checks {
		for attempt := 0; attempt < 10; attempt++ {
			if got := check(); got == "" {
				t.Fatalf("generator %d produced empty text", i)
			}
		}
	}
}

func TestLowerFirst(t *testing.T) {
	if got := lower("Let them go"); got != "let them go" {
		t.Errorf("lower(%q) = %q", "Let them go", got)
	}
	if got := lower(""); got != "" {
		t.Errorf("lower(\"\") = %q, want empty", got)
	}
}
