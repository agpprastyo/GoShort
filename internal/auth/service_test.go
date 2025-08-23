package auth

import (
	"GoShort/config"
	"GoShort/internal/commons"
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"GoShort/pkg/security"
	"GoShort/pkg/token"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository adalah implementasi mock untuk repository.Querier.
// Ia mengimplementasikan semua metode dari interface Querier.
type MockRepository struct {
	mock.Mock
}

// Memastikan MockRepository memenuhi interface repository.Querier saat kompilasi.
var _ repository.Querier = (*MockRepository)(nil)

// --- Implementasi Metode Mock untuk Auth ---
// Hanya metode yang relevan dengan Service yang memiliki implementasi detail.

func (m *MockRepository) GetUser(ctx context.Context, id uuid.UUID) (repository.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return repository.User{}, args.Error(1)
	}
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockRepository) GetUserByEmail(ctx context.Context, email string) (repository.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return repository.User{}, args.Error(1)
	}
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockRepository) GetUserByUsername(ctx context.Context, username string) (repository.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return repository.User{}, args.Error(1)
	}
	return args.Get(0).(repository.User), args.Error(1)
}

func (m *MockRepository) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return repository.User{}, args.Error(1)
	}
	return args.Get(0).(repository.User), args.Error(1)
}

// --- Metode Placeholder untuk Memenuhi Interface Querier ---
// Metode-metode berikut tidak digunakan dalam tes Service,
// jadi mereka hanya placeholder untuk memenuhi kontrak interface.

func (m *MockRepository) AdminGetShortLinkByID(ctx context.Context, id uuid.UUID) (repository.ShortLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) AdminGetShortLinksByUserID(ctx context.Context, arg repository.AdminGetShortLinksByUserIDParams) ([]repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.ShortLink), args.Error(1)
}

func (m *MockRepository) AdminListShortLinks(ctx context.Context, arg repository.AdminListShortLinksParams) ([]repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.ShortLink), args.Error(1)
}

func (m *MockRepository) AdminToggleShortLinkStatus(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) CheckShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	args := m.Called(ctx, shortCode)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) CountUserShortLinks(ctx context.Context, arg repository.CountUserShortLinksParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) CreateLinkStat(ctx context.Context, arg repository.CreateLinkStatParams) (repository.LinkStat, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.LinkStat), args.Error(1)
}

func (m *MockRepository) CreateShortLink(ctx context.Context, arg repository.CreateShortLinkParams) (repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) DeactivateShortLink(ctx context.Context, id uuid.UUID) (repository.ShortLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) DecrementClickLimit(ctx context.Context, id uuid.UUID) (repository.ShortLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteUserShortLink(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) GetActiveShortLinkByCode(ctx context.Context, shortCode string) (repository.ShortLink, error) {
	args := m.Called(ctx, shortCode)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) GetLinkStats(ctx context.Context, arg repository.GetLinkStatsParams) ([]repository.LinkStat, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.LinkStat), args.Error(1)
}

func (m *MockRepository) GetLinkStatsByDateRange(ctx context.Context, arg repository.GetLinkStatsByDateRangeParams) ([]repository.LinkStat, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.LinkStat), args.Error(1)
}

func (m *MockRepository) GetLinkStatsCount(ctx context.Context, arg repository.GetLinkStatsCountParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) GetLinkStatsGroupedByCountry(ctx context.Context, arg repository.GetLinkStatsGroupedByCountryParams) ([]repository.GetLinkStatsGroupedByCountryRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.GetLinkStatsGroupedByCountryRow), args.Error(1)
}

func (m *MockRepository) GetLinkStatsGroupedByDate(ctx context.Context, arg repository.GetLinkStatsGroupedByDateParams) ([]repository.GetLinkStatsGroupedByDateRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.GetLinkStatsGroupedByDateRow), args.Error(1)
}

func (m *MockRepository) GetShortLink(ctx context.Context, id uuid.UUID) (repository.ShortLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) GetShortLinkByCode(ctx context.Context, shortCode string) (repository.ShortLink, error) {
	args := m.Called(ctx, shortCode)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) ListShortLinks(ctx context.Context, arg repository.ListShortLinksParams) ([]repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.ShortLink), args.Error(1)
}

func (m *MockRepository) ListUserShortLinks(ctx context.Context, arg repository.ListUserShortLinksParams) ([]repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.ShortLink), args.Error(1)
}

func (m *MockRepository) ListUserShortLinksWithCountClick(ctx context.Context, arg repository.ListUserShortLinksWithCountClickParams) ([]repository.ListUserShortLinksWithCountClickRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.ListUserShortLinksWithCountClickRow), args.Error(1)
}

func (m *MockRepository) ListUsers(ctx context.Context, arg repository.ListUsersParams) ([]repository.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.User), args.Error(1)
}

func (m *MockRepository) ListUsersByRole(ctx context.Context, arg repository.ListUsersByRoleParams) ([]repository.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).([]repository.User), args.Error(1)
}

func (m *MockRepository) ToggleShortLinkStatus(ctx context.Context, id uuid.UUID) (repository.ShortLink, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) UpdateShortLink(ctx context.Context, arg repository.UpdateShortLinkParams) (repository.ShortLink, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.ShortLink), args.Error(1)
}

func (m *MockRepository) UpdateUser(ctx context.Context, arg repository.UpdateUserParams) (repository.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repository.User), args.Error(1)
}

// helper function untuk membuat user dummy
func createDummyUser(t *testing.T) (repository.User, string) {
	password := "password123"
	hashedPassword, err := security.HashPassword(password)
	require.NoError(t, err)

	return repository.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: hashedPassword,
		Role:         repository.UserRoleUser,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}, password
}

