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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserServiceForUserHandler struct {
	mock.Mock
}

func (m *MockUserServiceForUserHandler) Register(ctx context.Context, email, password, fullName string) (*models.User, error) {
	args := m.Called(ctx, email, password, fullName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserServiceForUserHandler) Login(ctx context.Context, email, password string) (*models.User, error) {
	args := m.Called(ctx, email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserServiceForUserHandler) GetProfile(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserServiceForUserHandler) ListUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserServiceForUserHandler) UpdateUser(ctx context.Context, id int64, fullName, role string) error {
	args := m.Called(ctx, id, fullName, role)
	return args.Error(0)
}

var _ service.UserService = (*MockUserServiceForUserHandler)(nil)

func sampleUserForHandler() *models.User {
	return &models.User{
		ID:        1,
		Email:     "alice@example.com",
		FullName:  "Alice",
		Role:      "user",
		CreatedAt: time.Now(),
	}
}

func TestGetProfile_Success(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)
	user := sampleUserForHandler()

	mockSvc.On("GetProfile", mock.Anything, int64(0)).Return(user, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.User
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, user.Email, resp.Email)
	assert.Equal(t, user.FullName, resp.FullName)
	mockSvc.AssertExpectations(t)
}

func TestGetProfile_NotFound(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	mockSvc.On("GetProfile", mock.Anything, int64(0)).
		Return(nil, errors.New("user not found")).Once()

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	rr := httptest.NewRecorder()

	h.GetProfile(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "user not found", resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestListUsers_Success(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	users := []models.User{*sampleUserForHandler()}
	mockSvc.On("ListUsers", mock.Anything).Return(users, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp []models.User
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "alice@example.com", resp[0].Email)
	mockSvc.AssertExpectations(t)
}

func TestListUsers_ServiceError(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	mockSvc.On("ListUsers", mock.Anything).
		Return(nil, errors.New("db error")).Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()

	h.ListUsers(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "failed to fetch users", resp["error"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateRole_Success(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)
	user := sampleUserForHandler()

	body := map[string]string{"role": "librarian"}
	raw, _ := json.Marshal(body)

	mockSvc.On("GetProfile", mock.Anything, int64(1)).Return(user, nil).Once()
	mockSvc.On("UpdateUser", mock.Anything, int64(1), user.FullName, "librarian").
		Return(nil).Once()

	req := httptest.NewRequest(http.MethodPut, "/users/1/role", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "role updated", resp["message"])
	mockSvc.AssertExpectations(t)
}

func TestUpdateRole_InvalidID(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/users/abc/role", bytes.NewReader([]byte(`{"role":"admin"}`)))
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid user id", resp["error"])
	mockSvc.AssertNotCalled(t, "GetProfile")
	mockSvc.AssertNotCalled(t, "UpdateUser")
}

func TestUpdateRole_ZeroID(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/users/0/role", bytes.NewReader([]byte(`{"role":"admin"}`)))
	req.SetPathValue("id", "0")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid user id", resp["error"])
}

func TestUpdateRole_InvalidJSON(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	req := httptest.NewRequest(http.MethodPut, "/users/1/role", bytes.NewReader([]byte(`{bad`)))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid request body", resp["error"])
	mockSvc.AssertNotCalled(t, "UpdateUser")
}

func TestUpdateRole_EmptyRole(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	body := map[string]string{"role": ""}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/users/1/role", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "role is required", resp["error"])
	mockSvc.AssertNotCalled(t, "UpdateUser")
}

func TestUpdateRole_UserNotFound(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)

	body := map[string]string{"role": "admin"}
	raw, _ := json.Marshal(body)

	mockSvc.On("GetProfile", mock.Anything, int64(99)).
		Return(nil, errors.New("not found")).Once()

	req := httptest.NewRequest(http.MethodPut, "/users/99/role", bytes.NewReader(raw))
	req.SetPathValue("id", "99")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "user not found", resp["error"])
	mockSvc.AssertNotCalled(t, "UpdateUser")
}

func TestUpdateRole_ServiceError(t *testing.T) {
	mockSvc := new(MockUserServiceForUserHandler)
	h := NewUserHandler(mockSvc)
	user := sampleUserForHandler()

	body := map[string]string{"role": "superadmin"}
	raw, _ := json.Marshal(body)

	mockSvc.On("GetProfile", mock.Anything, int64(1)).Return(user, nil).Once()
	mockSvc.On("UpdateUser", mock.Anything, int64(1), user.FullName, "superadmin").
		Return(errors.New("invalid role")).Once()

	req := httptest.NewRequest(http.MethodPut, "/users/1/role", bytes.NewReader(raw))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.UpdateRole(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp map[string]string
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	assert.Equal(t, "invalid role", resp["error"])
	mockSvc.AssertExpectations(t)
}
