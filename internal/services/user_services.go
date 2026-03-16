package services

import (
	"context"
	"fmt"
	"go-starter/internal/model"
	"go-starter/internal/repository"
)

type UserService interface {
	GetUserById(ctx context.Context, userId int) (*model.User, error)
}

type usvc struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &usvc{
		repo: repo,
	}
}

func (s *usvc) GetUserById(ctx context.Context, userId int) (*model.User, error) {
	userData, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("Error while get user data: %s", err.Error())
	}
	return userData, nil
}