// newTestLogger membuat instance logger yang valid untuk pengujian
func newTestLogger() *logger.Logger {
	cfg := &config.AppConfig{
		Logger: config.LoggerConfig{
			Output:     io.Discard, // Mengirim log ke "tempat sampah"
			Level:      "info",
			JSONFormat: false,
		},
	}
	return logger.New(cfg)
}

// newTestJWT membuat instance JWTMaker untuk pengujian
func newTestJWT() *token.JWTMaker {
	cfg := &config.JWT{
		Secret:   "test-secret-key-that-is-long-enough",
		Issuer:   "test-issuer",
		Audience: "test-audience",
	}
	jwtMaker := token.NewJWTMaker(cfg)
	return jwtMaker
}

func TestGetProfileByID(t *testing.T) {
	user, _ := createDummyUser(t)

	testCases := []struct {
		name          string
		userID        uuid.UUID
		buildStubs    func(repo *MockRepository)
		checkResponse func(t *testing.T, profile *ProfileResponse, err error)
	}{
		{
			name:   "Success",
			userID: user.ID,
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUser", mock.Anything, user.ID).Return(user, nil).Times(1)
			},
			checkResponse: func(t *testing.T, profile *ProfileResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, profile)
				require.Equal(t, user.ID, profile.ID)
			},
		},
		{
			name:   "User Not Found",
			userID: uuid.New(),
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUser", mock.Anything, mock.Anything).Return(repository.User{}, pgx.ErrNoRows).Times(1)
			},
			checkResponse: func(t *testing.T, profile *ProfileResponse, err error) {
				require.Error(t, err)
				// NOTE: ErrUserNotFound harus di-export dari package service agar bisa diakses di sini.
				require.ErrorIs(t, err, commons.ErrUserNotFound)
				require.Nil(t, profile)
			},
		},
		{
			name:   "Database Error",
			userID: user.ID,
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUser", mock.Anything, user.ID).Return(repository.User{}, errors.New("database connection failed")).Times(1)
			},
			checkResponse: func(t *testing.T, profile *ProfileResponse, err error) {
				require.Error(t, err)
				require.NotErrorIs(t, err, commons.ErrUserNotFound)
				require.Nil(t, profile)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tc.buildStubs(mockRepo)

			log := newTestLogger()
			jwtMaker := newTestJWT()

			authService := NewService(mockRepo, jwtMaker, log)
			profile, err := authGetProfileByID(context.Background(), tc.userID)
			tc.checkResponse(t, profile, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	user, password := createDummyUser(t)

	testCases := []struct {
		name          string
		req           LoginRequest
		buildStubs    func(repo *MockRepository)
		checkResponse func(t *testing.T, res *LoginResponse, err error)
	}{
		{
			name: "Success",
			req: LoginRequest{
				Email:    user.Email,
				Password: password,
			},
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res *LoginResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.NotEmpty(t, res.Token)
				require.Equal(t, user.ID, res.Data.ID)
			},
		},
		{
			name: "Invalid Credentials - User Not Found",
			req: LoginRequest{
				Email:    "wrong@email.com",
				Password: password,
			},
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUserByEmail", mock.Anything, "wrong@email.com").Return(repository.User{}, pgx.ErrNoRows).Times(1)
			},
			checkResponse: func(t *testing.T, res *LoginResponse, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, commons.ErrInvalidCredentials)
				require.Nil(t, res)
			},
		},
		{
			name: "Invalid Credentials - Wrong Password",
			req: LoginRequest{
				Email:    user.Email,
				Password: "wrongpassword",
			},
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUserByEmail", mock.Anything, user.Email).Return(user, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res *LoginResponse, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, commons.ErrInvalidCredentials)
				require.Nil(t, res)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tc.buildStubs(mockRepo)

			log := newTestLogger()
			jwtMaker := newTestJWT()

			authService := NewService(mockRepo, jwtMaker, log)
			res, err := authLogin(context.Background(), tc.req)
			tc.checkResponse(t, res, err)

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRegister(t *testing.T) {
	user, _ := createDummyUser(t)
	registerReq := RegisterRequest{
		Username: user.Username,
		Email:    user.Email,
		Password: "password123",
	}

	testCases := []struct {
		name          string
		req           RegisterRequest
		buildStubs    func(repo *MockRepository)
		checkResponse func(t *testing.T, res *RegisterResponse, err error)
	}{
		{
			name: "Success",
			req:  registerReq,
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUserByEmail", mock.Anything, registerReq.Email).Return(repository.User{}, pgx.ErrNoRows).Times(1)
				repo.On("GetUserByUsername", mock.Anything, registerReq.Username).Return(repository.User{}, pgx.ErrNoRows).Times(1)
				repo.On("CreateUser", mock.Anything, mock.AnythingOfType("repository.CreateUserParams")).
					Return(user, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res *RegisterResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, user.Username, res.Username)
				require.Equal(t, user.Email, res.Email)
			},
		},
		{
			name: "User Already Exists - Email",
			req:  registerReq,
			buildStubs: func(repo *MockRepository) {
				repo.On("GetUserByEmail", mock.Anything, registerReq.Email).Return(user, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res *RegisterResponse, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, commons.ErrUserAlreadyExists)
				require.Nil(t, res)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			tc.buildStubs(mockRepo)

			log := newTestLogger()
			jwtMaker := newTestJWT()

			authService := NewService(mockRepo, jwtMaker, log)
			res, err := authRegister(context.Background(), tc.req)
			tc.checkResponse(t, res, err)

			mockRepo.AssertExpectations(t)
		})
	}
}
