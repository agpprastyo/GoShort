package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type ShortLinkService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewShortLinkService(repo *repository.Queries) *ShortLinkService {
	return &ShortLinkService{repo: repo}
}

type ShortLinkServiceInterface interface {
	// CreateLink Create a new short link
	CreateLink(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error)

	// GetUserLinks Get all short links for a user with optional pagination
	GetUserLinks(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]dto.LinkResponse, error)

	// GetUserLink Get details of a specific link by ID (verifies ownership)
	GetUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)

	// UpdateUserLink Update an existing short link
	UpdateUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID, req dto.UpdateLinkRequest) (*dto.LinkResponse, error)

	// DeleteUserLink Delete a short link
	DeleteUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) error

	// ToggleUserLinkStatus Toggle active/inactive status
	ToggleUserLinkStatus(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error)

	// ShortCodeExists Check if a short code already exists
	ShortCodeExists(ctx context.Context, code string) (bool, error)
}

func (s *ShortLinkService) CreateLink(ctx context.Context, userID uuid.UUID, req dto.CreateLinkRequest) (*dto.LinkResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortLinkService) GetUserLinks(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]dto.LinkResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *ShortLinkService) GetUserLink(ctx context.Context, userID uuid.UUID, linkID uuid.UUID) (*dto.LinkResponse, error) {
	//TODO implement me
	panic("implement me")
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

	if req.ExpireAt.Valid {
		params.ExpiredAt = req.ExpireAt
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
	err := s.repo.DeleteUserShortLink(ctx, repository.DeleteUserShortLinkParams{
		ID:     linkID,
		UserID: userID,
	})

	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Warn("link not found for deletion",
				"user_id", userID.String(),
				"link_id", linkID.String())
			return ErrLinkNotFound
		}
		s.log.Error("failed to delete short link, error: %v", err)
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
