package combat

import "testing"

func TestRollD20Range(t *testing.T) {
	for i := 0; i < 200; i++ {
		roll := RollD20()
		if roll < 1 || roll > 20 {
			t.Fatalf("RollD20() = %d, want 1-20", roll)
		}
	}
}

func TestRollDiceRange(t *testing.T) {
	cases := []struct {
		expr     string
		min, max int
	}{
		{"1d6", 1, 6},
		{"1d6+1", 2, 7},
		{"2d4+2", 4, 10},
		{"1d8-1", 0, 7},
	}
	for _, c := range cases {
		for i := 0; i < 100; i++ {
			got := RollDice(c.expr)
			if got < c.min || got > c.max {
				t.Fatalf("RollDice(%q) = %d, want %d-%d", c.expr, got, c.min, c.max)
			}
		}
	}
}
