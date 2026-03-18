package user_services

import (
	"context"
	"go-starter/internal/core/domain"
)

type UserRepository interface {
	GetUserById(ctx context.Context, id int) (*domain.User, error)
}

type svc struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *svc {
	return &svc{
		repo: repo,
	}
}
