package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/helper"
	"GoShort/pkg/logger"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type IShortLinkService interface {
	GetUserLinkByID(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)
	CreateLinkFromDTO(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error)
	GetUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error)
	GetUserLinksWithCount(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponseWithTotalClicks, *dto.Pagination, error)
	UpdateUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error)
	DeleteUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error
	ToggleUserLinkStatus(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)
	ShortCodeExists(ctx context.Context, code string) (bool, error)
	GetUserLinkByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*dto.LinkResponse, error)
	CreateBulkShortLinks(ctx context.Context, userID uuid.UUID, links dto.BulkCreateLinkRequest) (dto.BulkCreateLinkResponse, error)
	DeleteBulkShortLinks(ctx context.Context, userID uuid.UUID, request dto.BulkDeleteLinkRequest) (dto.BulkDeleteLinkResponse, error)
	DeleteAllLinks(ctx context.Context, userID uuid.UUID) error
}

type ShortLinkService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewShortLinkService(repo *repository.Queries, log *logger.Logger) IShortLinkService {
	return &ShortLinkService{repo: repo, log: log}
}

// DeleteAllLinks deletes all short links for a user
func (s *ShortLinkService) DeleteAllLinks(ctx context.Context, userID uuid.UUID) error {
	// Call repository to delete all links for the user
	err := s.repo.DeleteUserShortLink(ctx, userID)
	if err != nil {
		s.log.Error("failed to delete all short links for user", "user_id", userID.String(), "error", err)
		return err
	}
	return nil
}

// DeleteBulkShortLinks deletes multiple short links for a user
func (s *ShortLinkService) DeleteBulkShortLinks(ctx context.Context, userID uuid.UUID, request dto.BulkDeleteLinkRequest) (dto.BulkDeleteLinkResponse, error) {
	if len(request.IDs) == 0 {
		return dto.BulkDeleteLinkResponse{}, errors.New("no link IDs provided")
	}

	var deleted []uuid.UUID
	var failed []dto.BulkDeleteLinkError

	for i, linkID := range request.IDs {
		err := s.DeleteUserLink(ctx, userID, linkID)
		if err != nil {
			s.log.Error("failed to delete link", "link_id", linkID.String(), "error", err)
			failed = append(failed, dto.BulkDeleteLinkError{
				Index: i,
				Error: err.Error(),
			})
			continue
		}
		deleted = append(deleted, linkID)
	}

	return dto.BulkDeleteLinkResponse{
		Deleted:      deleted,
		Failed:       failed,
		Total:        len(request.IDs),
		FailedCount:  len(failed),
		DeletedCount: len(deleted),
	}, nil
}

func (s *ShortLinkService) CreateBulkShortLinks(ctx context.Context, userID uuid.UUID, req dto.BulkCreateLinkRequest) (dto.BulkCreateLinkResponse, error) {
	if len(req.Links) == 0 {
		return dto.BulkCreateLinkResponse{}, errors.New("no links provided")
	}

	var created []dto.LinkResponse
	var failed []dto.BulkCreateLinkError

	for i, link := range req.Links {
		createdLink, err := s.CreateLinkFromDTO(ctx, userID, link)
		if err != nil {
			s.log.Error("failed to create link from DTO", "error", err)
			failed = append(failed, dto.BulkCreateLinkError{
				Index: i,
				Error: err.Error(),
			})
			continue
		}
		created = append(created, *createdLink)
	}

	return dto.BulkCreateLinkResponse{
		Created:      created,
		Failed:       failed,
		Total:        len(req.Links),
		FailedCount:  len(failed),
		CreatedCount: len(created),
	}, nil
}

// GetUserLinkByShortCode retrieves a short link by its short code for a specific user
func (s *ShortLinkService) GetUserLinkByShortCode(ctx context.Context, userID uuid.UUID, shortCode string) (*dto.LinkResponse, error) {

	// Call repository to get the short link by code
	link, err := s.repo.GetShortLinkByCode(ctx, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			s.log.Error("short link not found", "short_code", shortCode, "error", err)
			return nil, ErrLinkNotFound
		default:
			s.log.Error("unexpected error while getting short link by code", "short_code", shortCode, "error", err)
			return nil, err
		}
	}

	// Check if the link belongs to the user
	if link.UserID != userID {
		s.log.Warn("unauthorized access to link by short code",
			"user_id", userID.String(),
			"short_code", shortCode)
		return nil, ErrUnauthorized
	}

	// Convert to response DTO
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

