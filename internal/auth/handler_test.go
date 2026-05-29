package auth

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MaXonchik07/gym-backend/internal/models"
	"github.com/MaXonchik07/gym-backend/pkg/jwt"
	"github.com/MaXonchik07/gym-backend/pkg/middleware"
	"github.com/rs/zerolog"
)

type mockService struct {
	registerFn func(ctx context.Context, req *RegisterRequest) (*models.User, error)
	loginFn    func(ctx context.Context, req *LoginRequest) (string, error)
	updateFn   func(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error)
	getUserFn  func(ctx context.Context, id string) (*models.User, error)
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
func (m *mockService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	if m.getUserFn != nil {
		return m.getUserFn(ctx, id)
	}
	return nil, nil
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

func TestHandler_UpdateProfile_Success(t *testing.T) {
    svc := &mockService{
        updateFn: func(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error) {
            return &models.User{ID: userID, FirstName: req.FirstName}, nil
        },
    }
    handler := NewHandler(svc, zerolog.Nop())
    body := `{"first_name":"NewName","last_name":"NewLast","phone":"+79991112233","email":"new@example.com"}`
    req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewBufferString(body))
    req.Header.Set("Content-Type", "application/json")
    ctx := context.WithValue(req.Context(), middleware.UserContextKey, &jwt.Claims{UserID: "user123"})
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()

    handler.UpdateProfile(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("expected 200, got %d", rec.Code)
    }
}

func TestHandler_UpdateProfile_NoUserInContext(t *testing.T) {
    svc := &mockService{}
    handler := NewHandler(svc, zerolog.Nop())
    req := httptest.NewRequest(http.MethodPut, "/profile", nil)
    rec := httptest.NewRecorder()
    handler.UpdateProfile(rec, req)
    if rec.Code != http.StatusUnauthorized {
        t.Errorf("expected 401, got %d", rec.Code)
    }
}

func TestHandler_UpdateProfile_InvalidJSON(t *testing.T) {
    svc := &mockService{}
    handler := NewHandler(svc, zerolog.Nop())
    req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewBufferString(`{bad json`))
    req.Header.Set("Content-Type", "application/json")
    ctx := context.WithValue(req.Context(), middleware.UserContextKey, &jwt.Claims{UserID: "u1"})
    req = req.WithContext(ctx)
    rec := httptest.NewRecorder()
    handler.UpdateProfile(rec, req)
    if rec.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rec.Code)
    }
}

func TestHandler_Register_InvalidJSON(t *testing.T) {
    svc := &mockService{}
    handler := NewHandler(svc, zerolog.Nop())
    req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{not json`))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    handler.Register(rec, req)
    if rec.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rec.Code)
    }
}

func TestHandler_Login_InvalidJSON(t *testing.T) {
    svc := &mockService{}
    handler := NewHandler(svc, zerolog.Nop())
    req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{bad`))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    handler.Login(rec, req)
    if rec.Code != http.StatusBadRequest {
        t.Errorf("expected 400, got %d", rec.Code)
    }
}




