package user_services

import (
	"context"
	"fmt"
	"go-starter/internal/core/domain"
)

func (s *svc) GetUserById(ctx context.Context, userId int) (*domain.User, error) {
	userData, err := s.repo.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("Error while get user data: %s", err.Error())
	}
	return userData, nil
}