// GetUserLinkByID retrieves a short link by its ID for a specific user
func (s *ShortLinkService) GetUserLinkByID(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	// Call repository to get the short link
	link, err := s.repo.GetShortLink(ctx, linkID)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			s.log.Error("short link not found", "error", err)
			return nil, ErrLinkNotFound
		default:
			s.log.Error("unexpected error while getting short link", "error", err)
			return nil, err
		}
	}

	// Check if the link belongs to the user
	if link.UserID != userID {
		s.log.Warn("unauthorized access to link",
			"user_id", userID.String(),
			"link_id", linkID.String())
		return nil, ErrUnauthorized
	}

	// Convert to response DTO
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

func (s *ShortLinkService) CreateLinkFromDTO(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error) {

	linkID, err := uuid.NewV7()
	if err != nil {
		s.log.Error("failed to generate new UUID for link", "error", err)
		return nil, err
	}

	if req.ShortCode != nil {
		exists, err := s.ShortCodeExists(ctx, *req.ShortCode)
		if err != nil {
			s.log.Error("failed to check short code exists: %v", err)
			return nil, err
		}
		if exists {
			return nil, ErrShortCodeExists
		}
	} else {
		// Generate a random short code if not provided from helper function
		shortCode, err := helper.GenerateShortCode(7)
		if err != nil {
			s.log.Error("failed to generate short code", "error", err)
			return nil, err
		}
		req.ShortCode = &shortCode
	}

	if req.Title == nil {
		defaultTitle := ""
		req.Title = &defaultTitle
	}

	if req.ClickLimit == nil {
		defaultLimit := int32(1000)
		req.ClickLimit = &defaultLimit
	}

	if req.ExpireAt == nil {
		defaultExpire := time.Now().Add(30 * 24 * time.Hour)
		req.ExpireAt = &defaultExpire
	}

	params := repository.CreateShortLinkParams{
		ID:          linkID,
		UserID:      userID,
		OriginalUrl: req.OriginalURL,
		ShortCode:   *req.ShortCode,
		Title:       req.Title,
		IsActive:    true,
		ClickLimit:  req.ClickLimit,
		ExpiredAt: pgtype.Timestamp{
			Time: *req.ExpireAt,
		},
	}

	// Create the short link in the repository
	createdLink, err := s.repo.CreateShortLink(ctx, params)
	if err != nil {
		s.log.Error("failed to create short link", "error", err)
		return nil, err
	}

	// Convert to response DTO
	response := &dto.LinkResponse{
		ID:          createdLink.ID,
		OriginalURL: createdLink.OriginalUrl,
		ShortCode:   createdLink.ShortCode,
		Title:       createdLink.Title,
		IsActive:    createdLink.IsActive,
		ClickLimit:  createdLink.ClickLimit,
		ExpireAt:    createdLink.ExpiredAt.Time,
		CreatedAt:   createdLink.CreatedAt.Time,
		UpdatedAt:   createdLink.UpdatedAt.Time,
	}

	return response, nil

}

