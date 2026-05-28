package auth

import (
	"context"
	"errors"
	"github.com/MaXonchik07/gym-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
	"testing"
)

type mockRepo struct {
	users      map[string]*models.User
	createErr  error
	getUserErr error
	updateErr  error
}

type mockMessageRepo struct {
    saveErr     error
    getRecentFn func(ctx context.Context, uid string, limit int) ([]models.Message, error)
    getUsersFn  func(ctx context.Context) ([]string, error)
}

func (m *mockMessageRepo) SaveMessage(ctx context.Context, msg *models.Message) error {
    if m.saveErr != nil {
        return m.saveErr
    }
    msg.ID = "mock-msg-id"
    return nil
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

func (m *mockRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
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
		FirstName: "Иван", LastName: "Иванов", Email: "ivan@example.com",
		Phone: "+79001234567", Password: "secret123",
	}
	user, err := svc.Register(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != "mock-user-id" {
		t.Errorf("expected mock-user-id, got %s", user.ID)
	}
	if user.MembershipType != "basic" {
		t.Errorf("expected basic, got %s", user.MembershipType)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := &mockRepo{users: map[string]*models.User{"ivan@example.com": {Email: "ivan@example.com"}}}
	svc := NewService(repo, "test-secret")
	_, err := svc.Register(context.Background(), &RegisterRequest{Email: "ivan@example.com", Password: "secret123"})
	if err == nil || err.Error() != "пользователь с таким email уже существует" {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestRegister_RepoError(t *testing.T) {
	repo := &mockRepo{createErr: errors.New("db down")}
	svc := NewService(repo, "test-secret")
	_, err := svc.Register(context.Background(), &RegisterRequest{Email: "x@x.com", Password: "secret123"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	repo := &mockRepo{users: map[string]*models.User{"ivan@example.com": {Email: "ivan@example.com", PasswordHash: string(hash), Role: "user"}}}
	svc := NewService(repo, "test-secret")
	token, err := svc.Login(context.Background(), &LoginRequest{Email: "ivan@example.com", Password: "secret123"})
	if err != nil || token == "" {
		t.Fatalf("expected token, got error %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &mockRepo{users: map[string]*models.User{"x@x.com": {Email: "x@x.com", PasswordHash: string(hash)}}}
	svc := NewService(repo, "test-secret")
	_, err := svc.Login(context.Background(), &LoginRequest{Email: "x@x.com", Password: "wrong"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockRepo{users: make(map[string]*models.User)}
	svc := NewService(repo, "test-secret")
	_, err := svc.Login(context.Background(), &LoginRequest{Email: "no@no.com", Password: "x"})
	if err == nil || err.Error() != "пользователь не найден" {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestLogin_GetUserByEmailError(t *testing.T) {
	repo := &mockRepo{getUserErr: errors.New("db error")}
	svc := NewService(repo, "test-secret")
	_, err := svc.Login(context.Background(), &LoginRequest{Email: "x@x.com", Password: "x"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpdateProfile_Success(t *testing.T) {
	repo := &mockRepo{users: map[string]*models.User{"a@a.com": {ID: "u1", Email: "a@a.com", FirstName: "A"}}}
	svc := NewService(repo, "test-secret")
	u, err := svc.UpdateProfile(context.Background(), "u1", &UpdateProfileRequest{
		FirstName: "B", LastName: "C", Phone: "123", Email: "a@a.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.FirstName != "B" {
		t.Errorf("expected B, got %s", u.FirstName)
	}
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	repo := &mockRepo{users: make(map[string]*models.User)}
	svc := NewService(repo, "test-secret")
	_, err := svc.UpdateProfile(context.Background(), "bad-id", &UpdateProfileRequest{Email: "x@x.com"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestUpdateProfile_EmailAlreadyTaken(t *testing.T) {
	repo := &mockRepo{users: map[string]*models.User{
		"a@a.com": {ID: "u1", Email: "a@a.com"},
		"b@b.com": {ID: "u2", Email: "b@b.com"},
	}}
	svc := NewService(repo, "test-secret")
	_, err := svc.UpdateProfile(context.Background(), "u1", &UpdateProfileRequest{Email: "b@b.com"})
	if err == nil || err.Error() != "email уже используется" {
		t.Errorf("expected conflict, got %v", err)
	}
}

func TestUpdateProfile_RepoError(t *testing.T) {
	repo := &mockRepo{updateErr: errors.New("db down")}
	svc := NewService(repo, "test-secret")
	_, err := svc.UpdateProfile(context.Background(), "u1", &UpdateProfileRequest{Email: "a@a.com"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestGetUserByID_Success(t *testing.T) {
	repo := &mockRepo{users: map[string]*models.User{"a@a.com": {ID: "u1"}}}
	svc := NewService(repo, "test-secret")
	u, err := svc.GetUserByID(context.Background(), "u1")
	if err != nil || u == nil {
		t.Fatalf("expected user, got error %v", err)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	repo := &mockRepo{users: make(map[string]*models.User)}
	svc := NewService(repo, "test-secret")
	u, _ := svc.GetUserByID(context.Background(), "bad")
	if u != nil {
		t.Error("expected nil")
	}
}

func TestLogin_BcryptCompareError(t *testing.T) {
    repo := &mockRepo{users: map[string]*models.User{
        "bad@hash.com": {Email: "bad@hash.com", PasswordHash: "not-a-bcrypt-hash"},
    }}
    svc := NewService(repo, "secret")
    _, err := svc.Login(context.Background(), &LoginRequest{Email: "bad@hash.com", Password: "anything"})
    if err == nil {
        t.Error("expected error from bcrypt compare")
    }
}





