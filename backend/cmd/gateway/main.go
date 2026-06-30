// Command gateway is the WebSocket/HTTP front door: it manages client
// connections, parses incoming JSON actions (MOVE, RP_CHAT, CAST_VOTE,
// SWAP_CHARACTER, ...), and offloads heavy async work (dungeon generation,
// vote resolution) to RabbitMQ instead of blocking the request.
package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"

	"dnd5e-web/backend/internal/config"
	"dnd5e-web/backend/internal/queue"
	"dnd5e-web/backend/internal/redisstate"
	"dnd5e-web/backend/internal/store"
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
	if err := rs.Ping(ctx); err != nil {
		log.Fatalf("connect redis: %v", err)
	}

	q, err := queue.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("connect rabbitmq: %v", err)
	}
	defer q.Close()

	srv := newServer(st, rs, q)
	srv.startRedisListener(ctx)

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/healthz", func(c *fiber.Ctx) error { return c.SendString("ok") })

	app.Post("/api/accounts", srv.handleCreateAccount)
	app.Get("/api/accounts/:id", srv.handleGetAccount)

	app.Use("/ws/:accountId", srv.wsUpgrade)
	app.Get("/ws/:accountId", websocket.New(srv.wsHandler))

	log.Printf("gateway listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
