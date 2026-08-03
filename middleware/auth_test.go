package middleware

import (
	"book-management/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	utils.SetJWTSecret("test-secret-key-for-middleware-tests")
}

func nextHandler(t *testing.T, called *bool, gotID *int64, gotRole *string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		*called = true
		*gotID = GetUserID(r)
		*gotRole = GetUserRole(r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}
}

func decodeError(t *testing.T, body []byte) utils.ErrorResponse {
	t.Helper()
	var resp utils.ErrorResponse
	err := json.Unmarshal(body, &resp)
	require.NoError(t, err)
	return resp
}

func TestAuth_MissingHeader(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, called)

	resp := decodeError(t, rr.Body.Bytes())
	assert.Equal(t, "Authorization header required", resp.Message)
}

func TestAuth_InvalidTokenFormat(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc.def.ghi")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, called)

	resp := decodeError(t, rr.Body.Bytes())
	assert.Equal(t, "Invalid token format", resp.Message)
}

func TestAuth_InvalidToken(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.token")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, called)

	resp := decodeError(t, rr.Body.Bytes())
	assert.Equal(t, "Invalid or expired token", resp.Message)
}

func TestAuth_ValidToken(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	token, err := utils.CreateToken(42, "librarian")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)
	assert.Equal(t, int64(42), gotID)
	assert.Equal(t, "librarian", gotRole)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestAuth_ValidToken_UserRole(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	token, err := utils.CreateToken(7, "user")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, called)
	assert.Equal(t, int64(7), gotID)
	assert.Equal(t, "user", gotRole)
}

func TestGetUserID_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, int64(0), GetUserID(req))
}

func TestGetUserRole_NoContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	assert.Equal(t, "user", GetUserRole(req))
}

func TestGetUserID_And_GetUserRole_FromContext(t *testing.T) {
	var called bool
	var gotID int64
	var gotRole string

	handler := Auth(nextHandler(t, &called, &gotID, &gotRole))

	token, err := utils.CreateToken(99, "admin")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	require.True(t, called)
	assert.Equal(t, int64(99), gotID)
	assert.Equal(t, "admin", gotRole)
}
