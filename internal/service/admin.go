package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type IAdminService interface {
	ListAllLinks(ctx context.Context, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error)
	GetLinkByID(ctx context.Context, id uuid.UUID) (*dto.LinkResponse, error)
	ListUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error)
	ToggleLinkStatus(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context) (*dto.StatsResponse, error)
}

type AdminService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewAdminService(repo *repository.Queries, log *logger.Logger) IAdminService {
	return &AdminService{
		repo: repo,
		log:  log,
	}
}

// ListAllLinks retrieves all short links from the repository
func (s *AdminService) ListAllLinks(ctx context.Context, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error) {

	params := repository.AdminListShortLinksParams{
		SearchText: "",
		Limit:      10,
		Offset:     0,
		OrderBy:    repository.ShortlinkOrderColumnCreatedAt,
		Ascending:  false,
		StartDate:  pgtype.Timestamptz{}, // Default empty timestamp
		EndDate:    pgtype.Timestamptz{}, // Default empty timestamp

	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Limit != nil {
		params.Limit = *req.Limit
	} else {
		params.Limit = 10 // Default limit
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	} else {
		params.Offset = 0 // Default offset
	}

	if req.Order != nil {
		switch *req.Order {
		case "title":
			params.OrderBy = repository.ShortlinkOrderColumnTitle
		case "is_active":
			params.OrderBy = repository.ShortlinkOrderColumnIsActive
		case "created_at":
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = repository.ShortlinkOrderColumnUpdatedAt
		case "expired_at":
			params.OrderBy = repository.ShortlinkOrderColumnExpiredAt
		default:
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt // Default order
		}
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	} else {
		params.Ascending = false // Default descending order
	}

	if req.StartDate != nil {
		// Convert time.Time to pgtype.Timestamptz
		startTime := pgtype.Timestamptz{Time: *req.StartDate, Valid: true}
		params.StartDate = startTime
	} else {
		params.StartDate = pgtype.Timestamptz{} // Default empty timestamp
	}

	if req.EndDate != nil {
		// Convert time.Time to pgtype.Timestamptz
		endTime := pgtype.Timestamptz{Time: *req.EndDate, Valid: true}

		params.EndDate = endTime
	} else {
		params.EndDate = pgtype.Timestamptz{} // Default empty timestamp
	}

	links, err := s.repo.AdminListShortLinks(ctx, params)
	if err != nil {
		s.log.Error("failed to list short links", "error", err)
		return nil, nil, err
	}

	response := make([]dto.LinkResponse, len(links))

	for i, link := range links {
		response[i] = dto.LinkResponse{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortCode:   link.ShortCode,
			Title:       link.Title,
			IsActive:    link.IsActive,
			ClickLimit:  link.ClickLimit,
			ExpireAt:    link.ExpiredAt.Time,
			CreatedAt:   link.CreatedAt.Time,
			UpdatedAt:   link.UpdatedAt.Time,
		}
	}

	pagination := &dto.Pagination{
		Total:   len(response),
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: len(response) > int(params.Limit),
	}

	return response, pagination, nil
}

// GetLinkByID retrieves a specific short link by ID
func (s *AdminService) GetLinkByID(ctx context.Context, id uuid.UUID) (*dto.LinkResponse, error) {
	link, err := s.repo.AdminGetShortLinkByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get short link", "error", err)
		return nil, err
	}

	response := &dto.LinkResponse{
		ID:          link.ID,
		OriginalURL: link.OriginalUrl,
		ShortCode:   link.ShortCode,
		Title:       link.Title,
		IsActive:    link.IsActive,
		ClickLimit:  link.ClickLimit,
		ExpireAt:    link.ExpiredAt.Time,
		CreatedAt:   link.CreatedAt.Time,
		UpdatedAt:   link.UpdatedAt.Time,
	}

	return response, nil
}

