// Package honor implements the Honor & Alignment Matrix: an account-wide
// score from -100 to +100 that maps to an alignment band and drives city
// NPC reactivity, shop pricing, and which quest/choice types are offered.
package honor

import "dnd5e-web/backend/internal/models"

const (
	Min             = -100
	Max             = 100
	BetrayalPenalty = -20
)

type Alignment string

const (
	Good    Alignment = "good"
	Neutral Alignment = "neutral"
	Evil    Alignment = "evil"
)

type Reactivity struct {
	Alignment         Alignment `json:"alignment"`
	Greeting          string    `json:"greeting"`
	ShopPriceModifier float64   `json:"shopPriceModifier"`
	GuardBehavior     string    `json:"guardBehavior"`
	QuestTypes        []string  `json:"questTypes"`
}

var bands = []struct {
	min, max int
	profile  Reactivity
}{
	{
		min: 60, max: 100,
		profile: Reactivity{
			Alignment:         Good,
			Greeting:          "Cheerful greetings",
			ShopPriceModifier: -0.1,
			GuardBehavior:     "Guards assist in fights",
			QuestTypes:        []string{"Rescue civilians", "Defend caravans", "Holy relic recovery"},
		},
	},
	{
		min: -59, max: 59,
		profile: Reactivity{
			Alignment:         Neutral,
			Greeting:          "Standard, professional dialogue",
			ShopPriceModifier: 0,
			GuardBehavior:     "Default market prices",
			QuestTypes:        []string{"Monster cull", "Bounty hunting", "Material gathering"},
		},
	},
	{
		min: -100, max: -60,
		profile: Reactivity{
			Alignment:         Evil,
			Greeting:          "Suspicious/hostile dialogue",
			ShopPriceModifier: 0.2,
			GuardBehavior:     "Guards follow closely",
			QuestTypes:        []string{"Smuggling", "Infiltrate rival hideouts", "Assassination"},
		},
	},
}

func Clamp(value int) int {
	if value > Max {
		return Max
	}
	if value < Min {
		return Min
	}
	return value
}

func ReactivityForHonor(honorScore int) Reactivity {
	for _, b := range bands {
		if honorScore >= b.min && honorScore <= b.max {
			return b.profile
		}
	}
	// Unreachable given the three bands above cover [-100, 100] fully.
	return bands[1].profile
}

// ApplyChoice resolves the Honor & Alignment Impact Matrix: every Choice or
// Vote outcome carries a ChoiceTypology, which maps directly to a fixed
// Honor delta.
func ApplyChoice(currentHonor int, typology models.ChoiceTypology) int {
	return Clamp(currentHonor + models.HonorImpact[typology])
}

func ApplyBetrayal(currentHonor int) int {
	return Clamp(currentHonor + BetrayalPenalty)
}
