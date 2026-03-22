package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"go-starter/internal/core/config"
	"go-starter/internal/core/domain"
	auth_services "go-starter/internal/features/auth/services"

	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepo struct {
	createUserFn    func(ctx context.Context, user *domain.User) (*domain.User, error)
	getUserByMailFn func(ctx context.Context, mail string) (*domain.User, error)
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return m.createUserFn(ctx, user)
}

func (m *mockAuthRepo) GetUserByMail(ctx context.Context, mail string) (*domain.User, error) {
	return m.getUserByMailFn(ctx, mail)
}

func testCfg() *config.Config {
	return &config.Config{JwtSecret: "test-secret-key-32chars-long-xx"}
}

func fakeUser(id, mail string) *domain.User {
	now := time.Now()
	return &domain.User{ID: id, Mail: mail, CreatedAt: &now}
}

func bcryptHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(hash)
}

func TestRegister_PasswordTooShort(t *testing.T) {
	svc := auth_services.NewAuthService(&mockAuthRepo{}, testCfg())

	_, _, err := svc.Register(context.Background(), "a@b.com", "12345")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "6 characters") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRegister_RepoError(t *testing.T) {
	repo := &mockAuthRepo{
		createUserFn: func(_ context.Context, _ *domain.User) (*domain.User, error) {
			return nil, errors.New("db error")
		},
	}
	svc := auth_services.NewAuthService(repo, testCfg())

	_, _, err := svc.Register(context.Background(), "a@b.com", "password123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRegister_Success(t *testing.T) {
	repo := &mockAuthRepo{
		createUserFn: func(_ context.Context, user *domain.User) (*domain.User, error) {
			return fakeUser("1", user.Mail), nil
		},
	}
	svc := auth_services.NewAuthService(repo, testCfg())

	user, token, err := svc.Register(context.Background(), "new@test.com", "password123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.Mail != "new@test.com" {
		t.Errorf("expected mail %q, got %q", "new@test.com", user.Mail)
	}
	if token == "" {
		t.Fatal("expected JWT token, got empty string")
	}
	if user.Password != "" {
		t.Error("password must not be present in returned user")
	}
}

func TestRegister_PasswordLengthBoundary(t *testing.T) {
	repo := &mockAuthRepo{
		createUserFn: func(_ context.Context, user *domain.User) (*domain.User, error) {
			return fakeUser("1", user.Mail), nil
		},
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"empty", "", true},
		{"5 chars", "12345", true},
		{"exactly 6", "123456", false},
		{"long password", "superstrongpassword123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := auth_services.NewAuthService(repo, testCfg())
			_, _, err := svc.Register(context.Background(), "a@b.com", tt.password)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for password %q", tt.password)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for password %q: %v", tt.password, err)
			}
		})
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockAuthRepo{
		getUserByMailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, errors.New("no rows")
		},
	}
	svc := auth_services.NewAuthService(repo, testCfg())

	_, _, err := svc.Login(context.Background(), "noone@test.com", "password")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash := bcryptHash(t, "correctpassword")

	repo := &mockAuthRepo{
		getUserByMailFn: func(_ context.Context, mail string) (*domain.User, error) {
			u := fakeUser("1", mail)
			u.Password = hash
			return u, nil
		},
	}
	svc := auth_services.NewAuthService(repo, testCfg())

	_, _, err := svc.Login(context.Background(), "a@b.com", "wrongpassword")

	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if !strings.Contains(err.Error(), "credentials") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	const password = "correctpassword"
	hash := bcryptHash(t, password)

	repo := &mockAuthRepo{
		getUserByMailFn: func(_ context.Context, mail string) (*domain.User, error) {
			u := fakeUser("1", mail)
			u.Password = hash
			return u, nil
		},
	}
	svc := auth_services.NewAuthService(repo, testCfg())

	user, token, err := svc.Login(context.Background(), "a@b.com", password)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if token == "" {
		t.Fatal("expected JWT token, got empty string")
	}
}
