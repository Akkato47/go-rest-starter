package model

import (
	"time"
)

type User struct {
	ID        string     `json:"id" db:"id"`
	Mail      string     `json:"mail" db:"mail"`
	Password  string     `json:"-" db:"password"`
	CreatedAt *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
