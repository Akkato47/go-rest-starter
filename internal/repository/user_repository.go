package repository

import (
	"context"
	"fmt"
	"go-starter/internal/model"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *model.User) (*model.User, error) {
	if user.Mail == "" {
		return nil, fmt.Errorf("mail can not be empty")
	}

	query := `
		INSERT INTO users (mail, password)
		VALUES ($1, $2)
		RETURNING id, mail, created_at, updated_at;
	`

	var createdUser model.User

	err := pool.QueryRow(ctx, query, user.Mail, user.Password).Scan(
		&createdUser.ID,
		&createdUser.Mail,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23505":
				return nil, fmt.Errorf("user with such mail already exists")
			default:
				return nil, fmt.Errorf("database error: %s", pgErr.Message)
			}
		}
		return nil, fmt.Errorf("error while creating user: %w", err)
	}

	return &createdUser, nil
}

func GetUserByMail(ctx context.Context, pool *pgxpool.Pool, mail string) (*model.User, error) {
	query := `
		SELECT id, mail, password
		FROM users
		WHERE mail = $1
	`

	var user model.User

	err := pool.QueryRow(ctx, query, mail).Scan(
		&user.ID,
		&user.Mail,
		&user.Password,
	)
	if err != nil {
		return &model.User{}, err
	}
	return &user, nil
}

func GetUserById(ctx context.Context, pool *pgxpool.Pool, id int) (*model.User, error) {
	query := `
		SELECT id, mail, created_at
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Mail,
		&user.CreatedAt,
	)
	if err != nil {
		return &model.User{}, err
	}
	return &user, nil
}
