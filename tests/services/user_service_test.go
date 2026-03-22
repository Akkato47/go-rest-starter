package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-starter/internal/core/domain"
	user_services "go-starter/internal/features/user/services"
)

type mockUserRepo struct {
	getUserByIdFn func(ctx context.Context, id int) (*domain.User, error)
}

func (m *mockUserRepo) GetUserById(ctx context.Context, id int) (*domain.User, error) {
	return m.getUserByIdFn(ctx, id)
}

func TestGetUserById_Success(t *testing.T) {
	now := time.Now()
	expected := &domain.User{ID: "42", Mail: "user@test.com", CreatedAt: &now}

	repo := &mockUserRepo{
		getUserByIdFn: func(_ context.Context, id int) (*domain.User, error) {
			if id != 42 {
				t.Errorf("expected id 42, got %d", id)
			}
			return expected, nil
		},
	}
	svc := user_services.NewUserService(repo)

	user, err := svc.GetUserById(context.Background(), 42)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "42" {
		t.Errorf("expected ID %q, got %q", "42", user.ID)
	}
	if user.Mail != "user@test.com" {
		t.Errorf("expected mail %q, got %q", "user@test.com", user.Mail)
	}
}

func TestGetUserById_NotFound(t *testing.T) {
	repo := &mockUserRepo{
		getUserByIdFn: func(_ context.Context, _ int) (*domain.User, error) {
			return nil, errors.New("no rows in result set")
		},
	}
	svc := user_services.NewUserService(repo)

	user, err := svc.GetUserById(context.Background(), 999)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Error("expected nil user on error")
	}
}

func TestGetUserById_RepoError(t *testing.T) {
	repo := &mockUserRepo{
		getUserByIdFn: func(_ context.Context, _ int) (*domain.User, error) {
			return nil, errors.New("connection refused")
		},
	}
	svc := user_services.NewUserService(repo)

	_, err := svc.GetUserById(context.Background(), 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
