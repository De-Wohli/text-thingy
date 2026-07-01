// Package narrator generates Game-Master-voiced flavor text for game
// events (travel, entering a dungeon, combat swings, room outcomes, skill
// checks, NPC choices, party votes). It's template-based, not an LLM call
// — deliberately so, to keep the gateway's response latency and dependency
// surface small — but picks randomly among a handful of phrasings per
// event so the table doesn't see the exact same line twice in a row.
//
// The random-pick step is behind the small Backend interface below
// specifically so a future generative backend (e.g. a local Ollama model)
// can be swapped in via Active without touching every call site — it
// isn't built yet, this is just the seam for it.
package narrator

import (
	"fmt"
	"math/rand"
	"strings"
)

// Backend supplies the line for a given set of candidate phrasings. The
// template backend picks one at random; a future LLM-backed implementation
// could instead use the options as a style reference and generate fresh
// text, or simply rephrase one of them.
type Backend interface {
	Pick(options []string) string
}

type templateBackend struct{}

func (templateBackend) Pick(options []string) string {
	return options[rand.Intn(len(options))]
}

// Active is the narration backend in use. Swap it (e.g. at startup, based
// on an env var) to change how every narrator.* function below resolves
// its text — nothing else in the gateway needs to change.
var Active Backend = templateBackend{}

func pick(options []string) string {
	return Active.Pick(options)
}

// EnterLocation narrates arriving somewhere on the world map.
func EnterLocation(characterName, locationName, locationDescription string) string {
	return pick([]string{
		fmt.Sprintf("%s arrives at %s. %s", characterName, locationName, locationDescription),
		fmt.Sprintf("%s makes their way to %s. %s", characterName, locationName, locationDescription),
	})
}

// DungeonEntry narrates stepping through a quest hook into a dungeon.
func DungeonEntry(characterName string) string {
	return pick([]string{
		fmt.Sprintf("The ground gives way to old stonework as %s descends into the dark. Torchlight flickers ahead.", characterName),
		fmt.Sprintf("%s pushes through a curtain of roots into a forgotten passage. The air turns cold.", characterName),
		fmt.Sprintf("A draft carries the smell of damp earth as %s steps over the threshold. Somewhere below, something stirs.", characterName),
	})
}

// SceneDescription narrates arriving in a dungeon room, describing any
// monsters present. Prefer RoomEntry when a full room description is available.
func SceneDescription(roomLabel string, monsterNames []string) string {
	if len(monsterNames) == 0 {
		return pick([]string{
			fmt.Sprintf("The %s is empty, quiet but for your own footsteps.", roomLabel),
			fmt.Sprintf("Nothing stirs in the %s — for now.", roomLabel),
		})
	}
	foes := strings.Join(monsterNames, ", ")
	return pick([]string{
		fmt.Sprintf("The %s opens up ahead. %s block the way, weapons ready.", roomLabel, foes),
		fmt.Sprintf("You round the corner into the %s — %s are already watching you.", roomLabel, foes),
	})
}

// RoomEntry uses the room's own authored description as the primary narration,
// then appends a brief monster-presence line when enemies are present. This
// is richer than SceneDescription because themed rooms carry their own voice.
func RoomEntry(roomDescription string, monsterNames []string) string {
	if len(monsterNames) == 0 {
		return roomDescription
	}
	foes := strings.Join(monsterNames, ", ")
	suffix := pick([]string{
		fmt.Sprintf(" — and %s stand between you and the next door.", foes),
		fmt.Sprintf(" %s are waiting.", foes),
		fmt.Sprintf(" You spot %s. The fight is unavoidable.", foes),
	})
	return roomDescription + suffix
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

// Dodge narrates a combatant taking the Dodge action.
func Dodge(name string) string {
	return pick([]string{
		fmt.Sprintf("%s drops into a defensive stance, watching for the next strike.", name),
		fmt.Sprintf("%s keeps light on their feet, ready to weave away from any attack.", name),
	})
}

// Flee narrates a combatant retreating from a fight.
func Flee(name string) string {
	return pick([]string{
		fmt.Sprintf("%s breaks off and retreats from the fight.", name),
		fmt.Sprintf("%s falls back, putting distance between themself and the enemy.", name),
	})
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

// RoomDefeat narrates a lost encounter (no permadeath — the party retreats
// and is stabilized, see internal/combat for the rules).
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

// SkillCheckOutcome narrates a non-combat ability check (basic version,
// without knowing the specific mechanical outcome). Prefer SkillCheckOutcomeDetailed.
func SkillCheckOutcome(characterName, context string, success bool) string {
	return SkillCheckOutcomeDetailed(characterName, "", "", 0, success)
}

// SkillCheckOutcomeDetailed narrates a non-combat ability check incorporating
// the specific mechanical outcome so the GM description matches what happened.
func SkillCheckOutcomeDetailed(characterName, skill, outcome string, outcomeValue int, success bool) string {
	switch outcome {
	case "monster_removed":
		return pick([]string{
			fmt.Sprintf("%s searches carefully and spots a hidden threat — one enemy won't make it to the fight.", characterName),
			fmt.Sprintf("Good eyes. %s's investigation removes one of the threats before it can spring.", characterName),
		})
	case "player_first":
		return pick([]string{
			fmt.Sprintf("%s hears movement in time. The party will act before anything else can.", characterName),
			fmt.Sprintf("Quiet listening pays off — %s gives the signal, and the party moves first.", characterName),
		})
	case "sneak_attack":
		return pick([]string{
			fmt.Sprintf("%s slips forward undetected and lands a free strike before the fight even begins.", characterName),
			fmt.Sprintf("The shadows work in %s's favour — a free attack before initiative is rolled.", characterName),
		})
	case "attack_bonus":
		return fmt.Sprintf("%s reads the enemy's posture and stance — the party fights with +%d to attack rolls this encounter.", characterName, outcomeValue)
	case "damage_bonus":
		return fmt.Sprintf("%s identifies a vulnerability. The party deals +%d damage on every hit.", characterName, outcomeValue)
	case "temp_hp":
		return fmt.Sprintf("%s braces and centres — everyone nearby gains %d temporary HP going into the fight.", characterName, outcomeValue)
	case "trap_damage":
		return pick([]string{
			fmt.Sprintf("A wire snaps under %s's boot. A hidden trap fires — %d damage.", characterName, outcomeValue),
			fmt.Sprintf("The search goes wrong. %s triggers a trap for %d damage.", characterName, outcomeValue),
		})
	case "monster_ready":
		return pick([]string{
			fmt.Sprintf("Something gives %s away. The enemy is ready — they'll have an edge on the first round.", characterName),
			fmt.Sprintf("The monsters heard %s coming. They'll strike harder in the opening round.", characterName),
		})
	}
	// Generic fallback
	if success {
		return pick([]string{
			fmt.Sprintf("%s's instincts pay off.", characterName),
			fmt.Sprintf("A careful approach serves %s well.", characterName),
		})
	}
	return pick([]string{
		fmt.Sprintf("%s finds nothing useful this time.", characterName),
		fmt.Sprintf("Nothing comes of the attempt — %s comes up empty.", characterName),
	})
}

// PartyFormed narrates two adventurers teaming up.
func PartyFormed(leaderName, joinerName string) string {
	return pick([]string{
		fmt.Sprintf("%s and %s strike up a pact — partners for whatever comes next.", leaderName, joinerName),
		fmt.Sprintf("%s clasps arms with %s. The party is formed.", leaderName, joinerName),
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
