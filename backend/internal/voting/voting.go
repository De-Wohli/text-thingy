// Package voting implements the Dynamic Choice & Voting Engine. In Solo
// Mode the player's choice resolves immediately; in Party Mode a 30-second
// voting window collects votes from every party member before resolving,
// breaking ties by Guild Renown (falling back to a coin flip).
package voting

import (
	"errors"
	"math/rand"
	"time"

	"dnd5e-web/backend/internal/models"
)

const PartyVoteWindow = 30 * time.Second

var ErrVotingClosed = errors.New("voting window has closed")
var ErrUnknownOption = errors.New("option does not exist in this prompt")

type ChoicePrompt struct {
	ID      string
	Prompt  string
	Mode    models.ChoiceMode
	Options []models.ChoiceOption
}

type VoteRoom struct {
	Prompt   ChoicePrompt
	Deadline time.Time
	Votes    map[string]string // accountID -> optionID
}

func NewVoteRoom(prompt ChoicePrompt) *VoteRoom {
	return &VoteRoom{
		Prompt:   prompt,
		Deadline: time.Now().Add(PartyVoteWindow),
		Votes:    make(map[string]string),
	}
}

func (r *VoteRoom) optionExists(optionID string) bool {
	for _, o := range r.Prompt.Options {
		if o.ID == optionID {
			return true
		}
	}
	return false
}

func (r *VoteRoom) CastVote(accountID, optionID string) error {
	if time.Now().After(r.Deadline) {
		return ErrVotingClosed
	}
	if !r.optionExists(optionID) {
		return ErrUnknownOption
	}
	r.Votes[accountID] = optionID
	return nil
}

func (r *VoteRoom) Tally() map[string]int {
	tally := make(map[string]int, len(r.Prompt.Options))
	for _, o := range r.Prompt.Options {
		tally[o.ID] = 0
	}
	for _, optionID := range r.Votes {
		tally[optionID]++
	}
	return tally
}

type Resolution struct {
	OptionID     string
	TieBreakUsed bool
	CoinFlipUsed bool
}

// Resolve picks the winning option once the voting window has closed.
// renownByAccount supplies each voter's Guild Renown (proxied by Honor
// score) for tie-breaking; ties among renown fall back to a coin flip.
func Resolve(r *VoteRoom, renownByAccount map[string]int) Resolution {
	tally := r.Tally()

	maxVotes := -1
	for _, count := range tally {
		if count > maxVotes {
			maxVotes = count
		}
	}

	tied := make([]string, 0)
	for _, o := range r.Prompt.Options {
		if tally[o.ID] == maxVotes {
			tied = append(tied, o.ID)
		}
	}

	if len(tied) == 1 {
		return Resolution{OptionID: tied[0]}
	}

	// Tie-break by the highest Guild Renown among voters who backed one of
	// the tied options.
	bestOption := ""
	bestRenown := -1 << 31
	renownTied := false
	for accountID, optionID := range r.Votes {
		if !contains(tied, optionID) {
			continue
		}
		renown := renownByAccount[accountID]
		if renown > bestRenown {
			bestRenown = renown
			bestOption = optionID
			renownTied = false
		} else if renown == bestRenown && optionID != bestOption {
			renownTied = true
		}
	}

	if bestOption != "" && !renownTied {
		return Resolution{OptionID: bestOption, TieBreakUsed: true}
	}

	// Still tied (or nobody voted) — coin flip among the tied options.
	winner := tied[rand.Intn(len(tied))]
	return Resolution{OptionID: winner, TieBreakUsed: true, CoinFlipUsed: true}
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
