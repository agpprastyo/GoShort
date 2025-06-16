package dto

import (
	"GoShort/internal/repository"
	"github.com/google/uuid"
	"time"
)

type GetLinksRequest struct {
	Limit     *int64                           `json:"limit,omitempty"`
	Offset    *int64                           `json:"offset,omitempty"`
	Search    *string                          `json:"search,omitempty"`
	Order     *repository.ShortlinkOrderColumn `json:"order,omitempty"`
	Ascending *bool                            `json:"ascending,omitempty"`
	StartDate *time.Time                       `json:"start_date,omitempty"`
	EndDate   *time.Time                       `json:"end_date,omitempty"`
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
