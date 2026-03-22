package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-starter/internal/core/domain"
	user_handlers "go-starter/internal/features/user/transport/http"

	"github.com/gin-gonic/gin"
)

type mockUserService struct {
	getUserByIdFn func(ctx context.Context, userId int) (*domain.User, error)
}

func (m *mockUserService) GetUserById(ctx context.Context, userId int) (*domain.User, error) {
	return m.getUserByIdFn(ctx, userId)
}

func setupUserRouter(svc user_handlers.UserService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := user_handlers.NewUserHandler(svc)
	for _, route := range h.Routes() {
		r.Handle(route.Method, "/user"+route.Path, route.Handler)
	}
	return r
}

func TestGetUserById_NoUserID(t *testing.T) {
	svc := &mockUserService{
		getUserByIdFn: func(_ context.Context, id int) (*domain.User, error) {
			now := time.Now()
			return &domain.User{ID: "0", Mail: "a@b.com", CreatedAt: &now}, nil
		},
	}
	r := setupUserRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/user/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestGetUserById_ServiceError(t *testing.T) {
	svc := &mockUserService{
		getUserByIdFn: func(_ context.Context, _ int) (*domain.User, error) {
			return nil, errors.New("db connection refused")
		},
	}
	r := setupUserRouter(svc)

	req := httptest.NewRequest(http.MethodGet, "/user/data", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d — body: %s", w.Code, w.Body.String())
	}
	resp := parseResponse(t, w.Body)
	if resp["success"] == true {
		t.Error("expected success=false")
	}
}

func TestGetUserById_Success(t *testing.T) {
	now := time.Now()
	svc := &mockUserService{
		getUserByIdFn: func(_ context.Context, id int) (*domain.User, error) {
			return &domain.User{ID: "42", Mail: "user@test.com", CreatedAt: &now}, nil
		},
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := user_handlers.NewUserHandler(svc)
	for _, route := range h.Routes() {
		r.Handle(route.Method, "/user"+route.Path, func(c *gin.Context) {
			c.Set("user_id", 42)
			route.Handler(c)
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/user/data", nil)
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
