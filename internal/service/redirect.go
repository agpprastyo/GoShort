// internal/service/redirect.go
package service

import (
	"GoShort/internal/repository"
	"GoShort/pkg/logger"
	"context"
)

type RedirectService struct {
	repo *repository.Queries
	log  *logger.Logger
}

func NewRedirectService(repo *repository.Queries, log *logger.Logger) *RedirectService {
	return &RedirectService{
		repo: repo,
		log:  log,
	}
}

func (s *RedirectService) GetOriginalURL(ctx context.Context, code string) (string, bool, error) {
	link, err := s.repo.GetShortLinkByCode(ctx, code)
	if err != nil {
		return "", false, err
	}

	//// Increment clicks asynchronously
	//go func() {
	//	if err := s.repo.IncrementLinkClicks(context.Background(), link.ID); err != nil {
	//		s.log.Error("failed to increment link clicks", "error", err)
	//	}
	//}()

	return link.OriginalUrl, link.IsActive, nil
}
