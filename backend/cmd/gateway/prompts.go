package main

import (
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/voting"
)

// npcPrompts is the server-authoritative catalog of narrative choices an
// NPC can offer. A richer version would load these per-quest from
// Postgres; the prototype ships one fixed prompt per citizen interaction.
var npcPrompts = map[string]voting.ChoicePrompt{
	"bandit-camp": {
		ID:     "bandit-camp",
		Prompt: "A captured bandit begs for mercy. What do you do?",
		Options: []models.ChoiceOption{
			{ID: "merciful", Label: "Let them go and escort them to the city watch", Typology: models.ChoiceMerciful},
			{ID: "pragmatic", Label: "Take their gear and leave them bound in the cellar", Typology: models.ChoicePragmatic},
			{ID: "ruthless", Label: "Silence them so they can't warn the rest of the camp", Typology: models.ChoiceRuthless},
		},
	},
}

func defaultPrompt() voting.ChoicePrompt {
	return npcPrompts["bandit-camp"]
}
