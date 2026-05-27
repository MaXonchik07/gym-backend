package booking

import (
	"context"

	"github.com/MaXonchik07/gym-backend/internal/models"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *models.Message) error
	GetRecentMessages(ctx context.Context, limit int) ([]models.Message, error)
}

type messageRepo struct {
	pool DBPool
}

func NewMessageRepository(pool DBPool) MessageRepository {
	return &messageRepo{pool: pool}
}

func (r *messageRepo) SaveMessage(ctx context.Context, msg *models.Message) error {
	query := `INSERT INTO messages (sender_id, content) VALUES ($1, $2) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, msg.SenderID, msg.Content).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *messageRepo) GetRecentMessages(ctx context.Context, limit int) ([]models.Message, error) {
	query := `SELECT id, sender_id, content, created_at FROM messages ORDER BY created_at DESC LIMIT $1`
	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}