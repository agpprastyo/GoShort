package dto

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type GetUserLinksRequest struct {
	Limit     *int64              `json:"limit,omitempty"`
	Offset    *int64              `json:"offset,omitempty"`
	Search    *string             `json:"search,omitempty"`
	Order     *string             `json:"order,omitempty"`
	Ascending *bool               `json:"ascending,omitempty"`
	StartDate *pgtype.Timestamptz `json:"start_date,omitempty"`
	EndDate   *pgtype.Timestamptz `json:"end_date,omitempty"`
}

type CreateLinkRequest struct {
	OriginalURL string           `json:"original_url"`
	ShortCode   *string          `json:"short_code,omitempty"`
	Title       *string          `json:"title,omitempty"`
	ClickLimit  *int32           `json:"click_limit,omitempty"`
	ExpireAt    pgtype.Timestamp `json:"expire_at,omitempty"`
}

type UpdateLinkRequest struct {
	OriginalURL *string           `json:"original_url,omitempty"`
	ShortCode   *string           `json:"short_code,omitempty"`
	Title       *string           `json:"title,omitempty"`
	IsActive    *bool             `json:"is_active,omitempty"`
	ClickLimit  *int32            `json:"click_limit,omitempty"`
	ExpireAt    *pgtype.Timestamp `json:"expire_at,omitempty"`
}

type LinkResponse struct {
	ID          uuid.UUID        `json:"id"`
	OriginalURL string           `json:"original_url"`
	ShortCode   string           `json:"short_code"`
	Title       *string          `json:"title,omitempty"`
	IsActive    bool             `json:"is_active"`
	ClickLimit  *int32           `json:"click_limit,omitempty"`
	ExpireAt    pgtype.Timestamp `json:"expire_at,omitempty"`
	CreatedAt   pgtype.Timestamp `json:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at"`
}
