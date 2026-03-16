package services

import (
	"context"
	"fmt"
	"go-starter/internal/config"
	"go-starter/internal/model"
	"go-starter/internal/repository"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, mail, password string) (*model.User, string, error)
	Login(ctx context.Context, mail, password string) (*model.User, string, error)
}

type asvc struct {
	repo repository.UserRepository
	cfg  *config.Config
}

func NewAuthService(repo repository.UserRepository, cfg *config.Config) AuthService {
	return &asvc{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *asvc) Register(ctx context.Context, mail, password string) (*model.User, string, error) {
	if len(password) < 6 {
		return nil, "", fmt.Errorf("Password must be at least 6 characters long")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to hash password: %s", err.Error())
	}

	user := &model.User{
		Mail:     mail,
		Password: string(hashedPassword),
	}

	createdUser, err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create user: %s", err.Error())
	}

	tokenExp := time.Now().Add(1 * time.Hour)

	accessTokenString, err := generateJWT(createdUser.ID, tokenExp, s.cfg)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to generate token: %s", err.Error())
	}

	return createdUser, accessTokenString, nil
}

func (s *asvc) Login(ctx context.Context, mail, password string) (*model.User, string, error) {
	user, err := s.repo.GetUserByMail(ctx, mail)
	if err != nil {
		return nil, "", fmt.Errorf("Something went wrong: %s", err.Error())
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", fmt.Errorf("Invalid credentials")
	}

	tokenExp := time.Now().Add(1 * time.Hour)

	accessTokenString, err := generateJWT(user.ID, tokenExp, s.cfg)
	if err != nil {
		return nil, "", fmt.Errorf("Failed to generate token: %s", err.Error())
	}

	return user, accessTokenString, nil
}

func generateJWT(userId string, expTime time.Time, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"exp":     expTime.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return accessToken.SignedString([]byte(cfg.JwtSecret))
}
