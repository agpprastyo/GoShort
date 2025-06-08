package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"GoShort/internal/service"
	"GoShort/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"strconv"
	"time"
)

type ShortLinkHandler struct {
	service *service.ShortLinkService
	log     *logger.Logger
}

func NewShortLinkHandler(service *service.ShortLinkService, log *logger.Logger) *ShortLinkHandler {
	return &ShortLinkHandler{
		service: service,
		log:     log,
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
	link, err := h.service.CreateLinkFromDTO(c.Context(), userUUID, req)
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
	var req dto.GetUserLinksRequest

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

	// Convert DTO to repository params
	params := repository.ListUserShortLinksParams{
		UserID:     userUUID,
		Limit:      10,                                       // Default limit
		Offset:     0,                                        // Default offset
		SearchText: "",                                       // Default search
		OrderBy:    repository.ShortlinkOrderColumnCreatedAt, // Default order
		Ascending:  false,                                    // Default ascending
		StartDate:  pgtype.Timestamptz{Valid: false},
		EndDate:    pgtype.Timestamptz{Valid: false},
	}

	// Apply values from DTO if provided
	if req.Limit != nil {
		params.Limit = *req.Limit
	}

	if req.Offset != nil {
		params.Offset = *req.Offset
	}

	if req.Search != nil {
		params.SearchText = *req.Search
	}

	if req.Ascending != nil {
		params.Ascending = *req.Ascending
	}

	// Convert string order to enum
	if req.Order != nil {
		switch *req.Order {
		case "created_at":
			params.OrderBy = repository.ShortlinkOrderColumnCreatedAt
		case "updated_at":
			params.OrderBy = repository.ShortlinkOrderColumnUpdatedAt
		case "title":
			params.OrderBy = repository.ShortlinkOrderColumnTitle
		case "is_active":
			params.OrderBy = repository.ShortlinkOrderColumnIsActive
		}
	}

	// Apply date filters if provided
	if req.StartDate != nil && req.StartDate.Valid {
		params.StartDate = *req.StartDate
	}

	if req.EndDate != nil && req.EndDate.Valid {
		params.EndDate = *req.EndDate
	}

	links, err := h.service.GetUserLinks(ctx, params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if len(links) == 0 {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"links": links,
			"pagination": fiber.Map{
				"total":    0,
				"limit":    params.Limit,
				"offset":   params.Offset,
				"has_more": false,
			},
		})
	}

	// For successful case with data
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"links": links,
		"pagination": fiber.Map{
			"total":    len(links), // Ideally this would be total count from database
			"limit":    params.Limit,
			"offset":   params.Offset,
			"has_more": int64(len(links)) == params.Limit, // Basic estimation if there might be more
		},
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

	link, err := h.service.GetUserLinkByID(ctx, userUUID, linkUUID)
	if err != nil {
		switch err {
		case service.ErrLinkNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Short link not found",
			})
		case service.ErrUnauthorized:
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

	link, err := h.service.UpdateUserLink(ctx, userUUID, linkUUID, req)
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

	if err := h.service.DeleteUserLink(ctx, userUUID, linkUUID); err != nil {
		switch err {
		case service.ErrLinkNotFound:
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Short link not found",
			})
		case service.ErrUnauthorized:
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

	link, err := h.service.ToggleUserLinkStatus(ctx, userUUID, linkUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(link)
}
