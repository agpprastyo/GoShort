package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"GoShort/pkg/helper"

	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockShortLinkService is a mock implementation of IShortLinkService for testing.
type mockShortLinkService struct {
	CreateLinkFromDTOFunc    func(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error)
	GetUserLinksFunc         func(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error)
	GetUserLinkByIDFunc      func(ctx context.Context, userID, linkID uuid.UUID) (*dto.LinkResponse, error)
	UpdateUserLinkFunc       func(ctx context.Context, userID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error)
	DeleteUserLinkFunc       func(ctx context.Context, userID, linkID uuid.UUID) error
	ToggleUserLinkStatusFunc func(ctx context.Context, userID, linkID uuid.UUID) (*dto.LinkResponse, error)
	ShortCodeExistsFunc      func(ctx context.Context, code string) (bool, error)
}

// Ensure mockShortLinkService implements the service.IShortLinkService interface.
var _ service.IShortLinkService = (*mockShortLinkService)(nil)

func (m *mockShortLinkService) CreateLinkFromDTO(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error) {
	return m.CreateLinkFromDTOFunc(ctx, userID, req)
}

func (m *mockShortLinkService) GetUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error) {
	return m.GetUserLinksFunc(ctx, userID, req)
}

func (m *mockShortLinkService) GetUserLinkByID(ctx context.Context, userID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	return m.GetUserLinkByIDFunc(ctx, userID, linkID)
}

func (m *mockShortLinkService) UpdateUserLink(ctx context.Context, userID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error) {
	return m.UpdateUserLinkFunc(ctx, userID, linkID, req)
}

func (m *mockShortLinkService) DeleteUserLink(ctx context.Context, userID, linkID uuid.UUID) error {
	return m.DeleteUserLinkFunc(ctx, userID, linkID)
}

func (m *mockShortLinkService) ToggleUserLinkStatus(ctx context.Context, userID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	return m.ToggleUserLinkStatusFunc(ctx, userID, linkID)
}

func (m *mockShortLinkService) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	if m.ShortCodeExistsFunc != nil {
		return m.ShortCodeExistsFunc(ctx, code)
	}
	return false, errors.New("ShortCodeExistsFunc not implemented on mock")
}