// ListUserLinks retrieves all short links for a specific user
func (s *AdminService) ListUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error) {
	params := repository.AdminGetShortLinksByUserIDParams{
		UserID:     userID,
		SearchText: "",
		Limit:      10,
		Offset:     0,
		OrderBy:    repository.ShortlinkOrderColumnCreatedAt,
		Ascending:  false,
		StartDate:  pgtype.Timestamptz{}, // Default empty timestamp
		EndDate:    pgtype.Timestamptz{}, // Default empty timestamp
	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Limit != nil {
		params.Limit = *req.Limit
	} else {
		params.Limit = 10 // Default limit
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	} else {
		params.Offset = 0 // Default offset
	}

	if req.Order != nil {
		switch *req.Order {
		case "title":
			params.OrderBy = repository.ShortlinkOrderColumnTitle
		case "is_active":
			params.OrderBy = repository.ShortlinkOrderColumnIsActive
		case "created_at":
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = repository.ShortlinkOrderColumnUpdatedAt
		case "expired_at":
			params.OrderBy = repository.ShortlinkOrderColumnExpiredAt
		default:
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt // Default order
		}
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	} else {
		params.Ascending = false // Default descending order
	}

	if req.StartDate != nil {
		// Convert time.Time to pgtype.Timestamptz
		startTime := pgtype.Timestamptz{Time: *req.StartDate, Valid: true}
		params.StartDate = startTime
	} else {
		endTime := pgtype.Timestamptz{Time: *req.EndDate, Valid: true}
		params.StartDate = endTime
	}

	if req.EndDate != nil {
		endTime := pgtype.Timestamptz{Time: *req.EndDate, Valid: true}
		params.EndDate = endTime
	} else {
		params.EndDate = pgtype.Timestamptz{} // Default empty timestamp
	}

	userLinks, err := s.repo.AdminGetShortLinksByUserID(ctx, params)
	if err != nil {
		s.log.Error("failed to list user short links", "error", err)
		return nil, nil, err
	}
	response := make([]dto.LinkResponse, len(userLinks))
	for i, link := range userLinks {
		response[i] = dto.LinkResponse{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortCode:   link.ShortCode,
			Title:       link.Title,
			IsActive:    link.IsActive,
			ClickLimit:  link.ClickLimit,
			ExpireAt:    link.ExpiredAt.Time,
			CreatedAt:   link.CreatedAt.Time,
			UpdatedAt:   link.UpdatedAt.Time,
		}
	}

	pagination := &dto.Pagination{
		Total:   len(response),
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: len(response) > int(params.Limit),
	}

	return response, pagination, nil
}

// ToggleLinkStatus toggles the active status of a short link
func (s *AdminService) ToggleLinkStatus(ctx context.Context, id uuid.UUID) error {
	err := s.repo.AdminToggleShortLinkStatus(ctx, id)
	if err != nil {
		s.log.Error("failed to toggle short link status", "error", err)
		return err
	}
	return nil
}

func (s *AdminService) GetStats(ctx context.Context) (*dto.StatsResponse, error) {
	usersCh := make(chan int64, 1)
	linksCh := make(chan int64, 1)
	activeCh := make(chan int64, 1)
	inactiveCh := make(chan int64, 1)
	errCh := make(chan error, 4)

	go func() {
		u, err := s.repo.CountUsers(ctx)
		if err != nil {
			errCh <- err
			return
		}
		usersCh <- u
	}()

	go func() {
		l, err := s.repo.CountLinks(ctx)
		if err != nil {
			errCh <- err
			return
		}
		linksCh <- l
	}()

	go func() {
		a, err := s.repo.CountActiveLinks(ctx)
		if err != nil {
			errCh <- err
			return
		}
		activeCh <- a
	}()

	go func() {
		i, err := s.repo.CountInactiveLinks(ctx)
		if err != nil {
			errCh <- err
			return
		}
		inactiveCh <- i
	}()

	var (
		users, links, active, inactive int64
	)
	for i := 0; i < 4; i++ {
		select {
		case err := <-errCh:
			s.log.Error("failed to count stats", "error", err)
			return nil, err
		case users = <-usersCh:
		case links = <-linksCh:
		case active = <-activeCh:
		case inactive = <-inactiveCh:
		}
	}
	return &dto.StatsResponse{
		TotalUsers:    users,
		TotalLinks:    links,
		ActiveLinks:   active,
		InactiveLinks: inactive,
	}, nil
}
