package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-starter/internal/core/config"
	"go-starter/internal/core/domain"
	auth_handlers "go-starter/internal/features/auth/transport/http"

	"github.com/gin-gonic/gin"
)

type mockAuthService struct {
	registerFn func(ctx context.Context, mail, password string) (*domain.User, string, error)
	loginFn    func(ctx context.Context, mail, password string) (*domain.User, string, error)
}

func (m *mockAuthService) Register(ctx context.Context, mail, password string) (*domain.User, string, error) {
	return m.registerFn(ctx, mail, password)
}

func (m *mockAuthService) Login(ctx context.Context, mail, password string) (*domain.User, string, error) {
	return m.loginFn(ctx, mail, password)
}

func setupAuthRouter(svc auth_handlers.AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	cfg := &config.Config{ProductionStatus: false}
	h := auth_handlers.NewAuthHandler(svc, cfg)
	for _, route := range h.Routes() {
		r.Handle(route.Method, "/auth"+route.Path, route.Handler)
	}
	return r
}

func toJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

func parseResponse(t *testing.T, body *bytes.Buffer) map[string]any {
	t.Helper()
	var resp map[string]any
	if err := json.Unmarshal(body.Bytes(), &resp); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return resp
}

func fakeUser(mail string) *domain.User {
	now := time.Now()
	return &domain.User{ID: "1", Mail: mail, CreatedAt: &now}
}

func TestRegisterHandler_InvalidJSON(t *testing.T) {
	r := setupAuthRouter(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegisterHandler_InvalidEmail(t *testing.T) {
	r := setupAuthRouter(&mockAuthService{})

	body := toJSON(t, map[string]string{"mail": "not-an-email", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegisterHandler_ServiceError(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
			return nil, "", errors.New("mail already registered")
		},
	}
	r := setupAuthRouter(svc)

	body := toJSON(t, map[string]string{"mail": "a@b.com", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 on service error")
	}
	resp := parseResponse(t, w.Body)
	if resp["success"] == true {
		t.Error("expected success=false")
	}
}

func TestRegisterHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		registerFn: func(_ context.Context, mail, _ string) (*domain.User, string, error) {
			return fakeUser(mail), "fake-jwt-token", nil
		},
	}
	r := setupAuthRouter(svc)

	body := toJSON(t, map[string]string{"mail": "new@test.com", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w.Body)
	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	r := setupAuthRouter(&mockAuthService{})

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{bad}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginHandler_ServiceError(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, _, _ string) (*domain.User, string, error) {
			return nil, "", errors.New("invalid credentials")
		},
	}
	r := setupAuthRouter(svc)

	body := toJSON(t, map[string]string{"mail": "a@b.com", "password": "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 on login failure")
	}
}

func TestLoginHandler_Success(t *testing.T) {
	svc := &mockAuthService{
		loginFn: func(_ context.Context, mail, _ string) (*domain.User, string, error) {
			return fakeUser(mail), "fake-jwt-token", nil
		},
	}
	r := setupAuthRouter(svc)

	body := toJSON(t, map[string]string{"mail": "a@b.com", "password": "secret123"})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w.Body)
	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}
}
