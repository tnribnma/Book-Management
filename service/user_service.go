package service

import (
	"context"
	"fmt"
	"book-management/models"
	"book-management/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password, fullName string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
	GetProfile(ctx context.Context, userID int64) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
	UpdateUser(ctx context.Context, id int64, fullName, role string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) Register(ctx context.Context, email, password, fullName string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password are required")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if existing != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password")
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		Role:         "user", 
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to register user: %w", err)
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) Login(ctx context.Context, email, password string) (*models.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *userService) ListUsers(ctx context.Context) ([]models.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

func (s *userService) UpdateUser(ctx context.Context, id int64, fullName, role string) error {
	if role != "" && role != "user" && role != "librarian" && role != "admin" {
		return fmt.Errorf("invalid role")
	}

	user := &models.User{
		ID:       id,
		FullName: fullName,
		Role:     role,
	}

	return s.userRepo.Update(ctx, user)

}