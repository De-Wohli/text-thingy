// Package chat is the gateway's in-process WebSocket connection registry.
// Cross-instance fan-out happens over Redis pub/sub (see internal/redisstate);
// this hub is what turns a decoded pub/sub message into writes on the
// specific local sockets that should receive it.
package chat

import (
	"sync"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	Conn      *websocket.Conn
	AccountID string
	PartyID   string
	writeMu   sync.Mutex
}

func (c *Client) WriteJSON(v any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.Conn.WriteJSON(v)
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client // accountID -> client
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c.AccountID] = c
}

func (h *Hub) Unregister(accountID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, accountID)
}

func (h *Hub) SetPartyID(accountID, partyID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if c, ok := h.clients[accountID]; ok {
		c.PartyID = partyID
	}
}

func (h *Hub) SendTo(accountID string, payload any) {
	h.mu.RLock()
	c, ok := h.clients[accountID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	_ = c.WriteJSON(payload)
}

// BroadcastAll writes payload to every locally-connected client, used for
// global/guild/rp channels where the prototype does not yet model
// fine-grained zone membership.
func (h *Hub) BroadcastAll(payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		_ = c.WriteJSON(payload)
	}
}

// BroadcastToParty writes payload only to locally-connected clients whose
// PartyID matches.
func (h *Hub) BroadcastToParty(partyID string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		if c.PartyID == partyID {
			_ = c.WriteJSON(payload)
		}
	}
}

// BroadcastToAccounts writes payload to exactly the given accounts — used
// for groups that aren't necessarily a party, e.g. everyone standing at a
// location (who may not have partied up yet) or everyone present in a
// dungeon run.
func (h *Hub) BroadcastToAccounts(accountIDs []string, payload any) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, id := range accountIDs {
		if c, ok := h.clients[id]; ok {
			_ = c.WriteJSON(payload)
		}
	}
}
