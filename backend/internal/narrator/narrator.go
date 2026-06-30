// Package narrator generates Game-Master-voiced flavor text for game
// events (entering a dungeon, combat swings, room outcomes, NPC choices,
// party votes). It's template-based, not an LLM call — deliberately so,
// to keep the gateway's response latency and dependency surface small —
// but picks randomly among a handful of phrasings per event so the table
// doesn't see the exact same line twice in a row.
package narrator

import (
	"fmt"
	"math/rand"
)

func pick(options []string) string {
	return options[rand.Intn(len(options))]
}

// DungeonEntry narrates stepping through a [?] point of interest.
func DungeonEntry(characterName string) string {
	return pick([]string{
		fmt.Sprintf("The ground gives way to old stonework as %s descends into the dark. Torchlight flickers ahead.", characterName),
		fmt.Sprintf("%s pushes through a curtain of roots into a forgotten passage. The air turns cold.", characterName),
		fmt.Sprintf("A draft carries the smell of damp earth as %s steps over the threshold. Somewhere below, something stirs.", characterName),
	})
}

// AttackSwing narrates a single attack roll. damage is only meaningful
// when hit is true.
func AttackSwing(attacker, target string, hit, critical bool, damage int) string {
	switch {
	case critical && hit:
		return pick([]string{
			fmt.Sprintf("A perfect strike! %s's attack finds a gap in %s's guard for a brutal %d damage.", attacker, target, damage),
			fmt.Sprintf("Critical hit — %s drives the blow home, dealing %d damage to %s.", attacker, damage, target),
		})
	case hit:
		return pick([]string{
			fmt.Sprintf("%s's attack connects, dealing %d damage to %s.", attacker, damage, target),
			fmt.Sprintf("%s strikes true — %d damage to %s.", attacker, damage, target),
			fmt.Sprintf("%s lands a solid hit on %s for %d damage.", attacker, target, damage),
		})
	default:
		return pick([]string{
			fmt.Sprintf("%s's attack glances off %s, missing entirely.", attacker, target),
			fmt.Sprintf("%s swings at %s and misses.", attacker, target),
			fmt.Sprintf("%s narrowly fails to land a blow on %s.", attacker, target),
		})
	}
}

// RoomVictory narrates a cleared encounter.
func RoomVictory(characterName, roomLabel string, defeated []string) string {
	foes := "the last of the foes"
	if len(defeated) > 0 {
		foes = defeated[len(defeated)-1]
	}
	return pick([]string{
		fmt.Sprintf("%s strikes down %s. The %s falls silent.", characterName, foes, roomLabel),
		fmt.Sprintf("With the enemy defeated, %s catches their breath. The %s is clear.", characterName, roomLabel),
		fmt.Sprintf("The fight ends in %s's favor. The %s belongs to the party now.", characterName, roomLabel),
	})
}

// RoomDefeat narrates a lost encounter (no permadeath — the character
// retreats and is stabilized, see internal/combat for the rules).
func RoomDefeat(characterName, roomLabel string) string {
	return pick([]string{
		fmt.Sprintf("%s staggers, overwhelmed, and is forced to retreat from the %s. The wounds will heal, but the foe still stands.", characterName, roomLabel),
		fmt.Sprintf("Outmatched, %s falls back from the %s, barely keeping their footing.", characterName, roomLabel),
		fmt.Sprintf("The %s proves too much — %s withdraws to fight another day.", roomLabel, characterName),
	})
}

// DungeonResolved narrates the boss room falling and the reward.
func DungeonResolved(characterName string, goldAwarded int) string {
	return pick([]string{
		fmt.Sprintf("The depths fall quiet. %s emerges into daylight, %d gold heavier and a little wiser.", characterName, goldAwarded),
		fmt.Sprintf("With the boss defeated, %s loots the chamber — %d gold for the Guild vault — and makes for the surface.", characterName, goldAwarded),
		fmt.Sprintf("The dungeon is cleared. %s returns to the city gates carrying %d gold in spoils.", characterName, goldAwarded),
	})
}

// ChoiceResolution narrates the consequence of a solo or resolved party
// choice, keyed by the typology of the option chosen.
func ChoiceResolution(optionLabel, typology string) string {
	switch typology {
	case "merciful":
		return pick([]string{
			fmt.Sprintf("\"%s,\" you decide. Word of your mercy will travel — the city remembers kindness.", optionLabel),
			fmt.Sprintf("You choose to %s. It costs you nothing but earns trust that gold can't buy.", lower(optionLabel)),
		})
	case "ruthless":
		return pick([]string{
			fmt.Sprintf("\"%s.\" The choice is made, cold and final. Some will hear of this, and fear you for it.", optionLabel),
			fmt.Sprintf("You %s without hesitation. Whatever else this buys you, it isn't goodwill.", lower(optionLabel)),
		})
	default:
		return pick([]string{
			fmt.Sprintf("\"%s.\" A pragmatic call — no glory in it, but no regret either.", optionLabel),
			fmt.Sprintf("You %s and move on. The world keeps turning, indifferent.", lower(optionLabel)),
		})
	}
}

// VoteResolution narrates a party vote's outcome.
func VoteResolution(optionLabel string, typology string, tieBreak bool) string {
	base := ChoiceResolution(optionLabel, typology)
	if !tieBreak {
		return "The party agrees. " + base
	}
	return "The vote is split — the deciding voice carries it. " + base
}

func lower(s string) string {
	if s == "" {
		return s
	}
	b := []byte(s)
	if b[0] >= 'A' && b[0] <= 'Z' {
		b[0] += 'a' - 'A'
	}
	return string(b)
}
