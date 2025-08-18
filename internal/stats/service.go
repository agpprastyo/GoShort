package stats

import (
	"GoShort/internal/datastore"
	"GoShort/pkg/logger"
	"context"

	"github.com/google/uuid"
)

type ShortLinksStatsService struct {
	repo *datastore.Queries
	log  *logger.Logger
}

func NewShortLinksStatsService(repo *datastore.Queries, log *logger.Logger) IShortLinksStatsService {
	return &ShortLinksStatsService{
		repo: repo,
		log:  log,
	}
}

type IShortLinksStatsService interface {
	GetShortLinksStats(ctx context.Context, userID uuid.UUID) ([]datastore.LinkStat, error)
}

func (s *ShortLinksStatsService) GetShortLinksStats(ctx context.Context, userID uuid.UUID) ([]datastore.LinkStat, error) {
	return []datastore.LinkStat{}, nil
}
