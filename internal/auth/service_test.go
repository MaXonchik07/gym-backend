package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

type mockRepo struct {
	users     map[string]*models.User
	createErr error
	getUserErr error
	updateErr  error
}

func (m *mockRepo) CreateUser(ctx context.Context, user *models.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.users == nil {
		m.users = make(map[string]*models.User)
	}
	if _, exists := m.users[user.Email]; exists {
		return errors.New("пользователь с таким email уже существует")
	}
	user.ID = "mock-user-id"
	user.JoinDate = time.Now()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	m.users[user.Email] = user
	return nil
}

func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}
	user, ok := m.users[email]
	if !ok {
		return nil, nil
	}
	return user, nil
}

func (m *mockRepo) UpdateUser(ctx context.Context, user *models.User) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if m.users == nil {
		return errors.New("user not found")
	}
	if _, ok := m.users[user.Email]; !ok {
		return errors.New("user not found")
	}
	m.users[user.Email] = user
	return nil
}

func TestRegister_Success(t *testing.T) {
	repo := &mockRepo{users: make(map[string]*models.User)}
	svc := NewService(repo, "test-secret")
	req := &RegisterRequest{
		FirstName: "Иван",
		LastName:  "Иванов",
		Email:     "ivan@example.com",
		Phone:     "+79001234567",
		Password:  "secret123",
	}
	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "mock-user-id" {
		t.Errorf("expected mock-user-id, got %s", user.ID)
	}
	if user.MembershipType != "basic" {
		t.Errorf("expected basic membership, got %s", user.MembershipType)
	}
	if user.Role != "user" {
		t.Errorf("expected user role, got %s", user.Role)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("secret123")); err != nil {
		t.Error("password hash mismatch")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockRepo{
		users: map[string]*models.User{
			"ivan@example.com": {Email: "ivan@example.com"},
		},
	}
	svc := NewService(repo, "test-secret")
	req := &RegisterRequest{
		Email:    "ivan@example.com",
		Password: "secret123",
	}
	_, err := svc.Register(context.Background(), req)
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "пользователь с таким email уже существует" {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	repo := &mockRepo{
		users: map[string]*models.User{
			"ivan@example.com": {
				Email:        "ivan@example.com",
				PasswordHash: string(hashed),
				Role:         "user",
			},
		},
	}
	svc := NewService(repo, "test-secret")
	token, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "ivan@example.com",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Error("expected token, got empty string")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &mockRepo{
		users: map[string]*models.User{
			"ivan@example.com": {
				Email:        "ivan@example.com",
				PasswordHash: string(hashed),
			},
		},
	}
	svc := NewService(repo, "test-secret")
	_, err := svc.Login(context.Background(), &LoginRequest{
		Email:    "ivan@example.com",
		Password: "wrong",
	})
	if err == nil {
		t.Error("expected error, got nil")
	}	
}
func TestRegister_RepositoryError(t *testing.T) {
    repo := &mockRepo{
        users:     make(map[string]*models.User),
        createErr: errors.New("database connection refused"),
    }
    svc := NewService(repo, "test-secret")
    req := &RegisterRequest{
        Email:    "test@example.com",
        Password: "secret123",
    }
    _, err := svc.Register(context.Background(), req)
    if err == nil {
        t.Error("expected error, got nil")
    }
    if err.Error() != "database connection refused" {
        t.Errorf("expected database error, got %v", err)
    }
}

func TestUpdateProfile_Success(t *testing.T) {
    hashed, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
    repo := &mockRepo{
        users: map[string]*models.User{
            "ivan@example.com": {
                ID:           "mock-user-id",
                Email:        "ivan@example.com",
                FirstName:    "Иван",
                LastName:     "Иванов",
                Phone:        "+79001234567",
                PasswordHash: string(hashed),
                Role:         "user",
                MembershipType: "basic",
            },
        },
    }
    svc := NewService(repo, "test-secret")
    req := &UpdateProfileRequest{
        FirstName:      "Пётр",
        LastName:       "Петров",
        Email:          "ivan@example.com",
        Phone:          "+79007654321",
        MembershipType: "premium",
    }
    user, err := svc.UpdateProfile(context.Background(), "mock-user-id", req)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if user.FirstName != "Пётр" {
        t.Errorf("expected Пётр, got %s", user.FirstName)
    }
    if user.MembershipType != "premium" {
        t.Errorf("expected premium, got %s", user.MembershipType)
    }
}

func TestLogin_UserNotFound(t *testing.T) {
    repo := &mockRepo{users: make(map[string]*models.User)}
    svc := NewService(repo, "test-secret")
    _, err := svc.Login(context.Background(), &LoginRequest{
        Email:    "notfound@example.com",
        Password: "secret123",
    })
    if err == nil {
        t.Error("expected error, got nil")
    }
    if err.Error() != "пользователь не найден" {
        t.Errorf("expected not found error, got %v", err)
    }
}
func TestUpdateProfile_UserNotFound(t *testing.T) {
    repo := &mockRepo{
        users: make(map[string]*models.User),
    }
    svc := NewService(repo, "test-secret")
    req := &UpdateProfileRequest{
        Email: "notfound@example.com",
    }
    _, err := svc.UpdateProfile(context.Background(), "non-existent", req)
    if err == nil {
        t.Error("expected error, got nil")
    }
}

func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
    if m.users == nil {
        return nil, nil
    }
    for _, u := range m.users {
        if u.ID == id {
            return u, nil
        }
    }
    return nil, nil
}