package dto

import (
	"GoShort/internal/repository"
	"github.com/google/uuid"
	"time"
)

type GetLinksRequest struct {
	Limit     *int64                           `json:"limit,omitempty" query:"limit,omitempty"`
	Offset    *int64                           `json:"offset,omitempty" query:"offset,omitempty"`
	Search    *string                          `json:"search,omitempty" query:"search,omitempty"`
	Order     *repository.ShortlinkOrderColumn `json:"order,omitempty" query:"order,omitempty"`
	Ascending *bool                            `json:"ascending,omitempty" query:"ascending,omitempty"`
	StartDate *time.Time                       `json:"start_date,omitempty" query:"start_date,omitempty"`
	EndDate   *time.Time                       `json:"end_date,omitempty" query:"end_date,omitempty"`
}

type CreateLinkRequest struct {
	OriginalURL string     `json:"original_url"`
	ShortCode   *string    `json:"short_code,omitempty"`
	Title       *string    `json:"title,omitempty"`
	ClickLimit  *int32     `json:"click_limit,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
}

type UpdateLinkRequest struct {
	OriginalURL *string    `json:"original_url,omitempty"`
	ShortCode   *string    `json:"short_code,omitempty"`
	Title       *string    `json:"title,omitempty"`
	IsActive    *bool      `json:"is_active,omitempty"`
	ClickLimit  *int32     `json:"click_limit,omitempty"`
	ExpireAt    *time.Time `json:"expire_at,omitempty"`
}

type LinkResponse struct {
	ID          uuid.UUID `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	Title       *string   `json:"title,omitempty"`
	IsActive    bool      `json:"is_active"`
	ClickLimit  *int32    `json:"click_limit,omitempty"`
	ExpireAt    time.Time `json:"expire_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LinkResponseWithTotalClicks struct {
	ID          uuid.UUID `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortCode   string    `json:"short_code"`
	Title       *string   `json:"title,omitempty"`
	IsActive    bool      `json:"is_active"`
	ClickLimit  *int32    `json:"click_limit,omitempty"`
	ExpireAt    time.Time `json:"expire_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	TotalClicks int32     `json:"total_clicks"`
}

type BulkCreateLinkRequest struct {
	Links []CreateLinkRequest `json:"links" validate:"required,dive" query:"links"`
}

type BulkCreateLinkResponse struct {
	Created      []LinkResponse        `json:"created"`
	Failed       []BulkCreateLinkError `json:"failed"`
	Total        int                   `json:"total"`
	FailedCount  int                   `json:"failed_count"`
	CreatedCount int                   `json:"created_count"`
}

type BulkCreateLinkError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

type BulkDeleteLinkRequest struct {
	IDs []uuid.UUID `json:"ids" validate:"required,dive,uuid" query:"ids"`
}

type BulkDeleteLinkResponse struct {
	Deleted      []uuid.UUID           `json:"deleted"`
	Failed       []BulkDeleteLinkError `json:"failed"`
	Total        int                   `json:"total"`
	FailedCount  int                   `json:"failed_count"`
	DeletedCount int                   `json:"deleted_count"`
}

type BulkDeleteLinkError struct {
	Index int    `json:"index"`
	Error string `json:"error" json:"error"`
}

type LinksUserStatsResponse struct {
}
