package service

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"time"
)

type IRedirectService interface {
	GetOriginalURL(ctx context.Context, code string) (originalUrl string, linkID uuid.UUID, isActive bool, err error)
	RecordLinkStat(ctx context.Context, linkID uuid.UUID, info dto.CreateLinkStatRequest) error
}

type RedirectService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewRedirectService(repo *repository.Queries, log *logger.Logger) IRedirectService {
	return &RedirectService{
		repo: repo,
		log:  log,
	}
}

func (s *RedirectService) GetOriginalURL(ctx context.Context, code string) (originalUrl string, linkID uuid.UUID, isActive bool, err error) {
	link, err := s.repo.GetShortLinkByCode(ctx, code)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			s.log.Warn("link not found", "code", code)
			return "", uuid.Nil, false, ErrLinkNotFound
		default:
			s.log.Error("failed to retrieve link by code", "code", code, "error", err)
			return "", uuid.Nil, false, err
		}
	}

	// Check if the link is active
	if !link.IsActive {
		s.log.Warn("attempted to access inactive link", "code", code, "link_id", link.ID)
		return "", uuid.Nil, false, ErrLinkNotActive
	}

	// Check if the link has expired
	if link.ExpiredAt.Valid {
		now := time.Now()
		if link.ExpiredAt.Time.Before(now) {
			s.log.Warn("attempted to access expired link", "code", code, "link_id", link.ID)
			return "", uuid.Nil, false, ErrLinkExpired
		}
	}

	// Check if the link has reached its click limit
	if link.ClickLimit != nil && *link.ClickLimit <= 0 {
		s.log.Warn("attempted to access link with no remaining clicks ", "code: ", code, " link_id: ", link.ID)
		return "", uuid.Nil, false, ErrClickLimitExceeded
	}

	// Log the access
	s.log.Info("redirecting to original URL", "code", code, "link_id", link.ID, "original_url", link.OriginalUrl)

	// Increment clicks asynchronously
	go func() {
		//if err := s.repo.IncrementLinkClicks(context.Background(), link.ID); err != nil {
		//	s.log.Error("failed to increment link clicks", "error", err)
		//}

		if link.ClickLimit != nil && *link.ClickLimit > 0 {
			if _, err := s.repo.DecrementClickLimit(ctx, link.ID); err != nil {
				s.log.Error("failed to decrement link click limit", "error", err)
			}
		}

	}()

	return link.OriginalUrl, link.ID, link.IsActive, nil
}

// RecordLinkStat records a click in the link_stats table
func (s *RedirectService) RecordLinkStat(ctx context.Context, linkID uuid.UUID, info dto.CreateLinkStatRequest) error {

	recordUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Error("failed to generate UUID for link stat", "error: ", err, "link_id: ", linkID)
		return err
	}

	params := repository.CreateLinkStatParams{
		ClickTime:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
		ID:         recordUUID,
		LinkID:     linkID,
		IpAddress:  info.IpAddress,
		UserAgent:  info.UserAgent,
		Referrer:   info.Referrer,
		Country:    info.Country,
		DeviceType: info.DeviceType,
	}

	// Insert the record
	err = s.repo.CreateLinkStat(ctx, params)
	if err != nil {
		s.log.Error("failed to record link stat", "error", err, "link_id", linkID)
		return err
	}

	s.log.Infof("record link stat successfully", "link_id", linkID)

	return nil
}
