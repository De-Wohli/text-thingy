package honor

import (
	"testing"

	"dnd5e-web/backend/internal/models"
)

func TestReactivityForHonor(t *testing.T) {
	cases := []struct {
		honor int
		want  Alignment
	}{
		{75, Good}, {60, Good}, {100, Good},
		{0, Neutral}, {59, Neutral}, {-59, Neutral},
		{-60, Evil}, {-100, Evil},
	}
	for _, c := range cases {
		got := ReactivityForHonor(c.honor).Alignment
		if got != c.want {
			t.Errorf("ReactivityForHonor(%d).Alignment = %s, want %s", c.honor, got, c.want)
		}
	}
}

func TestClamp(t *testing.T) {
	if Clamp(150) != 100 {
		t.Errorf("expected clamp to ceiling")
	}
	if Clamp(-150) != -100 {
		t.Errorf("expected clamp to floor")
	}
	if Clamp(42) != 42 {
		t.Errorf("expected unchanged value within range")
	}
}

func TestApplyChoice(t *testing.T) {
	if got := ApplyChoice(90, models.ChoiceMerciful); got != 100 {
		t.Errorf("ApplyChoice merciful clamp = %d, want 100", got)
	}
	if got := ApplyChoice(-90, models.ChoiceRuthless); got != -100 {
		t.Errorf("ApplyChoice ruthless clamp = %d, want -100", got)
	}
	if got := ApplyChoice(0, models.ChoicePragmatic); got != 0 {
		t.Errorf("ApplyChoice pragmatic = %d, want 0", got)
	}
}

func TestApplyBetrayal(t *testing.T) {
	if got := ApplyBetrayal(50); got != 30 {
		t.Errorf("ApplyBetrayal(50) = %d, want 30", got)
	}
	if got := ApplyBetrayal(-90); got != -100 {
		t.Errorf("ApplyBetrayal(-90) = %d, want -100 (clamped)", got)
	}
}
