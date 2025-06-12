package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockAuthService adalah implementasi mock dari IAuthService untuk pengujian.
type mockAuthService struct {
	RegisterFunc       func(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	LoginFunc          func(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GetProfileByIDFunc func(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error)
}

// Baris ini adalah pengecekan waktu kompilasi untuk memastikan
// mockAuthService benar-benar memenuhi kontrak service.IAuthService.
// Pastikan interface IAuthService sudah didefinisikan di package service.
var _ service.IAuthService = (*mockAuthService)(nil)

func (m *mockAuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, req)
	}
	return nil, errors.New("RegisterFunc not implemented on mock")
}

func (m *mockAuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, req)
	}
	return nil, errors.New("LoginFunc not implemented on mock")
}

func (m *mockAuthService) GetProfileByID(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error) {
	if m.GetProfileByIDFunc != nil {
		return m.GetProfileByIDFunc(ctx, id)
	}
	return nil, errors.New("GetProfileByIDFunc not implemented on mock")
}

func TestAuthHandler_Register(t *testing.T) {
	testUserID := uuid.New()
	testCases := []struct {
		name           string
		requestBody    map[string]string
		setupMock      func(mock *mockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Success",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.RegisterFunc = func(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
					return &dto.RegisterResponse{
						UserID:   testUserID.String(),
						Username: "testuser",
						Email:    "test@example.com",
						Role:     "user",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"message":"User registered successfully"`,
		},
		{
			name: "User Already Exists",
			requestBody: map[string]string{
				"username": "existinguser",
				"email":    "existing@example.com",
				"password": "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.RegisterFunc = func(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
					return nil, service.ErrUserAlreadyExists
				}
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"User already exists"}`,
		},
		{
			name: "Invalid Request Body - Missing Field",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock:      func(mock *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Email, password, and username are required"}`,
		},
		{
			name: "Internal Server Error",
			requestBody: map[string]string{
				"username": "testuser",
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.RegisterFunc = func(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
					return nil, errors.New("some unexpected error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Server error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &mockAuthService{}
			tc.setupMock(mockService)

			// Setelah Anda mengubah NewAuthHandler untuk menerima interface, baris ini akan valid.
			authHandler := NewAuthHandler(mockService)

			app := fiber.New()
			app.Post("/auth/register", authHandler.Register)

			// Buat request
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Eksekusi request
			resp, err := app.Test(req, -1) // -1 untuk tidak ada timeout
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	testUserID := uuid.New()

	testCases := []struct {
		name           string
		requestBody    map[string]string
		setupMock      func(mock *mockAuthService)
		expectedStatus int
		expectedBody   string
		expectCookie   bool
	}{
		{
			name: "Success",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func(mock *mockAuthService) {
				mock.LoginFunc = func(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
					return &dto.LoginResponse{
						Token:     "fake-jwt-token",
						ExpiresAt: time.Now().Add(time.Hour).Unix(),
						Data: dto.ProfileResponse{
							ID:       testUserID,
							Username: "testuser",
							Email:    "test@example.com",
							Role:     "user",
						},
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"logged_in":true`,
			expectCookie:   true,
		},
		{
			name: "Invalid Credentials",
			requestBody: map[string]string{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			setupMock: func(mock *mockAuthService) {
				mock.LoginFunc = func(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
					return nil, service.ErrInvalidCredentials
				}
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid email or password"}`,
			expectCookie:   false,
		},
		{
			name: "Missing Password",
			requestBody: map[string]string{
				"email": "test@example.com",
			},
			setupMock:      func(mock *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Email and password are required"}`,
			expectCookie:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &mockAuthService{}
			tc.setupMock(mockService)
			authHandler := NewAuthHandler(mockService)
			app := fiber.New()
			app.Post("/auth/login", authHandler.Login)

			// Buat request
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			// Eksekusi
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(respBody), tc.expectedBody)

			// Cek cookie
			cookieHeader := resp.Header.Get("Set-Cookie")
			if tc.expectCookie {
				require.NotEmpty(t, cookieHeader)
				require.Contains(t, cookieHeader, "access_token=fake-jwt-token")
			} else {
				require.Empty(t, cookieHeader)
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	testUserID := uuid.New()

	testCases := []struct {
		name           string
		userIDLocal    interface{} // Apa yang diset di c.Locals("userID")
		setupMock      func(mock *mockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userIDLocal: testUserID.String(),
			setupMock: func(mock *mockAuthService) {
				mock.GetProfileByIDFunc = func(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error) {
					require.Equal(t, testUserID, id)
					return &dto.ProfileResponse{
						ID:       testUserID,
						Username: "testuser",
						Email:    "test@example.com",
						Role:     "user",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"username":"testuser"`,
		},
		{
			name:           "Unauthorized - No UserID in Locals",
			userIDLocal:    nil, // Tidak ada userID
			setupMock:      func(mock *mockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Unauthorized"}`,
		},
		{
			name:           "Invalid User ID Format",
			userIDLocal:    "not-a-uuid",
			setupMock:      func(mock *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID"}`,
		},
		{
			name:        "User Not Found",
			userIDLocal: testUserID.String(),
			setupMock: func(mock *mockAuthService) {
				mock.GetProfileByIDFunc = func(ctx context.Context, id uuid.UUID) (*dto.ProfileResponse, error) {
					return nil, service.ErrUserNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"User not found"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockService := &mockAuthService{}
			tc.setupMock(mockService)
			authHandler := NewAuthHandler(mockService)
			app := fiber.New()

			// Middleware dummy untuk set c.Locals
			app.Use(func(c *fiber.Ctx) error {
				if tc.userIDLocal != nil {
					c.Locals("userID", tc.userIDLocal)
				}
				return c.Next()
			})
			app.Get("/auth/profile", authHandler.GetProfile)

			// Buat request
			req := httptest.NewRequest(http.MethodGet, "/auth/profile", nil)

			// Eksekusi
			resp, err := app.Test(req, -1)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Assertions
			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			require.Contains(t, string(body), tc.expectedBody)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Setup
		authHandler := NewAuthHandler(nil) // Tidak perlu service untuk logout
		app := fiber.New()
		app.Post("/auth/logout", authHandler.Logout)

		// Buat request
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		// Eksekusi
		resp, err := app.Test(req, -1)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Assertions
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Cek body
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, `{"logged_out":true, "message":"Successfully logged out"}`, string(body))

		// Cek cookie yang sudah dihapus
		cookieHeader := resp.Header.Get("Set-Cookie")
		require.NotEmpty(t, cookieHeader)
		require.Contains(t, cookieHeader, "access_token=;")
		// PERBAIKAN: Gunakan 'expires' (lowercase) sesuai standar header HTTP
		require.Contains(t, cookieHeader, "expires=")
	})
}
