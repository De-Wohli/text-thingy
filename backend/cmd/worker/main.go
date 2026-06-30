// Command worker consumes the two asynchronous queues the gateway offloads
// heavy work to: dungeon_generation_queue (procedural grid generation) and
// vote_resolution_queue (batched Honor writes once a vote window closes).
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"dnd5e-web/backend/internal/config"
	"dnd5e-web/backend/internal/dungeon"
	"dnd5e-web/backend/internal/models"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/store"
	"dnd5e-web/backend/internal/wsproto"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	st, err := store.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	rs := redisstate.New(cfg.RedisAddr)
	defer rs.Close()

	q, err := queue.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer q.Close()

	dungeonDeliveries, err := q.Consume(queue.QueueDungeonGeneration, "worker-dungeon")
	if err != nil {
		log.Fatalf("consume %s: %v", queue.QueueDungeonGeneration, err)
	}
	voteDeliveries, err := q.Consume(queue.QueueVoteResolution, "worker-vote")
	if err != nil {
		log.Fatalf("consume %s: %v", queue.QueueVoteResolution, err)
	}

	go handleDungeonJobs(ctx, st, rs, dungeonDeliveries)
	go handleVoteJobs(ctx, st, rs, voteDeliveries)

	log.Println("worker ready, consuming dungeon_generation_queue and vote_resolution_queue")
	select {} // block forever; both consumer goroutines run for the process lifetime
}

func handleDungeonJobs(ctx context.Context, st *store.Store, rs *redisstate.Client, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var job queue.DungeonGenerationJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			log.Printf("dungeon job: decode: %v", err)
			d.Nack(false, false)
			continue
		}

		partyKey := job.PartyID
		if partyKey == "" {
			partyKey = job.AccountID
		}

		generated := dungeon.Generate(uuid.NewString(), partyKey, job.CharacterLevel)
		if err := st.SaveDungeon(ctx, generated); err != nil {
			log.Printf("dungeon job %s: save: %v", job.JobID, err)
			d.Nack(false, true)
			continue
		}

		event := wsproto.DungeonReadyEvent{
			JobID:     job.JobID,
			AccountID: job.AccountID,
			PartyID:   job.PartyID,
			Dungeon:   generated,
		}
		if err := rs.Publish(ctx, redisstate.ChannelDungeonReady(job.JobID), event); err != nil {
			log.Printf("dungeon job %s: publish: %v", job.JobID, err)
		}
		d.Ack(false)
	}
}

func handleVoteJobs(ctx context.Context, st *store.Store, rs *redisstate.Client, deliveries <-chan amqp.Delivery) {
	for d := range deliveries {
		var job queue.VoteResolutionJob
		if err := json.Unmarshal(d.Body, &job); err != nil {
			log.Printf("vote job: decode: %v", err)
			d.Nack(false, false)
			continue
		}

		results := make([]wsproto.VoteResolvedResult, 0, len(job.AccountIDs))
		delta := models.HonorImpact[job.Typology]
		for _, accountID := range job.AccountIDs {
			newHonor, err := st.ApplyHonorDelta(ctx, accountID, delta, job.Typology, "party vote: "+job.PromptID)
			if err != nil {
				log.Printf("vote job %s: apply honor for %s: %v", job.PromptID, accountID, err)
				continue
			}
			results = append(results, wsproto.VoteResolvedResult{
				AccountID:  accountID,
				HonorDelta: delta,
				NewHonor:   newHonor,
			})
		}

		event := wsproto.VoteResolvedEvent{
			PromptID: job.PromptID,
			OptionID: job.WinningOptionID,
			TieBreak: job.TieBreak,
			Results:  results,
		}
		if err := rs.Publish(ctx, redisstate.ChannelVoteResolved(job.PromptID), event); err != nil {
			log.Printf("vote job %s: publish: %v", job.PromptID, err)
		}
		d.Ack(false)
	}
}
