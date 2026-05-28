package booking

import (
	"context"
	"testing"
	"time"

	"github.com/MaXonchik07/gym-backend/internal/models"
	"github.com/pashagolub/pgxmock/v2"
)

func TestSaveMessage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	msg := &models.Message{
		SenderID:    "user-1",
		RecipientID: "support",
		Content:     "Hello",
	}

	mock.ExpectQuery(`INSERT INTO messages`).
		WithArgs("user-1", "support", "Hello").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).
			AddRow("msg-1", time.Now()))

	err = repo.SaveMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg-1" {
		t.Errorf("expected msg-1, got %s", msg.ID)
	}
}

func TestGetRecentMessagesForUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)
	userID := "user-1"
	limit := 10

	mock.ExpectQuery(`SELECT (.+) FROM messages WHERE sender_id = \$1 OR recipient_id = \$1 ORDER BY created_at ASC LIMIT \$2`).
		WithArgs(userID, limit).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sender_id", "recipient_id", "content", "created_at"}).
			AddRow("m1", "user-1", "support", "Hi", time.Now()).
			AddRow("m2", "support", "user-1", "Hello", time.Now()))

	msgs, err := repo.GetRecentMessagesForUser(context.Background(), userID, limit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}
