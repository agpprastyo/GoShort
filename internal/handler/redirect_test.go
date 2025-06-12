package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"

	"context"
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

// mockRedirectService adalah implementasi mock dari IRedirectService untuk pengujian.
type mockRedirectService struct {
	GetOriginalURLFunc func(ctx context.Context, code string) (string, uuid.UUID, bool, error)
	RecordLinkStatFunc func(ctx context.Context, linkID uuid.UUID, req dto.CreateLinkStatRequest) error
}

// Memastikan mockRedirectService memenuhi kontrak service.IRedirectService.
var _ service.IRedirectService = (*mockRedirectService)(nil)

func (m *mockRedirectService) GetOriginalURL(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
	return m.GetOriginalURLFunc(ctx, code)
}

func (m *mockRedirectService) RecordLinkStat(ctx context.Context, linkID uuid.UUID, req dto.CreateLinkStatRequest) error {
	return m.RecordLinkStatFunc(ctx, linkID, req)
}

func TestRedirectHandler_RedirectToOriginalURL(t *testing.T) {
	linkID := uuid.New()
	originalURL := "https://example.com/very/long/url"
	testCode := "abcdef"

	// NOTE: Untuk isolasi test yang sempurna, ubah `func fetchIPInfo...` menjadi
	// `var fetchIPInfo = func...` di redirect.go. Ini akan memungkinkan Anda
	// untuk mem-mock panggilan jaringan eksternal. Untuk saat ini, kita akan
	// menghapus mock untuk memperbaiki kompilasi.
	//
	// originalFetchIPInfo := fetchIPInfo
	// fetchIPInfo = func(ipAddress string) (*dto.IPAPIResponse, error) {
	// 	return &dto.IPAPIResponse{Status: "success", Country: "Test Country"}, nil
	// }
	// t.Cleanup(func() {
	// 	fetchIPInfo = originalFetchIPInfo
	// })

	testCases := []struct {
		name                 string
		codeParam            string
		setupMock            func(mock *mockRedirectService, recordCalled chan bool)
		expectedStatus       int
		expectedLocation     string
		expectedBodyContains string
		verifyGoroutine      bool
	}{
		{
			name:      "Success",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					require.Equal(t, testCode, code)
					return originalURL, linkID, true, nil
				}
				mock.RecordLinkStatFunc = func(ctx context.Context, id uuid.UUID, req dto.CreateLinkStatRequest) error {
					require.Equal(t, linkID, id)
					// Menandakan bahwa fungsi ini telah dipanggil
					recordCalled <- true
					return nil
				}
			},
			expectedStatus:   http.StatusFound, // 302
			expectedLocation: originalURL,
			verifyGoroutine:  true,
		},
		{
			name:      "Link Not Found",
			codeParam: "notfound",
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					return "", uuid.Nil, false, service.ErrLinkNotFound
				}
			},
			expectedStatus:       http.StatusNotFound,
			expectedBodyContains: "Link not found",
		},
		{
			name:      "Link Not Active",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					// Skenario 1: Service mengembalikan error
					return "", uuid.Nil, false, service.ErrLinkNotActive
				}
			},
			expectedStatus:       http.StatusForbidden,
			expectedBodyContains: "Link is inactive",
		},
		{
			name:      "Link Not Active - Second Check",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					// Skenario 2: Service mengembalikan isActive = false
					return originalURL, linkID, false, nil
				}
			},
			expectedStatus:       http.StatusForbidden,
			expectedBodyContains: "Link is inactive",
		},
		{
			name:      "Link Expired",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					return "", uuid.Nil, false, service.ErrLinkExpired
				}
			},
			expectedStatus:       http.StatusGone,
			expectedBodyContains: "Link has expired",
		},
		{
			name:      "Click Limit Exceeded",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					return "", uuid.Nil, false, service.ErrClickLimitExceeded
				}
			},
			expectedStatus:       http.StatusTooManyRequests,
			expectedBodyContains: "Click limit exceeded",
		},
		{
			name:      "Generic Service Error",
			codeParam: testCode,
			setupMock: func(mock *mockRedirectService, recordCalled chan bool) {
				mock.GetOriginalURLFunc = func(ctx context.Context, code string) (string, uuid.UUID, bool, error) {
					return "", uuid.Nil, false, errors.New("database connection lost")
				}
			},
			expectedStatus:       http.StatusInternalServerError,
			expectedBodyContains: "Internal app error",
		},
		{
			name:                 "Empty Code Param",
			codeParam:            "",
			setupMock:            func(mock *mockRedirectService, recordCalled chan bool) {},
			expectedStatus:       http.StatusNotFound,
			expectedBodyContains: "Cannot GET /",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &mockRedirectService{}
			recordCalled := make(chan bool, 1)
			tc.setupMock(mockService, recordCalled)

			handler := NewRedirectHandler(mockService, newTestLogger())

			app := fiber.New()
			app.Get("/:code", handler.RedirectToOriginalURL)

			req := httptest.NewRequest(http.MethodGet, "/"+tc.codeParam, nil)
			// FIX: app.Test timeout adalah int dalam milidetik, bukan time.Duration.
			resp, err := app.Test(req, 10000) // 10 detik timeout
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedLocation != "" {
				location, err := resp.Location()
				require.NoError(t, err)
				require.Equal(t, tc.expectedLocation, location.String())
			}

			if tc.expectedBodyContains != "" {
				body, _ := io.ReadAll(resp.Body)
				require.Contains(t, string(body), tc.expectedBodyContains)
			}

			if tc.verifyGoroutine {
				// Pastikan goroutine untuk mencatat statistik telah dipanggil
				select {
				case <-recordCalled:
					// Sukses, fungsi dipanggil
				case <-time.After(2 * time.Second): // Perpanjang waktu tunggu karena ada panggilan jaringan
					t.Fatal("expected RecordLinkStat to be called, but it was not")
				}
			}
		})
	}
}
