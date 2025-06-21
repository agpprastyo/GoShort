package service

import (
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"context"
	"github.com/google/uuid"
)

type ShortLinksStatsService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewShortLinksStatsService(repo *repository.Queries, log *logger.Logger) IShortLinksStatsService {
	return &ShortLinksStatsService{
		repo: repo,
		log:  log,
	}
}

type IShortLinksStatsService interface {
	GetShortLinksStats(ctx context.Context, userID uuid.UUID) ([]repository.LinkStat, error)
}

func (s *ShortLinksStatsService) GetShortLinksStats(ctx context.Context, userID uuid.UUID) ([]repository.LinkStat, error) {
	return []repository.LinkStat{}, nil
}