// GetUserLinks retrieves a user's short links with filtering and pagination
func (s *ShortLinkService) GetUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error) {
	// Convert DTO to repository params with defaults
	params := repository.ListUserShortLinksParams{
		UserID:     userID,
		Limit:      10, // Default limit
		Offset:     0,  // Default offset
		SearchText: "",
		OrderBy:    repository.ShortlinkOrderColumnCreatedAt,
		Ascending:  false,
		StartDate:  pgtype.Timestamptz{}, // Default empty timestamp
		EndDate:    pgtype.Timestamptz{}, // Default empty timestamp
	}

	// Apply values from DTO if provided
	if req.Limit != nil {
		params.Limit = *req.Limit
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Order != nil {
		switch *req.Order {
		case "created_at":
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = repository.ShortlinkOrderColumnUpdatedAt
		case "is_active":
			params.OrderBy = repository.ShortlinkOrderColumnIsActive
		case "title":
			params.OrderBy = repository.ShortlinkOrderColumnTitle
		}
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	}

	if req.StartDate != nil {
		startTime := pgtype.Timestamptz{Time: *req.StartDate}
		params.StartDate = startTime
	}

	if req.EndDate != nil {
		endTime := pgtype.Timestamptz{Time: *req.EndDate}
		params.EndDate = endTime
	}

	// Call repository
	links, err := s.repo.ListUserShortLinks(ctx, params)
	if err != nil {
		return nil, nil, err
	}

	// Convert repository results to DTOs
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

	// Use the global helper for pagination
	pagination := helper.BuildPaginationInfo(len(links), len(links), params.Limit, params.Offset)

	return response, &pagination, nil
}

// GetUserLinksWithCount retrieves a user's short links with click counts and pagination
func (s *ShortLinkService) GetUserLinksWithCount(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponseWithTotalClicks, *dto.Pagination, error) {

	params := repository.ListUserShortLinksWithCountClickParams{
		UserID:     userID,
		Limit:      10,
		Offset:     0,
		SearchText: "",
		OrderBy:    repository.ShortlinkOrderColumnCreatedAt,
		Ascending:  false,
		StartDate:  pgtype.Timestamptz{},
		EndDate:    pgtype.Timestamptz{},
	}

	countParams := repository.CountUserShortLinksParams{
		UserID:     userID,
		SearchText: params.SearchText,
		StartDate:  params.StartDate,
		EndDate:    params.EndDate,
	}

	totalCount, err := s.repo.CountUserShortLinks(ctx, countParams)
	if err != nil {
		s.log.Error("failed to count user short links", "error", err)
		return nil, nil, err
	}

	if req.Limit != nil {
		params.Limit = *req.Limit
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Order != nil {
		switch *req.Order {
		case "created_at":
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = repository.ShortlinkOrderColumnUpdatedAt
		case "is_active":
			params.OrderBy = repository.ShortlinkOrderColumnIsActive
		case "title":
			params.OrderBy = repository.ShortlinkOrderColumnTitle

		default:
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt // Default order by created_at
		}
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	}

	if req.StartDate != nil {
		startTime := pgtype.Timestamptz{Time: *req.StartDate}
		params.StartDate = startTime
	}

	if req.EndDate != nil {
		endTime := pgtype.Timestamptz{Time: *req.EndDate}
		params.EndDate = endTime
	}

	// Call repository
	results, err := s.repo.ListUserShortLinksWithCountClick(ctx, params)
	s.log.Infof("result : %v", results)
	if err != nil {
		return nil, nil, err
	}

	response := make([]dto.LinkResponseWithTotalClicks, len(results))
	for i, link := range results {
		response[i] = dto.LinkResponseWithTotalClicks{
			ID:          link.ID,
			OriginalURL: link.OriginalUrl,
			ShortCode:   link.ShortCode,
			Title:       link.Title,
			IsActive:    link.IsActive,
			ClickLimit:  link.ClickLimit,
			ExpireAt:    link.ExpiredAt.Time,
			CreatedAt:   link.CreatedAt.Time,
			UpdatedAt:   link.UpdatedAt.Time,
			TotalClicks: int32(link.TotalClicks),
		}
	}
	// Use the global helper with total count from count query
	pagination := helper.BuildPaginationInfo(int(totalCount), len(response), params.Limit, params.Offset)
	s.log.Println("Pagination info:", pagination)

	return response, &pagination, nil

}

func (s *ShortLinkService) UpdateUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error) {
	// First verify the link belongs to the user
	link, err := s.repo.GetShortLink(ctx, linkID)
	if err != nil {
		s.log.Error("failed to get short link, error: %v", err)
		return nil, ErrLinkNotFound
	}

	// Check ownership
	if link.UserID != userID {
		s.log.Warn("unauthorized update attempt",
			"user_id", userID.String(),
			"link_id", linkID.String())
		return nil, ErrUnauthorized
	}

	// Check if the short code is changed and if it already exists
	if req.ShortCode != nil && *req.ShortCode != link.ShortCode {
		exists, err := s.ShortCodeExists(ctx, *req.ShortCode)
		if err != nil {
			s.log.Error("failed to check short code exists: %v", err)
			return nil, err
		}
		if exists {
			return nil, ErrShortCodeExists
		}
	}

	// Prepare update parameters
	params := repository.UpdateShortLinkParams{
		ID: linkID,
	}

	if req.OriginalURL != nil {
		params.OriginalUrl = *req.OriginalURL
	} else {
		params.OriginalUrl = link.OriginalUrl // Keep existing if not provided
	}

	if req.ShortCode != nil {
		params.ShortCode = *req.ShortCode
	} else {
		params.ShortCode = link.ShortCode // Keep existing if not provided
	}

	if req.Title != nil {
		params.Title = req.Title
	} else {
		params.Title = link.Title // Keep existing if not provided
	}

	if req.ClickLimit != nil {
		params.ClickLimit = req.ClickLimit
	} else {
		params.ClickLimit = link.ClickLimit // Keep existing if not provided
	}

	if req.ExpireAt != nil {
		expiredTime := pgtype.Timestamp{Time: *req.ExpireAt}
		params.ExpiredAt = expiredTime
	} else {
		params.ExpiredAt = link.ExpiredAt // Keep existing if not provided
	}
	// If IsActive is not provided, keep the existing value
	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	} else {
		params.IsActive = link.IsActive // Keep existing if not provided
	}

	// Update the link
	updatedLink, err := s.repo.UpdateShortLink(ctx, params)
	if err != nil {
		s.log.Error("failed to update short link: %v", err)
		return nil, err
	}

	// Convert to response DTO
	response := &dto.LinkResponse{
		ID:          updatedLink.ID,
		OriginalURL: updatedLink.OriginalUrl,
		ShortCode:   updatedLink.ShortCode,
		Title:       updatedLink.Title,
		IsActive:    updatedLink.IsActive,
		ClickLimit:  updatedLink.ClickLimit,
		ExpireAt:    updatedLink.ExpiredAt.Time,
		CreatedAt:   updatedLink.CreatedAt.Time,
		UpdatedAt:   updatedLink.UpdatedAt.Time,
	}

	return response, nil
}

