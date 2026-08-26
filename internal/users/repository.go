package users

import (
	"context"
	"database/sql"
	"errors"
	"golang-api-template/internal/auth/dtos"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	GetAllAdmins(ctx context.Context) ([]User, error)
	Create(ctx context.Context, req *dtos.CreateUserRequest) (*User, error)
	Delete(ctx context.Context, id int64) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE email= $1`, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.SelectContext(ctx, &users, `SELECT * FROM users ORDER BY created_at DESC`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return users, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.SelectContext(ctx, &users, `SELECT * FROM users WHERE type = user ORDER BY created_at DESC`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return users, nil
}
func (r *userRepository) GetAllAdmins(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.SelectContext(ctx, &users, `SELECT * FROM users WHERE type = admin ORDER BY created_at DESC`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, req *dtos.CreateUserRequest) (*User, error) {
	var user User
	err := r.db.QueryRowxContext(ctx, `
		INSERT INTO users (fullname, Email ,password_hash)
		VALUES ($1, $2, $3)
		RETURNING *`,
		req.Fullname, req.Email, req.Password,
	).StructScan(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}
