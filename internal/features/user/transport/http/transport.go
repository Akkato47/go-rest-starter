package user_handlers

import (
	"context"
	"go-starter/internal/core/domain"
	"go-starter/internal/core/transport/http/server"
	"net/http"
)

type UserService interface {
	GetUserById(ctx context.Context, userId int) (*domain.User, error)
}

type UserHandler interface {
	Routes() []server.Route
}

type userHandler struct {
	service UserService
}

func NewUserHandler(service UserService) UserHandler {
	return &userHandler{service: service}
}

func (h *userHandler) Routes() []server.Route {
	return []server.Route{
		{Method: http.MethodGet, Path: "/data", Handler: h.getUserById},
	}
}
