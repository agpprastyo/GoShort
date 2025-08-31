package admin

import (
	"GoShort/config"
	"GoShort/internal/shortlink"
	"GoShort/pkg/helper"
	"GoShort/pkg/logger" // Diperlukan untuk inisialisasi logger
	"context"
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

// mockAdminService adalah implementasi mock dari IService untuk pengujian.
type mockAdminService struct {
	ListAllLinksFunc     func(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error)
	GetLinkByIDFunc      func(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error)
	ListUserLinksFunc    func(ctx context.Context, userID uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error)
	ToggleLinkStatusFunc func(ctx context.Context, id uuid.UUID) error
}

// Memastikan mockAdminService memenuhi kontrak service.IService.
var _ IService = (*mockAdminService)(nil)

func (m *mockAdminService) ListAllLinks(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
	return m.ListAllLinksFunc(ctx, req)
}

func (m *mockAdminService) GetLinkByID(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error) {
	return m.GetLinkByIDFunc(ctx, id)
}

func (m *mockAdminService) ListUserLinks(ctx context.Context, userID uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
	return m.ListUserLinksFunc(ctx, userID, req)
}

func (m *mockAdminService) ToggleLinkStatus(ctx context.Context, id uuid.UUID) error {
	return m.ToggleLinkStatusFunc(ctx, id)
}

// newTestLogger membuat instance logger yang valid untuk pengujian yang membuang semua output.
func newTestLogger() *logger.Logger {
	// FIX: Inisialisasi config dengan benar dan teruskan sebagai pointer.
	// Ini menyelesaikan error kompilasi dan mencegah panic saat runtime.
	cfg := &config.AppConfig{
		Logger: config.LoggerConfig{
			Output:     io.Discard, // Mengirim log ke "tempat sampah"
			Level:      "info",
			JSONFormat: false,
		},
	}
	return logger.New(cfg)
}

func TestAdminHandler_ListAllLinks(t *testing.T) {
	mockLinks := []shortlink.LinkResponse{
		{ID: uuid.New(), ShortCode: "abc", OriginalURL: "https://example.com/1"},
	}
	mockPagination := &helper.Pagination{Total: 1, Limit: 10, Offset: 0}

	testCases := []struct {
		name           string
		queryParams    string
		setupMock      func(mock *mockAdminService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success - No Params",
			queryParams: "",
			setupMock: func(mock *mockAdminService) {
				mock.ListAllLinksFunc = func(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
					return mockLinks, mockPagination, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"short_code":"abc"`,
		},
		{
			name:        "Success - With Params",
			queryParams: "?limit=5&offset=10&search=test&order_by=created_at&ascending=true&start_date=2023-01-01T00:00:00Z&end_date=2023-12-31T23:59:59Z",
			setupMock: func(mock *mockAdminService) {
				mock.ListAllLinksFunc = func(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
					// Verifikasi bahwa parameter telah di-parse dengan benar
					require.NotNil(t, req.Limit)
					require.Equal(t, int64(5), *req.Limit)
					require.NotNil(t, req.Offset)
					require.Equal(t, int64(10), *req.Offset)
					require.NotNil(t, req.Search)
					require.Equal(t, "test", *req.Search)
					return mockLinks, mockPagination, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total":1`,
		},
		{
			name:        "Service Error",
			queryParams: "",
			setupMock: func(mock *mockAdminService) {
				mock.ListAllLinksFunc = func(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
					return nil, nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"database error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockAdminService{}
			tc.setupMock(mockService)
			// FIX: Inisialisasi handler dengan logger yang valid untuk mencegah panic.
			handler := NewHandler(mockService, newTestLogger())

			app := fiber.New()
			app.Get("/admin/links", handler.ListAllLinks)

			req := httptest.NewRequest(http.MethodGet, "/admin/links"+tc.queryParams, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(body), tc.expectedBody)
		})
	}
}

func TestAdminHandler_GetLink(t *testing.T) {
	linkID := uuid.New()
	mockLink := &shortlink.LinkResponse{ID: linkID, ShortCode: "xyz", OriginalURL: "https://example.com/2"}

	testCases := []struct {
		name           string
		linkIDParam    string
		setupMock      func(mock *mockAdminService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			linkIDParam: linkID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.GetLinkByIDFunc = func(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error) {
					require.Equal(t, linkID, id)
					return mockLink, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"short_code":"xyz"`,
		},
		{
			name:           "Invalid Link ID",
			linkIDParam:    "not-a-uuid",
			setupMock:      func(mock *mockAdminService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid link ID"}`,
		},
		{
			name:        "Service Error",
			linkIDParam: linkID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.GetLinkByIDFunc = func(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error) {
					return nil, errors.New("not found")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"not found"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockAdminService{}
			tc.setupMock(mockService)
			// FIX: Inisialisasi handler dengan logger yang valid untuk mencegah panic.
			handler := NewHandler(mockService, newTestLogger())

			app := fiber.New()
			app.Get("/admin/links/:id", handler.GetLink)

			req := httptest.NewRequest(http.MethodGet, "/admin/links/"+tc.linkIDParam, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(body), tc.expectedBody)
		})
	}
}

func TestAdminHandler_ListUserLinks(t *testing.T) {
	userID := uuid.New()
	mockLinks := []shortlink.LinkResponse{
		{ID: uuid.New(), ShortCode: "userlink", OriginalURL: "https://user.com/1"},
	}
	mockPagination := &helper.Pagination{Total: 1, Limit: 10, Offset: 0}

	testCases := []struct {
		name           string
		userIDParam    string
		setupMock      func(mock *mockAdminService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			userIDParam: userID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.ListUserLinksFunc = func(ctx context.Context, id uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
					require.Equal(t, userID, id)
					return mockLinks, mockPagination, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"short_code":"userlink"`,
		},
		{
			name:           "Invalid User ID",
			userIDParam:    "bad-uuid",
			setupMock:      func(mock *mockAdminService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid user ID"}`,
		},
		{
			name:        "Service Error",
			userIDParam: userID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.ListUserLinksFunc = func(ctx context.Context, id uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
					return nil, nil, errors.New("user link error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"user link error"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockAdminService{}
			tc.setupMock(mockService)
			// FIX: Inisialisasi handler dengan logger yang valid untuk mencegah panic.
			handler := NewHandler(mockService, newTestLogger())

			app := fiber.New()
			app.Get("/admin/users/:userId/links", handler.ListUserLinks)

			url := fmt.Sprintf("/admin/users/%s/links", tc.userIDParam)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)
			body, _ := io.ReadAll(resp.Body)
			require.Contains(t, string(body), tc.expectedBody)
		})
	}
}

func TestAdminHandler_ToggleLinkStatus(t *testing.T) {
	linkID := uuid.New()

	testCases := []struct {
		name           string
		linkIDParam    string
		setupMock      func(mock *mockAdminService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success",
			linkIDParam: linkID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.ToggleLinkStatusFunc = func(ctx context.Context, id uuid.UUID) error {
					require.Equal(t, linkID, id)
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   ``, // Body harus kosong untuk 204
		},
		{
			name:           "Invalid Link ID",
			linkIDParam:    "not-a-uuid",
			setupMock:      func(mock *mockAdminService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Invalid link ID"}`,
		},
		{
			name:        "Service Error",
			linkIDParam: linkID.String(),
			setupMock: func(mock *mockAdminService) {
				mock.ToggleLinkStatusFunc = func(ctx context.Context, id uuid.UUID) error {
					return errors.New("toggle failed")
				}
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"toggle failed"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockAdminService{}
			tc.setupMock(mockService)
			// FIX: Inisialisasi handler dengan logger yang valid untuk mencegah panic.
			handler := NewHandler(mockService, newTestLogger())

			app := fiber.New()
			app.Patch("/admin/links/:id/toggle", handler.ToggleLinkStatus)

			url := fmt.Sprintf("/admin/links/%s/toggle", tc.linkIDParam)
			req := httptest.NewRequest(http.MethodPatch, url, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			// Untuk 204 No Content, body harus kosong. Untuk kasus lain, cek isinya.
			if tc.expectedStatus != http.StatusNoContent {
				body, _ := io.ReadAll(resp.Body)
				require.Contains(t, string(body), tc.expectedBody)
			}
		})
	}
}
