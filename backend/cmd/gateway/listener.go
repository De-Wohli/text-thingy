package main

import (
	"context"
	"encoding/json"
	"log"

	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/wsproto"
)

// startRedisListener bridges Redis pub/sub onto the local WebSocket hub.
// Chat channels fan out to every gateway instance subscribed to the same
// channel; this is what lets the architecture scale to multiple gateway
// replicas even though the prototype only runs one.
func (s *server) startRedisListener(ctx context.Context) {
	go s.listenFixedChannel(ctx, redisstate.ChannelChatGlobal, s.hub.BroadcastAll)
	go s.listenFixedChannel(ctx, redisstate.ChannelChatGuild, s.hub.BroadcastAll)
	go s.listenFixedChannel(ctx, redisstate.ChannelChatRP, s.hub.BroadcastAll)
	go s.listenPartyChat(ctx)
	go s.listenDungeonReady(ctx)
	go s.listenVoteResolved(ctx)
}

func (s *server) listenFixedChannel(ctx context.Context, channel string, deliver func(any)) {
	pubsub := s.redis.Subscribe(ctx, channel)
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		var broadcast wsproto.ChatBroadcast
		if err := json.Unmarshal([]byte(msg.Payload), &broadcast); err != nil {
			log.Printf("listener %s: decode: %v", channel, err)
			continue
		}
		deliver(broadcast)
	}
}

func (s *server) listenPartyChat(ctx context.Context) {
	pubsub := s.redis.SubscribePattern(ctx, "chat:party:*")
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		var broadcast wsproto.ChatBroadcast
		if err := json.Unmarshal([]byte(msg.Payload), &broadcast); err != nil {
			log.Printf("listener chat:party: decode: %v", err)
			continue
		}
		partyID, ok := suffixAfter(msg.Channel, "chat:party:")
		if !ok {
			continue
		}
		s.hub.BroadcastToParty(partyID, broadcast)
	}
}

func (s *server) listenDungeonReady(ctx context.Context) {
	pubsub := s.redis.SubscribePattern(ctx, "dungeon:ready:*")
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		var event wsproto.DungeonReadyEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("listener dungeon:ready: decode: %v", err)
			continue
		}

		key := event.AccountID
		if event.PartyID != "" {
			key = event.PartyID
		}
		s.dungeonsMu.Lock()
		d := event.Dungeon
		s.dungeons[key] = &d
		s.dungeonsMu.Unlock()

		ready := wsproto.NewDungeonReady(event.Dungeon)
		if event.PartyID != "" {
			s.hub.BroadcastToParty(event.PartyID, ready)
		} else {
			s.hub.SendTo(event.AccountID, ready)
		}
	}
}

func (s *server) listenVoteResolved(ctx context.Context) {
	pubsub := s.redis.SubscribePattern(ctx, "vote:resolved:*")
	defer pubsub.Close()
	for msg := range pubsub.Channel() {
		var event wsproto.VoteResolvedEvent
		if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
			log.Printf("listener vote:resolved: decode: %v", err)
			continue
		}
		for _, result := range event.Results {
			s.hub.SendTo(result.AccountID, wsproto.VoteResolved{
				Type:       "VOTE_RESOLVED",
				PromptID:   event.PromptID,
				OptionID:   event.OptionID,
				HonorDelta: result.HonorDelta,
				NewHonor:   result.NewHonor,
				TieBreak:   event.TieBreak,
			})
		}
	}
}

func suffixAfter(s, prefix string) (string, bool) {
	if len(s) <= len(prefix) || s[:len(prefix)] != prefix {
		return "", false
	}
	return s[len(prefix):], true
}
