package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-starter/internal/core/config"
	"go-starter/internal/core/transport/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

const testSecret = "test-secret-key-32chars-long-xx"

func testCfg() *config.Config {
	return &config.Config{JwtSecret: testSecret}
}

func makeToken(t *testing.T, userID string, secret string, exp time.Time) string {
	t.Helper()
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     exp.Unix(),
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func authMiddlewareRouter(cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", middleware.AuthMiddleware(cfg), func(c *gin.Context) {
		id := c.GetInt("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": id})
	})
	return r
}

func TestAuthMiddleware_NoCookie(t *testing.T) {
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_EmptyCookie(t *testing.T) {
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "not.a.jwt"})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	token := makeToken(t, "1", "wrong-secret-key-32chars-long-xx", time.Now().Add(time.Hour))
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	token := makeToken(t, "1", testSecret, time.Now().Add(-time.Hour))
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	token := makeToken(t, "42", testSecret, time.Now().Add(time.Hour))
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_NonNumericUserID(t *testing.T) {
	token := makeToken(t, "not-a-number", testSecret, time.Now().Add(time.Hour))
	r := authMiddlewareRouter(testCfg())

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
