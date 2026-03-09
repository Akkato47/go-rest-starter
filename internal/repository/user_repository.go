package repository

import (
	"context"
	"fmt"
	"go-starter/internal/model"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func CreateUser(conn *pgx.Conn, user *model.User) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Проверка обязательных полей
	if user.Mail == "" {
		return nil, fmt.Errorf("email не может быть пустым")
	}

	query := `
		INSERT INTO users (mail, password)
		VALUES ($1, $2)
		RETURNING id, mail, created_at, updated_at;
	`

	var createdUser model.User

	err := conn.QueryRow(ctx, query, user.Mail, user.Password).Scan(
		&createdUser.ID,
		&createdUser.Mail,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			switch pgErr.Code {
			case "23505": // unique_violation
				return nil, fmt.Errorf("пользователь с таким email уже существует")
			default:
				return nil, fmt.Errorf("ошибка базы данных: %s", pgErr.Message)
			}
		}
		return nil, fmt.Errorf("не удалось создать пользователя: %w", err)
	}

	return &createdUser, nil
}

func GetuserByMail(conn *pgx.Conn, mail string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	query := `
		SELECT id, mail
		FROM users
		WHERE mail = $1
	`

	var user model.User

	err := conn.QueryRow(ctx, query, mail).Scan(
		&user.ID,
		&user.Mail,
	)
	if err != nil {
		return &model.User{}, err
	}
	return &user, nil
}

func GetUserById(conn *pgx.Conn, id string) (*model.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	query := `
		SELECT id, mail
		FROM users
		WHERE id = $1
	`

	var user model.User

	err := conn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Mail,
	)
	if err != nil {
		return &model.User{}, err
	}
	return &user, nil
}