// setupAppWithUserID is a helper to create a Fiber app instance and a dummy middleware
// that sets the user ID in the context for testing authenticated routes.
func setupAppWithUserID(handler *ShortLinkHandler, userID string) *fiber.App {
	app := fiber.New(fiber.Config{
		// Add a custom error handler to prevent panics in tests when
		// the handler doesn't handle a specific error case.
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(func(c *fiber.Ctx) error {
		// Add a nil check to prevent panic if middleware is used without a userID
		if userID != "" {
			c.Locals("user_id", userID)
		}
		return c.Next()
	})
	return app
}

func TestShortLinkHandler_CreateShortLink(t *testing.T) {
	userID := uuid.New()
	mockResponse := &dto.LinkResponse{ID: uuid.New(), OriginalURL: "https://example.com"}

	testCases := []struct {
		name           string
		userID         string
		requestBody    interface{}
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userID:      userID.String(),
			requestBody: dto.CreateLinkRequest{OriginalURL: "https://example.com"},
			setupMock: func(mock *mockShortLinkService) {
				mock.CreateLinkFromDTOFunc = func(ctx context.Context, id uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error) {
					require.Equal(t, userID, id)
					return mockResponse, nil
				}
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `"original_url":"https://example.com"`,
		},
		{
			name:           "Invalid Request Body",
			userID:         userID.String(),
			requestBody:    "not json",
			setupMock:      func(mock *mockShortLinkService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid request body"}`,
		},
		{
			name:           "Missing Original URL",
			userID:         userID.String(),
			requestBody:    dto.CreateLinkRequest{},
			setupMock:      func(mock *mockShortLinkService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Original URL is required"}`,
		},
		{
			name:           "Invalid User ID",
			userID:         "not-a-uuid",
			requestBody:    dto.CreateLinkRequest{OriginalURL: "https://example.com"},
			setupMock:      func(mock *mockShortLinkService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Post("/links", handler.CreateShortLink)

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}

func TestShortLinkHandler_GetUserLinks(t *testing.T) {
	userID := uuid.New()
	mockLinks := []dto.LinkResponse{{ID: uuid.New(), OriginalURL: "https://test.com"}}
	mockPagination := &dto.Pagination{Total: 1, Limit: 10, Offset: 0}

	testCases := []struct {
		name           string
		userID         string
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: userID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.GetUserLinksFunc = func(ctx context.Context, id uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error) {
					require.Equal(t, userID, id)
					return mockLinks, mockPagination, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"original_url":"https://test.com"`,
		},
		{
			name:           "Invalid User ID",
			userID:         "bad-uuid",
			setupMock:      func(mock *mockShortLinkService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Get("/links", handler.GetUserLinks)

			req := httptest.NewRequest(http.MethodGet, "/links", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}

func TestShortLinkHandler_GetUserLinkByID(t *testing.T) {
	userID := uuid.New()
	linkID := uuid.New()
	mockResponse := &dto.LinkResponse{ID: linkID, OriginalURL: "https://specific.com"}

	testCases := []struct {
		name           string
		userID         string
		linkID         string
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.GetUserLinkByIDFunc = func(ctx context.Context, uID, lID uuid.UUID) (*dto.LinkResponse, error) {
					require.Equal(t, userID, uID)
					require.Equal(t, linkID, lID)
					return mockResponse, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"original_url":"https://specific.com"`,
		},
		{
			name:   "Not Found",
			userID: userID.String(),
			linkID: uuid.New().String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.GetUserLinkByIDFunc = func(ctx context.Context, uID, lID uuid.UUID) (*dto.LinkResponse, error) {
					return nil, service.ErrLinkNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Short link not found"}`,
		},
		{
			name:   "Unauthorized",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.GetUserLinkByIDFunc = func(ctx context.Context, uID, lID uuid.UUID) (*dto.LinkResponse, error) {
					return nil, service.ErrUnauthorized
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"You are not authorized to access this link"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Get("/links/:id", handler.GetUserLinkByID)

			url := fmt.Sprintf("/links/%s", tc.linkID)
			req := httptest.NewRequest(http.MethodGet, url, nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}

func TestShortLinkHandler_UpdateLink(t *testing.T) {
	userID := uuid.New()
	linkID := uuid.New()
	// FIX: Create a string variable to get its pointer for the request DTO.
	newTitle := "New Title"
	updateReq := dto.UpdateLinkRequest{Title: &newTitle}
	mockResponse := &dto.LinkResponse{ID: linkID, Title: helper.StringToPtr("New Title")}

	testCases := []struct {
		name           string
		userID         string
		linkID         string
		requestBody    interface{}
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userID:      userID.String(),
			linkID:      linkID.String(),
			requestBody: updateReq,
			setupMock: func(mock *mockShortLinkService) {
				mock.UpdateUserLinkFunc = func(ctx context.Context, uID, lID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error) {
					require.Equal(t, userID, uID)
					require.Equal(t, linkID, lID)
					// FIX: Dereference the pointer to check the string value.
					require.NotNil(t, req.Title)
					require.Equal(t, "New Title", *req.Title)
					return mockResponse, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"title":"New Title"`,
		},
		{
			name:           "Invalid Link ID",
			userID:         userID.String(),
			linkID:         "bad-uuid",
			requestBody:    updateReq,
			setupMock:      func(mock *mockShortLinkService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid link ID"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Put("/links/:id", handler.UpdateLink)

			url := fmt.Sprintf("/links/%s", tc.linkID)
			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}

func TestShortLinkHandler_DeleteLink(t *testing.T) {
	userID := uuid.New()
	linkID := uuid.New()

	testCases := []struct {
		name           string
		userID         string
		linkID         string
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.DeleteUserLinkFunc = func(ctx context.Context, uID, lID uuid.UUID) error {
					require.Equal(t, userID, uID)
					require.Equal(t, linkID, lID)
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "Link Not Found",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.DeleteUserLinkFunc = func(ctx context.Context, uID, lID uuid.UUID) error {
					return service.ErrLinkNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"Short link not found"}`,
		},
		{
			name:   "Unauthorized",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.DeleteUserLinkFunc = func(ctx context.Context, uID, lID uuid.UUID) error {
					return service.ErrUnauthorized
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `{"error":"You are not authorized to delete this link"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Delete("/links/:id", handler.DeleteLink)

			url := fmt.Sprintf("/links/%s", tc.linkID)
			req := httptest.NewRequest(http.MethodDelete, url, nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			if tc.expectedBody != "" {
				respBody, _ := io.ReadAll(resp.Body)
				require.JSONEq(t, tc.expectedBody, string(respBody))
			}
		})
	}
}

func TestShortLinkHandler_ToggleLinkStatus(t *testing.T) {
	userID := uuid.New()
	linkID := uuid.New()
	mockResponse := &dto.LinkResponse{ID: linkID, IsActive: true}

	testCases := []struct {
		name           string
		userID         string
		linkID         string
		setupMock      func(mock *mockShortLinkService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Success",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.ToggleUserLinkStatusFunc = func(ctx context.Context, uID, lID uuid.UUID) (*dto.LinkResponse, error) {
					require.Equal(t, userID, uID)
					require.Equal(t, linkID, lID)
					return mockResponse, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"is_active":true`,
		},
		{
			name:   "Service Error",
			userID: userID.String(),
			linkID: linkID.String(),
			setupMock: func(mock *mockShortLinkService) {
				mock.ToggleUserLinkStatusFunc = func(ctx context.Context, uID, lID uuid.UUID) (*dto.LinkResponse, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockShortLinkService{}
			tc.setupMock(mockService)
			handler := NewShortLinkHandler(mockService, newTestLogger())
			app := setupAppWithUserID(handler, tc.userID)
			app.Patch("/links/:id/toggle", handler.ToggleLinkStatus)

			url := fmt.Sprintf("/links/%s/toggle", tc.linkID)
			req := httptest.NewRequest(http.MethodPatch, url, nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			respBody, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(respBody), tc.expectedBody)
		})
	}
}
