package admin

import (
	"GoShort/internal/datastore"
	"GoShort/pkg/helper"

	"GoShort/internal/shortlink"
	"GoShort/internal/stats"
	"GoShort/pkg/logger"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type IService interface {
	ListAllLinks(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error)
	GetLinkByID(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error)
	ListUserLinks(ctx context.Context, userID uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error)
	ToggleLinkStatus(ctx context.Context, id uuid.UUID) error
	GetStats(ctx context.Context) (*stats.StatsResponse, error)
}

type Service struct {
	repo datastore.Querier
	log  *logger.Logger
}

func NewService(repo datastore.Querier, log *logger.Logger) IService {
	return &Service{
		repo: repo,
		log:  log,
	}
}

// ListAllLinks retrieves all short links from the datastore
func (s *Service) ListAllLinks(ctx context.Context, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
	params := datastore.AdminListShortLinksParams{
		SearchText: "",
		Limit:      10,
		Offset:     0,
		OrderBy:    datastore.ShortlinkOrderColumnCreatedAt,
		Ascending:  false,
		StartDate:  pgtype.Timestamptz{},
		EndDate:    pgtype.Timestamptz{},
	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Limit != nil {
		params.Limit = *req.Limit
	} else {
		params.Limit = 10
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	} else {
		params.Offset = 0
	}

	if req.Order != nil {
		switch *req.Order {
		case "title":
			params.OrderBy = datastore.ShortlinkOrderColumnTitle
		case "is_active":
			params.OrderBy = datastore.ShortlinkOrderColumnIsActive
		case "created_at":
			params.OrderBy = datastore.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = datastore.ShortlinkOrderColumnUpdatedAt
		case "expired_at":
			params.OrderBy = datastore.ShortlinkOrderColumnExpiredAt
		default:
			params.OrderBy = datastore.ShortlinkOrderColumnCreatedAt
		}
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	} else {
		params.Ascending = false
	}

	if req.StartDate != nil {
		startTime := pgtype.Timestamptz{Time: *req.StartDate, Valid: true}
		params.StartDate = startTime
	} else {
		params.StartDate = pgtype.Timestamptz{}
	}

	if req.EndDate != nil {
		endTime := pgtype.Timestamptz{Time: *req.EndDate, Valid: true}
		params.EndDate = endTime
	} else {
		params.EndDate = pgtype.Timestamptz{}
	}

	links, err := s.repo.AdminListShortLinks(ctx, params)
	if err != nil {
		s.log.Error("failed to list short links", "error", err)
		return nil, nil, err
	}

	response := make([]shortlink.LinkResponse, len(links))

	for i, link := range links {
		response[i] = shortlink.LinkResponse{
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

	pagination := &helper.Pagination{
		Total:   len(response),
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: len(response) > int(params.Limit),
	}

	return response, pagination, nil
}

// GetLinkByID retrieves a specific short link by ID
func (s *Service) GetLinkByID(ctx context.Context, id uuid.UUID) (*shortlink.LinkResponse, error) {
	link, err := s.repo.AdminGetShortLinkByID(ctx, id)
	if err != nil {
		s.log.Error("failed to get short link", "error", err)
		return nil, err
	}

	response := &shortlink.LinkResponse{
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
func (s *Service) ListUserLinks(ctx context.Context, userID uuid.UUID, req shortlink.GetLinksRequest) ([]shortlink.LinkResponse, *helper.Pagination, error) {
	params := datastore.AdminGetShortLinksByUserIDParams{
		UserID:     userID,
		SearchText: "",
		Limit:      10,
		Offset:     0,
		OrderBy:    datastore.ShortlinkOrderColumnCreatedAt,
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
			params.OrderBy = datastore.ShortlinkOrderColumnTitle
		case "is_active":
			params.OrderBy = datastore.ShortlinkOrderColumnIsActive
		case "created_at":
			params.OrderBy = datastore.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = datastore.ShortlinkOrderColumnUpdatedAt
		case "expired_at":
			params.OrderBy = datastore.ShortlinkOrderColumnExpiredAt
		default:
			params.OrderBy = datastore.ShortlinkOrderColumnCreatedAt // Default order
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
	response := make([]shortlink.LinkResponse, len(userLinks))
	for i, link := range userLinks {
		response[i] = shortlink.LinkResponse{
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

	pagination := &helper.Pagination{
		Total:   len(response),
		Limit:   params.Limit,
		Offset:  params.Offset,
		HasMore: len(response) > int(params.Limit),
	}

	return response, pagination, nil
}

// ToggleLinkStatus toggles the active status of a short link
func (s *Service) ToggleLinkStatus(ctx context.Context, id uuid.UUID) error {
	err := s.repo.AdminToggleShortLinkStatus(ctx, id)
	if err != nil {
		s.log.Error("failed to toggle short link status", "error", err)
		return err
	}
	return nil
}

func (s *Service) GetStats(ctx context.Context) (*stats.StatsResponse, error) {
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
	return &stats.StatsResponse{
		TotalUsers:    users,
		TotalLinks:    links,
		ActiveLinks:   active,
		InactiveLinks: inactive,
	}, nil
}
