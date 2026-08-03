package service

import (
	"book-management/models"
	"book-management/repository"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) List(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

var _ repository.UserRepository = (*MockUserRepository)(nil)

func hashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

func sampleStoredUser(t *testing.T) *models.User {
	t.Helper()
	return &models.User{
		ID:           1,
		Email:        "alice@example.com",
		PasswordHash: hashPassword(t, "secret123"),
		FullName:     "Alice",
		Role:         "user",
		CreatedAt:    time.Now(),
	}
}

func TestRegister_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "alice@example.com").
		Return(nil, errors.New("user not found")).Once()

	mockRepo.On("Create", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.Email == "alice@example.com" &&
			u.FullName == "Alice" &&
			u.Role == "user" &&
			u.PasswordHash != ""
	})).Run(func(args mock.Arguments) {
		u := args.Get(1).(*models.User)
		u.ID = 1
		u.CreatedAt = time.Now()
	}).Return(nil).Once()

	user, err := svc.Register(ctx, "alice@example.com", "secret123", "Alice")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "Alice", user.FullName)
	assert.Equal(t, "user", user.Role)
	assert.Empty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestRegister_EmptyEmailOrPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	user, err := svc.Register(ctx, "", "secret123", "Alice")
	assert.Nil(t, user)
	assert.EqualError(t, err, "email and password are required")

	user, err = svc.Register(ctx, "alice@example.com", "", "Alice")
	assert.Nil(t, user)
	assert.EqualError(t, err, "email and password are required")

	mockRepo.AssertNotCalled(t, "GetByEmail")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestRegister_PasswordTooShort(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	user, err := svc.Register(ctx, "alice@example.com", "12345", "Alice")

	assert.Nil(t, user)
	assert.EqualError(t, err, "password must be at least 6 characters")
	mockRepo.AssertNotCalled(t, "GetByEmail")
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	existing := sampleStoredUser(t)
	mockRepo.On("GetByEmail", ctx, "alice@example.com").Return(existing, nil).Once()

	user, err := svc.Register(ctx, "alice@example.com", "secret123", "Alice")

	assert.Nil(t, user)
	assert.EqualError(t, err, "user with this email already exists")
	mockRepo.AssertNotCalled(t, "Create")
}

func TestRegister_CreateError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "alice@example.com").
		Return(nil, errors.New("user not found")).Once()
	mockRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
		Return(errors.New("db error")).Once()

	user, err := svc.Register(ctx, "alice@example.com", "secret123", "Alice")

	assert.Nil(t, user)
	assert.ErrorContains(t, err, "failed to register user")
	mockRepo.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	stored := sampleStoredUser(t)
	mockRepo.On("GetByEmail", ctx, "alice@example.com").Return(stored, nil).Once()

	user, err := svc.Login(ctx, "alice@example.com", "secret123")

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Equal(t, "Alice", user.FullName)
	assert.Empty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByEmail", ctx, "missing@example.com").
		Return(nil, errors.New("user not found")).Once()

	user, err := svc.Login(ctx, "missing@example.com", "secret123")

	assert.Nil(t, user)
	assert.EqualError(t, err, "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestLogin_WrongPassword(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	stored := sampleStoredUser(t)
	mockRepo.On("GetByEmail", ctx, "alice@example.com").Return(stored, nil).Once()

	user, err := svc.Login(ctx, "alice@example.com", "wrongpassword")

	assert.Nil(t, user)
	assert.EqualError(t, err, "invalid email or password")
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	stored := sampleStoredUser(t)
	mockRepo.On("GetByID", ctx, int64(1)).Return(stored, nil).Once()

	user, err := svc.GetProfile(ctx, 1)

	require.NoError(t, err)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.Empty(t, user.PasswordHash)
	mockRepo.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, int64(99)).
		Return(nil, errors.New("user not found")).Once()

	user, err := svc.GetProfile(ctx, 99)

	assert.Nil(t, user)
	assert.EqualError(t, err, "user not found")
	mockRepo.AssertExpectations(t)
}

func TestListUsers_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	users := []models.User{
		{ID: 1, Email: "alice@example.com", PasswordHash: "hash1", FullName: "Alice", Role: "user"},
		{ID: 2, Email: "bob@example.com", PasswordHash: "hash2", FullName: "Bob", Role: "librarian"},
	}
	mockRepo.On("List", ctx).Return(users, nil).Once()

	result, err := svc.ListUsers(ctx)

	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Empty(t, result[0].PasswordHash)
	assert.Empty(t, result[1].PasswordHash)
	assert.Equal(t, "alice@example.com", result[0].Email)
	mockRepo.AssertExpectations(t)
}

func TestListUsers_Error(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctx).Return(nil, errors.New("db error")).Once()

	result, err := svc.ListUsers(ctx)

	assert.Nil(t, result)
	assert.EqualError(t, err, "db error")
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_Success(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == 1 && u.FullName == "Alice Updated" && u.Role == "librarian"
	})).Return(nil).Once()

	err := svc.UpdateUser(ctx, 1, "Alice Updated", "librarian")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_EmptyRole_Allowed(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
		return u.ID == 1 && u.Role == ""
	})).Return(nil).Once()

	err := svc.UpdateUser(ctx, 1, "Alice", "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateUser_InvalidRole(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	err := svc.UpdateUser(ctx, 1, "Alice", "superadmin")

	assert.EqualError(t, err, "invalid role")
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateUser_ValidRoles(t *testing.T) {
	roles := []string{"user", "librarian", "admin"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			svc := NewUserService(mockRepo)
			ctx := context.Background()

			mockRepo.On("Update", ctx, mock.MatchedBy(func(u *models.User) bool {
				return u.Role == role
			})).Return(nil).Once()

			err := svc.UpdateUser(ctx, 1, "Name", role)
			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateUser_RepoError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := NewUserService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Update", ctx, mock.AnythingOfType("*models.User")).
		Return(errors.New("user not found")).Once()

	err := svc.UpdateUser(ctx, 99, "Nobody", "user")

	assert.EqualError(t, err, "user not found")
	mockRepo.AssertExpectations(t)
}
