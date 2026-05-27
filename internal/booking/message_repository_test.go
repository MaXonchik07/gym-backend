package booking

import (
	"context"
	"testing"
	"time"

	"github.com/pashagolub/pgxmock/v2"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

func TestSaveMessage(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)

	mock.ExpectQuery(`INSERT INTO messages`).
		WithArgs("user-1", "Hello").
		WillReturnRows(pgxmock.NewRows([]string{"id", "created_at"}).
			AddRow("msg-1", time.Now()))

	msg := &models.Message{
		SenderID: "user-1",
		Content:  "Hello",
	}
	err = repo.SaveMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID != "msg-1" {
		t.Errorf("expected msg-1, got %s", msg.ID)
	}
}

func TestGetRecentMessages(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	repo := NewMessageRepository(mock)

	mock.ExpectQuery(`SELECT (.+) FROM messages`).
		WithArgs(10).
		WillReturnRows(pgxmock.NewRows([]string{"id", "sender_id", "content", "created_at"}).
			AddRow("m1", "user-1", "Hi", time.Now()).
			AddRow("m2", "user-2", "Hello", time.Now()))

	msgs, err := repo.GetRecentMessages(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}