package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"errors"

	"github.com/rs/zerolog"
	"github.com/MaXonchik07/gym-backend/internal/models"
)

type mockService struct {
	registerFn func(ctx context.Context, req *RegisterRequest) (*models.User, error)
	loginFn    func(ctx context.Context, req *LoginRequest) (string, error)
	updateFn   func(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error)
}

func (m *mockService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	return m.registerFn(ctx, req)
}
func (m *mockService) Login(ctx context.Context, req *LoginRequest) (string, error) {
	return m.loginFn(ctx, req)
}
func (m *mockService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error) {
	return m.updateFn(ctx, userID, req)
}

func TestHandler_Register_Success(t *testing.T) {
	svc := &mockService{
		registerFn: func(ctx context.Context, req *RegisterRequest) (*models.User, error) {
			return &models.User{ID: "123", Email: req.Email}, nil
		},
	}
	handler := NewHandler(svc, zerolog.Nop())

	body := `{"first_name":"Иван","last_name":"Иванов","email":"ivan@example.com","phone":"+79001234567","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestHandler_Login_Success(t *testing.T) {
	svc := &mockService{
		loginFn: func(ctx context.Context, req *LoginRequest) (string, error) {
			return "test-token", nil
		},
	}
	handler := NewHandler(svc, zerolog.Nop())

	body := `{"email":"ivan@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["token"] != "test-token" {
		t.Errorf("expected test-token, got %s", resp["token"])
	}
}
func TestHandler_Login_UserNotFound(t *testing.T) {
    svc := &mockService{
        loginFn: func(ctx context.Context, req *LoginRequest) (string, error) {
            return "", errors.New("пользователь не найден")
        },
    }
    handler := NewHandler(svc, zerolog.Nop())

    body := `{"email":"notfound@example.com","password":"secret123"}`
    req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.Login(rec, req)

    if rec.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rec.Code)
    }
}
func TestHandler_Register_Duplicate(t *testing.T) {
    svc := &mockService{
        registerFn: func(ctx context.Context, req *RegisterRequest) (*models.User, error) {
            return nil, errors.New("пользователь с таким email уже существует")
        },
    }
    handler := NewHandler(svc, zerolog.Nop())

    body := `{"first_name":"Иван","last_name":"Иванов","email":"ivan@example.com","phone":"+79001234567","password":"secret123"}`
    req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    handler.Register(rec, req)

    if rec.Code != http.StatusConflict {
        t.Errorf("expected 409, got %d", rec.Code)
    }
}