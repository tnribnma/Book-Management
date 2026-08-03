package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"book-management/models"
	"book-management/service"
	"book-management/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(ctx context.Context, email, password, fullName string) (*models.User, error) {
	args := m.Called(ctx, email, password, fullName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int64, fullName, role string) error {
	args := m.Called(ctx, id, fullName, role)
	return args.Error(0)
}

var _ service.UserService = (*MockUserService)(nil)

func sampleUser() *models.User {
	return &models.User{
		ID:        1,
		Email:     "alice@example.com",
		FullName:  "Alice",
		Role:      "user",
		CreatedAt: time.Now(),
	}
}

func init() {
	utils.SetJWTSecret("test-secret-key-for-unit-tests")
}

func TestRegister_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	body := models.UserRequest{
		Email:    "alice@example.com",
		Password: "secret123",
		FullName: "Alice",
	}
	raw, _ := json.Marshal(body)

	mockSvc.On("Register", mock.Anything, body.Email, body.Password, body.FullName).
		Return(sampleUser(), nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var resp map[string]any
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, float64(1), resp["id"])
	assert.Equal(t, "alice@example.com", resp["email"])
	mockSvc.AssertExpectations(t)
}

func TestRegister_InvalidJSON(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(`{invalid`)))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid request body", resp["error"])
	mockSvc.AssertNotCalled(t, "Register")
}

func TestRegister_ServiceError(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	body := models.UserRequest{
		Email:    "alice@example.com",
		Password: "secret123",
		FullName: "Alice",
	}
	raw, _ := json.Marshal(body)

	mockSvc.On("Register", mock.Anything, body.Email, body.Password, body.FullName).
		Return(nil, errors.New("user with this email already exists")).Once()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(raw))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "user with this email already exists", resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestLogin_Success(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	body := models.UserLoginRequest{
		Email:    "alice@example.com",
		Password: "secret123",
	}
	raw, _ := json.Marshal(body)
	user := sampleUser()

	mockSvc.On("Login", mock.Anything, body.Email, body.Password).
		Return(user, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(raw))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]any
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)

	assert.NotEmpty(t, resp["token"])
	assert.NotNil(t, resp["user"])
	mockSvc.AssertExpectations(t)
}

func TestLogin_InvalidJSON(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(`not-json`)))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid request body", resp["error"])
	mockSvc.AssertNotCalled(t, "Login")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	mockSvc := new(MockUserService)
	h := NewAuthHandler(mockSvc)

	body := models.UserLoginRequest{
		Email:    "alice@example.com",
		Password: "wrong",
	}
	raw, _ := json.Marshal(body)

	mockSvc.On("Login", mock.Anything, body.Email, body.Password).
		Return(nil, errors.New("invalid email or password")).Once()

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(raw))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid credentials", resp["error"])
	mockSvc.AssertExpectations(t)
}
