package service

import (
	"context"
	"errors"

	"book-management/models"
	"book-management/repository"
	"book-management/utils"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(ctx context.Context, req models.UserRequest) (int64, error) {
	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return 0, err
	}

	user := models.User{
		Email:    req.Email,
		FullName: req.FullName,
		Role:     "user",
	}

	return s.repo.Create(ctx, user, hash)
}

func (s *UserService) Login(ctx context.Context, req models.UserLoginRequest) (string, models.User, error) {
	user, passwordHash, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return "", user, errors.New("invalid credentials")
	}

	if err := utils.CheckPassword(passwordHash, req.Password); err != nil {
		return "", user, errors.New("invalid credentials")
	}

	token, err := utils.CreateToken(user.ID, user.Role)
	if err != nil {
		return "", user, err
	}

	return token, user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (models.UserProfile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return models.UserProfile{}, err
	}

	return models.UserProfile{User: user}, nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.List(ctx)
}

func (s *UserService) UpdateRole(ctx context.Context, id int64, role string) error {
	switch role {
	case "user", "librarian", "admin":
	default:
		return errors.New("invalid role")
	}
	return s.repo.UpdateRole(ctx, id, role)
}