package users

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `db:"id"         json:"id"`
	FullName string    `db:"fullname"       json:"fullname"`
	Email    string    `db:"email"       json:"email"`
	// Username      string    `db:"username"      json:"username"`
	Password_hash string    `db:"password_hash" json:"-"`
	Type          string    `db:"type" json:"type"` // type=(admin, user)
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}
