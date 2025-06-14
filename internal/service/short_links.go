package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/helper"
	"GoShort/pkg/logger"
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type IShortLinkService interface {
	GetUserLinkByID(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)
	CreateLinkFromDTO(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error)
	GetUserLinks(ctx context.Context, userID uuid.UUID, req dto.GetLinksRequest) ([]dto.LinkResponse, *dto.Pagination, error)
	UpdateUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error)
	DeleteUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error
	ToggleUserLinkStatus(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)
	ShortCodeExists(ctx context.Context, code string) (bool, error)
}

type ShortLinkService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewShortLinkService(repo *repository.Queries, log *logger.Logger) IShortLinkService {
	return &ShortLinkService{repo: repo, log: log}
}

// GetUserLinkByID retrieves a short link by its ID for a specific user
func (s *ShortLinkService) GetUserLinkByID(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	// Call repository to get the short link
	link, err := s.repo.GetShortLink(ctx, linkID)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
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
		ExpireAt:    link.ExpiredAt,
		CreatedAt:   link.CreatedAt,
		UpdatedAt:   link.UpdatedAt,
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

	params := repository.CreateShortLinkParams{
		ID:          linkID,
		UserID:      userID,
		OriginalUrl: req.OriginalURL,
		ShortCode:   *req.ShortCode,
		Title:       req.Title,
		IsActive:    true,
		ClickLimit:  req.ClickLimit,
		ExpiredAt:   req.ExpireAt,
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
		ExpireAt:    createdLink.ExpiredAt,
		CreatedAt:   createdLink.CreatedAt,
		UpdatedAt:   createdLink.UpdatedAt,
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
		params.StartDate = *req.StartDate
	}

	if req.EndDate != nil {
		params.EndDate = *req.EndDate
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
			ExpireAt:    link.ExpiredAt,
			CreatedAt:   link.CreatedAt,
			UpdatedAt:   link.UpdatedAt,
		}
	}

	// Use the global helper for pagination
	pagination := helper.BuildPaginationInfo(len(links), params.Limit, params.Offset)

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

	if req.ExpireAt != nil && req.ExpireAt.Valid {
		params.ExpiredAt = *req.ExpireAt
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
		ExpireAt:    updatedLink.ExpiredAt,
		CreatedAt:   updatedLink.CreatedAt,
		UpdatedAt:   updatedLink.UpdatedAt,
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
		ExpireAt:    updatedLink.ExpiredAt,
		CreatedAt:   updatedLink.CreatedAt,
		UpdatedAt:   updatedLink.UpdatedAt,
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
