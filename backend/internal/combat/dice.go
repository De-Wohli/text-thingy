package combat

import (
	"math/rand"
	"strconv"
	"strings"
)

// RollD20 simulates a single d20 roll.
func RollD20() int {
	return rand.Intn(20) + 1
}

// RollDice evaluates a simple SRD damage expression like "1d6+1", "2d4",
// or "1d8-1" and returns the total.
func RollDice(expr string) int {
	expr = strings.TrimSpace(expr)
	sign := 1
	bonus := 0
	dicePart := expr

	if idx := strings.IndexAny(expr, "+-"); idx > 0 {
		dicePart = expr[:idx]
		bonusStr := expr[idx:]
		if bonusStr[0] == '-' {
			sign = -1
		}
		b, _ := strconv.Atoi(strings.TrimLeft(bonusStr, "+-"))
		bonus = sign * b
	}

	parts := strings.SplitN(dicePart, "d", 2)
	if len(parts) != 2 {
		return bonus
	}
	count, _ := strconv.Atoi(parts[0])
	sides, _ := strconv.Atoi(parts[1])
	if sides <= 0 {
		return bonus
	}

	total := bonus
	for i := 0; i < count; i++ {
		total += rand.Intn(sides) + 1
	}
	return total
}
