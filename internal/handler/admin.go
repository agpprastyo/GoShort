package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"GoShort/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"time"
)

type AdminHandler struct {
	adminService service.IAdminService
	log          *logger.Logger
}

func NewAdminHandler(adminService service.IAdminService, log *logger.Logger) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		log:          log,
	}
}

// ListAllLinks lists all short links in the system
func (h *AdminHandler) ListAllLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	var req dto.GetLinksRequest

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			req.Limit = &limit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
			req.Offset = &offset
		}
	}

	// Parse search
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	// Parse order
	if order := c.Query("order_by"); order != "" {
		req.Order = &order
	}

	// Parse ascending
	if ascStr := c.Query("ascending"); ascStr != "" {
		asc := ascStr == "true"
		req.Ascending = &asc
	}

	// Parse start date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate := pgtype.Timestamptz{Time: startTime, Valid: true}
			req.StartDate = &startDate
		}
	}

	// Parse end date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate := pgtype.Timestamptz{Time: endTime, Valid: true}
			req.EndDate = &endDate
		}
	}

	links, pagination, err := h.adminService.ListAllLinks(ctx, req)
	if err != nil {
		h.log.Error("failed to list all links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"links":      links,
		"pagination": pagination,
	})

}

// GetLink retrieves a specific short link by ID
func (h *AdminHandler) GetLink(c *fiber.Ctx) error {
	ctx := c.Context()

	linkID := c.Params("id")

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		h.log.Error("invalid link ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	link, err := h.adminService.GetLinkByID(ctx, linkUUID)
	if err != nil {
		h.log.Error("failed to get link", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"link": link,
	})

}

// ListUserLinks lists all short links for a specific user
func (h *AdminHandler) ListUserLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	userID := c.Params("userId")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Error("invalid user ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req dto.GetLinksRequest

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			req.Limit = &limit
		}
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.ParseInt(offsetStr, 10, 64); err == nil {
			req.Offset = &offset
		}
	}

	// Parse search
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	// Parse order
	if order := c.Query("order_by"); order != "" {
		req.Order = &order
	}

	// Parse ascending
	if ascStr := c.Query("ascending"); ascStr != "" {
		asc := ascStr == "true"
		req.Ascending = &asc
	}

	// Parse start date
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startTime, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate := pgtype.Timestamptz{Time: startTime, Valid: true}
			req.StartDate = &startDate
		}
	}

	// Parse end date
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endTime, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate := pgtype.Timestamptz{Time: endTime, Valid: true}
			req.EndDate = &endDate
		}
	}

	userLinks, pagination, err := h.adminService.ListUserLinks(ctx, userUUID, req)
	if err != nil {
		h.log.Error("failed to list user links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})

	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"links":      userLinks,
		"pagination": pagination,
	})
}

// ToggleLinkStatus toggles the active status of a short link
func (h *AdminHandler) ToggleLinkStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	linkID := c.Params("id")

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		h.log.Error("invalid link ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	err = h.adminService.ToggleLinkStatus(ctx, linkUUID)
	if err != nil {
		h.log.Error("failed to toggle link status", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(
		fiber.Map{
			"message": "Link status toggled successfully",
		})

}
