package service

import (
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"github.com/google/uuid"
	"time"
)

var (
	ErrUserNotAuthorized      = "user not authorized"
	ErrUserNotEnabled         = "user not enabled"
	ErrUserNotFoundByID       = "user not found by ID"
	ErrUserNotFoundByName     = "user not found by name"
	ErrUserNotFoundByEmail    = "user not found by email"
	ErrUserNotFoundByUsername = "user not found by username"
	ErrUserNotFoundByPhone    = "user not found by phone"
	ErrUserNotFoundByRole     = "user not found by role"
)

type AdminService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewAdminService(repo *repository.Queries, log *logger.Logger) *AdminService {
	return &AdminService{
		repo: repo,
		log:  log,
	}
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Firstname string    `json:"firstname"`
	LastName  string    `json:"lastname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Role      string    `json:"role"`
}
