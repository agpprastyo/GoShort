package handler

import (
	"GoShort/internal/dto"
	"errors"

	"GoShort/internal/service"
	"GoShort/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"time"
)

type ShortLinkHandler struct {
	svr service.IShortLinkService
	log *logger.Logger
}

func NewShortLinkHandler(service service.IShortLinkService, log *logger.Logger) *ShortLinkHandler {
	return &ShortLinkHandler{
		svr: service,
		log: log,
	}
}

// CreateShortLink handles creation of a new short link
func (h *ShortLinkHandler) CreateShortLink(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Parse request body into DTO
	var req dto.CreateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.OriginalURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Original URL is required",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Let the service layer handle the creation using the request DTO
	link, err := h.svr.CreateLinkFromDTO(c.Context(), userUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(link)
}

// GetUserLinks retrieves all user's short links
func (h *ShortLinkHandler) GetUserLinks(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Extract query parameters into DTO
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

	// Pass the DTO directly to service
	links, pagination, err := h.svr.GetUserLinks(ctx, userUUID, req)
	if err != nil {
		h.log.Error("failed to get user links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"links":      links,
		"pagination": pagination,
	})
}

// GetUserLinkByID retrieves a specific short link by ID
func (h *ShortLinkHandler) GetUserLinkByID(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	link, err := h.svr.GetUserLinkByID(ctx, userUUID, linkUUID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Short link not found",
			})
		case errors.Is(err, service.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not authorized to access this link",
			})
		default:
			h.log.Error("failed to get user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve short link",
			})
		}
	}

	return c.JSON(link)
}

// UpdateLink updates an existing short link
func (h *ShortLinkHandler) UpdateLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	var req dto.UpdateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	link, err := h.svr.UpdateUserLink(ctx, userUUID, linkUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(link)
}

// DeleteLink deletes a short link
func (h *ShortLinkHandler) DeleteLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	h.log.Println("userID:", userID, "linkID:", linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	if err := h.svr.DeleteUserLink(ctx, userUUID, linkUUID); err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Short link not found",
			})
		case errors.Is(err, service.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "You are not authorized to delete this link",
			})
		default:
			h.log.Error("failed to delete user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete short link",
			})
		}

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ToggleLinkStatus activates or deactivates a short link
func (h *ShortLinkHandler) ToggleLinkStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid link ID",
		})
	}

	link, err := h.svr.ToggleUserLinkStatus(ctx, userUUID, linkUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(link)
}
