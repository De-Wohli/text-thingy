package narrator

import "testing"

func TestAllGeneratorsProduceNonEmptyText(t *testing.T) {
	checks := []func() string{
		func() string {
			return EnterLocation("Fuyune", "The Yawning Flask", "Lantern light and the smell of stew.")
		},
		func() string { return DungeonEntry("Fuyune") },
		func() string { return SceneDescription("corridor", []string{"Goblin", "Goblin"}) },
		func() string { return SceneDescription("corridor", nil) },
		func() string { return AttackSwing("Fuyune", "Bandit", true, false, 5) },
		func() string { return AttackSwing("Fuyune", "Bandit", true, true, 12) },
		func() string { return AttackSwing("Fuyune", "Bandit", false, false, 0) },
		func() string { return Dodge("Fuyune") },
		func() string { return Flee("Fuyune") },
		func() string { return RoomVictory("Fuyune", "corridor", []string{"Bandit"}) },
		func() string { return RoomDefeat("Fuyune", "corridor") },
		func() string { return DungeonResolved("Fuyune", 25) },
		func() string { return SkillCheckOutcome("Fuyune", "search for traps", true) },
		func() string { return SkillCheckOutcome("Fuyune", "search for traps", false) },
		func() string { return PartyFormed("Fuyune", "Brak") },
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

// stubBackend lets the test prove narrator.* functions actually go through
// Active rather than calling math/rand directly — this is the seam a
// future LLM backend would plug into.
type stubBackend struct{ calls int }

func (s *stubBackend) Pick(options []string) string {
	s.calls++
	return options[0]
}

func TestActiveBackendIsUsedForEveryPick(t *testing.T) {
	original := Active
	defer func() { Active = original }()

	stub := &stubBackend{}
	Active = stub

	_ = DungeonEntry("Fuyune")
	_ = AttackSwing("Fuyune", "Bandit", true, false, 5)

	if stub.calls != 2 {
		t.Fatalf("expected the swapped-in backend to be called twice, got %d calls", stub.calls)
	}
}
