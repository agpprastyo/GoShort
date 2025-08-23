package redirect

import (
	"GoShort/internal/commons"
	"GoShort/internal/datastore"
	"errors"

	"GoShort/internal/stats"
	"GoShort/pkg/logger"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"time"
)

type IService interface {
	GetOriginalURL(ctx context.Context, code string) (originalUrl string, linkID uuid.UUID, isActive bool, err error)
	RecordLinkStat(ctx context.Context, linkID uuid.UUID, info stats.CreateLinkStatRequest) error
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

func (s *Service) GetOriginalURL(ctx context.Context, code string) (originalUrl string, linkID uuid.UUID, isActive bool, err error) {
	link, err := s.repo.GetShortLinkByCode(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			s.log.Warn("link not found", "code", code)
			return "", uuid.Nil, false, commons.ErrLinkNotFound
		default:
			s.log.Error("failed to retrieve link by code", "code", code, "error", err)
			return "", uuid.Nil, false, err
		}
	}

	// Check if the link is active
	if !link.IsActive {
		s.log.Warn("attempted to access inactive link", "code", code, "link_id", link.ID)
		return "", uuid.Nil, false, commons.ErrLinkNotActive
	}

	// Check if the link has expired
	if link.ExpiredAt.Valid {
		now := time.Now()
		if link.ExpiredAt.Time.Before(now) {
			s.log.Warn("attempted to access expired link", "code", code, "link_id", link.ID)
			return "", uuid.Nil, false, commons.ErrLinkExpired
		}
	}

	// Check if the link has reached its click limit
	if link.ClickLimit != nil && *link.ClickLimit <= 0 {
		s.log.Warn("attempted to access link with no remaining clicks ", "code: ", code, " link_id: ", link.ID)
		return "", uuid.Nil, false, commons.ErrClickLimitExceeded
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
func (s *Service) RecordLinkStat(ctx context.Context, linkID uuid.UUID, info stats.CreateLinkStatRequest) error {

	recordUUID, err := uuid.NewV7()
	if err != nil {
		s.log.Error("failed to generate UUID for link stat", "error: ", err, "link_id: ", linkID)
		return err
	}

	params := datastore.CreateLinkStatParams{
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

	s.log.Info("record link stat successfully", "link_id", linkID)

	return nil
}
