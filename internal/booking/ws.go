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
	clients map[*websocket.Conn]string
	logger  zerolog.Logger
	msgRepo MessageRepository
}

func NewHub(logger zerolog.Logger, msgRepo MessageRepository) *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]string),
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

	userID := r.URL.Query().Get("user_id")
	h.mu.Lock()
	h.clients[conn] = userID
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	if userID != "" {
		msgs, _ := h.msgRepo.GetRecentMessagesForUser(context.Background(), userID, 50)
		history, _ := json.Marshal(msgs)
		conn.WriteMessage(websocket.TextMessage, history)
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("[]"))
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg models.Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		msg.ID = uuid.New().String()
		msg.CreatedAt = time.Now()
		msg.SenderID = userID

		if msg.RecipientID == "" && userID != "" {
			msg.RecipientID = "support"
		}

		if err := h.msgRepo.SaveMessage(context.Background(), &msg); err != nil {
			h.logger.Error().Err(err).Msg("failed to save message")
			continue
		}

		broadcast, _ := json.Marshal(msg)
		h.mu.RLock()
		for c, uid := range h.clients {
			if uid == msg.SenderID || uid == msg.RecipientID || uid == "support" {
				c.WriteMessage(websocket.TextMessage, broadcast)
			}
		}
		h.mu.RUnlock()
	}
}