func (s *ShortLinkService) DeleteUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error {
	// First verify the link belongs to the user
	link, err := s.repo.GetShortLink(ctx, linkID)
	if err != nil {
		s.log.Error("failed to get short link, error: %v", err)
		return ErrLinkNotFound
	}

	// Check ownership
	if link.UserID != userID {
		s.log.Warn("unauthorized delete attempt",
			"user_id", userID.String(),
			"link_id", linkID.String())
		return ErrUnauthorized
	}

	// Delete the link
	err = s.repo.DeleteUserShortLink(ctx, linkID)
	if err != nil {
		s.log.Error("failed to delete short link: %v", err)
		return err
	}
	return nil
}

func (s *ShortLinkService) ToggleUserLinkStatus(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	// First verify the link belongs to the user
	link, err := s.repo.GetShortLink(ctx, linkID)
	if err != nil {
		s.log.Error("failed to get short link, error: %v", err)
		return nil, ErrLinkNotFound
	}

	// Check ownership
	if link.UserID != userID {
		s.log.Warn("unauthorized toggle attempt",
			"user_id", userID.String(),
			"link_id", linkID.String())
		return nil, ErrUnauthorized
	}

	// Toggle the status
	updatedLink, err := s.repo.ToggleShortLinkStatus(ctx, linkID)
	if err != nil {
		s.log.Error("failed to toggle link status: %v", err)
		return nil, err
	}

	// Convert to response DTO
	response := &dto.LinkResponse{
		ID:          updatedLink.ID,
		OriginalURL: updatedLink.OriginalUrl,
		ShortCode:   updatedLink.ShortCode,
		Title:       updatedLink.Title,
		IsActive:    updatedLink.IsActive,
		ClickLimit:  updatedLink.ClickLimit,
		ExpireAt:    updatedLink.ExpiredAt.Time,
		CreatedAt:   updatedLink.CreatedAt.Time,
		UpdatedAt:   updatedLink.UpdatedAt.Time,
	}

	return response, nil
}

func (s *ShortLinkService) ShortCodeExists(ctx context.Context, code string) (bool, error) {
	if code == "" {
		return false, nil
	}

	exists, err := s.repo.CheckShortCodeExists(ctx, code)
	if err != nil {
		s.log.Error("failed to check if short code exists: %v", err)
		return false, err
	}

	return exists, nil
}
