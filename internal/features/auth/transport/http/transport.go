package auth_handlers

import (
	"context"
	"go-starter/internal/core/domain"
	"go-starter/internal/core/transport/http/server"
)

type AuthService interface {
	Register(ctx context.Context, mail, password string) (*domain.User, string, error)
	Login(ctx context.Context, mail, password string) (*domain.User, string, error)
}

type AuthHandler interface {
	Routes() []server.Route
}
