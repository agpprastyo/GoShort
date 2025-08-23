package stats

import (
	"GoShort/pkg/logger"

	"github.com/gofiber/fiber/v2"
)

type ShortLinksStatsHandler struct {
	svr IShortLinksStatsService
	log *logger.Logger
}

func NewShortLinksStatsHandler(svr IShortLinksStatsService, log *logger.Logger) *ShortLinksStatsHandler {
	return &ShortLinksStatsHandler{
		svr: svr,
		log: log,
	}
}

// GetUserStats retrieves statistics for the authenticated user
func (h *ShortLinksStatsHandler) GetUserStats(c *fiber.Ctx) error {
	//userID, err := h.svr.GetAuthenticatedUserID(c)
	//if err != nil {
	//	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	//}
	//
	//stats, err := h.svr.GetShortLinksStats(c.Context(), userID)
	//if err != nil {
	//	h.log.Error("Failed to get user stats", "error", err)
	//	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve statistics"})
	//}

	return c.JSON(nil)
}
