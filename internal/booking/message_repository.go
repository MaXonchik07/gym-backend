package booking

import (
	"context"

	"github.com/MaXonchik07/gym-backend/internal/models"
)

type MessageRepository interface {
	SaveMessage(ctx context.Context, msg *models.Message) error
	GetRecentMessagesForUser(ctx context.Context, userID string, limit int) ([]models.Message, error)
	GetChatUsers(ctx context.Context) ([]string, error)
}

type messageRepo struct {
	pool DBPool
}

func NewMessageRepository(pool DBPool) MessageRepository {
	return &messageRepo{pool: pool}
}

func (r *messageRepo) SaveMessage(ctx context.Context, msg *models.Message) error {
	query := `INSERT INTO messages (sender_id, recipient_id, content) VALUES ($1, $2, $3) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, msg.SenderID, msg.RecipientID, msg.Content).Scan(&msg.ID, &msg.CreatedAt)
}

func (r *messageRepo) GetRecentMessagesForUser(ctx context.Context, userID string, limit int) ([]models.Message, error) {
	query := `
		SELECT id, sender_id, COALESCE(recipient_id, ''), content, created_at
		FROM messages
		WHERE sender_id = $1 OR recipient_id = $1
		ORDER BY created_at ASC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.RecipientID, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (r *messageRepo) GetChatUsers(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT sender_id FROM messages WHERE recipient_id = 'support'`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		users = append(users, userID)
	}
	return users, rows.Err()
}
