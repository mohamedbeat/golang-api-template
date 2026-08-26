package users

import (
	"context"
	"errors"
	"golang-api-template/internal/auth/dtos"

	"github.com/google/uuid"
)

type UserService interface {
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetAllUsers(ctx context.Context) ([]User, error)
	CreateUser(ctx context.Context, req *dtos.CreateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id int64) error
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {

	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.GetByEmail(ctx, email)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.repo.GetAll(ctx)
}

func (s *userService) CreateUser(ctx context.Context, req *dtos.CreateUserRequest) (*User, error) {
	return s.repo.Create(ctx, req)
}

func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("invalid user id")
	}
	return s.repo.Delete(ctx, id)
}
