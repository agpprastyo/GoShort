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

func NewShortLinksStatsService(repo *repository.Queries, log *logger.Logger) *ShortLinksStatsService {
	return &ShortLinksStatsService{
		repo: repo,
		log:  log,
	}
}

type IShortLinksStatsService interface {
	GetShortLinksStats(ctx context.Context, userID uuid.UUID) ([]repository.LinkStat, error)
}
