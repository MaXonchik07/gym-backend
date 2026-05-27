package booking

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/MaXonchik07/gym-backend/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
	logger  zerolog.Logger
	msgRepo MessageRepository
}

func NewHub(logger zerolog.Logger, msgRepo MessageRepository) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
		logger:  logger,
		msgRepo: msgRepo,
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("websocket upgrade failed")
		return
	}
	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	msgs, _ := h.msgRepo.GetRecentMessages(context.Background(), 50)
	history, _ := json.Marshal(msgs)
	conn.WriteMessage(websocket.TextMessage, history)

	go func() {
		defer func() {
			conn.Close()
			h.mu.Lock()
			delete(h.clients, conn)
			h.mu.Unlock()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				h.logger.Error().Err(err).Msg("read error")
				break
			}
			h.logger.Info().Str("message", string(message)).Msg("received message")

			var msg models.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				h.logger.Error().Err(err).Msg("unmarshal error")
				continue
			}

			msg.ID = uuid.New().String()
			msg.CreatedAt = time.Now()

			if err := h.msgRepo.SaveMessage(context.Background(), &msg); err != nil {
				h.logger.Error().Err(err).Msg("failed to save message")
			}

			broadcast, _ := json.Marshal(msg)
			h.logger.Info().Str("broadcast", string(broadcast)).Msg("broadcasting")
			h.Broadcast(broadcast)
		}
	}()
}

func (h *Hub) Broadcast(message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			h.logger.Error().Err(err).Msg("broadcast error")
		}
	}
}