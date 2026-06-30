package voting

import (
	"testing"
	"time"

	"dnd5e-web/backend/internal/models"
)

func samplePrompt() ChoicePrompt {
	return ChoicePrompt{
		ID:     "p1",
		Prompt: "Spare the bandit captain?",
		Mode:   models.ChoiceModeParty,
		Options: []models.ChoiceOption{
			{ID: "spare", Label: "Spare them", Typology: models.ChoiceMerciful},
			{ID: "execute", Label: "Execute them", Typology: models.ChoiceRuthless},
		},
	}
}

func TestCastVoteRejectsUnknownOption(t *testing.T) {
	room := NewVoteRoom(samplePrompt())
	if err := room.CastVote("acct-1", "nonexistent"); err != ErrUnknownOption {
		t.Fatalf("expected ErrUnknownOption, got %v", err)
	}
}

func TestCastVoteRejectsAfterDeadline(t *testing.T) {
	room := NewVoteRoom(samplePrompt())
	room.Deadline = time.Now().Add(-time.Second)
	if err := room.CastVote("acct-1", "spare"); err != ErrVotingClosed {
		t.Fatalf("expected ErrVotingClosed, got %v", err)
	}
}

func TestResolvePicksClearMajority(t *testing.T) {
	room := NewVoteRoom(samplePrompt())
	_ = room.CastVote("acct-1", "spare")
	_ = room.CastVote("acct-2", "spare")
	_ = room.CastVote("acct-3", "execute")

	res := Resolve(room, map[string]int{"acct-1": 0, "acct-2": 0, "acct-3": 0})
	if res.OptionID != "spare" {
		t.Fatalf("expected 'spare' to win, got %s", res.OptionID)
	}
	if res.TieBreakUsed {
		t.Fatal("did not expect a tie-break for a clear majority")
	}
}

func TestResolveBreaksTiesByRenown(t *testing.T) {
	room := NewVoteRoom(samplePrompt())
	_ = room.CastVote("acct-1", "spare")
	_ = room.CastVote("acct-2", "execute")

	res := Resolve(room, map[string]int{"acct-1": 10, "acct-2": 90})
	if res.OptionID != "execute" {
		t.Fatalf("expected the higher-renown voter's option ('execute') to win, got %s", res.OptionID)
	}
	if !res.TieBreakUsed {
		t.Fatal("expected tie-break to be flagged")
	}
}

func TestResolveFallsBackToCoinFlipWhenRenownTies(t *testing.T) {
	room := NewVoteRoom(samplePrompt())
	_ = room.CastVote("acct-1", "spare")
	_ = room.CastVote("acct-2", "execute")

	res := Resolve(room, map[string]int{"acct-1": 50, "acct-2": 50})
	if !res.CoinFlipUsed {
		t.Fatal("expected coin flip when renown also ties")
	}
	if res.OptionID != "spare" && res.OptionID != "execute" {
		t.Fatalf("unexpected winner %s", res.OptionID)
	}
}
