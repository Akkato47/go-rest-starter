package auth_services

import (
	"context"
	"go-starter/internal/core/config"
	"go-starter/internal/core/domain"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByMail(ctx context.Context, mail string) (*domain.User, error)
}

type svc struct {
	repo UserRepository
	cfg  *config.Config
}

func NewAuthService(repo UserRepository, cfg *config.Config) *svc {
	return &svc{
		repo: repo,
		cfg:  cfg,
	}
}
