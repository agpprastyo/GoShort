package shortlink

import (
	"GoShort/internal/commons"
	"GoShort/internal/datastore"

	"errors"

	"GoShort/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	svr IService
	log *logger.Logger
}

func NewHandler(service IService, log *logger.Logger) *Handler {
	return &Handler{
		svr: service,
		log: log,
	}
}

// DeleteAllLinks deletes all short links for the authenticated user
// @Godoc DeleteAllLinks
// @Summary Delete all short links for the authenticated user
// @Description Delete all short links created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Success 204 "All short links deleted successfully"
// @Failure 401 {object} dto.ErrorResponse "Unauthorized"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteAllLinks(c *fiber.Ctx) error {
	ctx := c.Context()
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{Error: "Unauthorized"})
	}
	uuidUser, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "Invalid user ID"})
	}

	err = h.svr.DeleteAllLinks(ctx, uuidUser)
	if err != nil {
		h.log.Error("failed to delete all links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{Error: "Failed to delete all links"})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// DeleteBulkShortLinks deletes multiple short links
// @Godoc DeleteBulkShortLinks
// @Summary Delete multiple short links for the authenticated user
// @Description Delete multiple short links created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param request body dto.BulkDeleteLinkRequest true "Bulk Delete Link Request"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.LinkResponse} "Bulk short links deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/bulk-delete [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteBulkShortLinks(c *fiber.Ctx) error {
	ctx := c.Context()
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{Error: "Unauthorized"})
	}
	uuidUser, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "Invalid user ID"})
	}

	var req BulkDeleteLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "Invalid request body"})
	}

	if len(req.IDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "No link IDs provided"})
	}

	resp, err := h.svr.DeleteBulkShortLinks(ctx, uuidUser, req)
	if err != nil {
		h.log.Error("failed to delete bulk short links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{Error: "Bulk delete failed"})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Bulk short links deleted successfully",
		Data:    resp,
	})

}

// CreateBulkShortLinks creates multiple short links
// @Godoc CreateBulkShortLinks
// @Summary Create multiple short links for the authenticated user
// @Description Create multiple short links in bulk for the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param request body dto.BulkCreateLinkRequest true "Bulk Create Link Request"
// @Success 201 {object} dto.SuccessResponse{data=dto.BulkCreateLinkResponse} "Bulk short links created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/bulk-create [post]
// @Security ApiKeyAuth
func (h *Handler) CreateBulkShortLinks(c *fiber.Ctx) error {
	ctx := c.Context()
	var req BulkCreateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "Invalid request body"})
	}

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{Error: "Unauthorized"})
	}
	uuidUser, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{Error: "Invalid user ID"})
	}

	resp, err := h.svr.CreateBulkShortLinks(ctx, uuidUser, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{Error: "Bulk create failed"})
	}
	return c.Status(fiber.StatusCreated).JSON(commons.SuccessResponse{
		Message: "Bulk short links created successfully",
		Data:    resp,
	})
}

// GetUserLinkByShortCode retrieves a short link by its code
// @Godoc GetUserLinkByShortCode
// @Summary Get a short link by its short code for the authenticated user
// @Description Retrieve a specific short link created by the authenticated user using its short code
// @Tags Short Links
// @Accept json
// @Produce json
// @Param shortCode path string true "Short link code"
// @Success 200 {object} dto.SuccessResponse{data=dto.LinkResponse} "Short link retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid short code"
// @Failure 404 {object} dto.ErrorResponse "Short link not found"
// @Failure 403 {object} dto.ErrorResponse "Unauthorized access to this link"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/short-code/{shortCode} [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserLinkByShortCode(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	shortCode := c.Params("shortCode")
	if shortCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Short code is required",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	link, err := h.svr.GetUserLinkByShortCode(ctx, userUUID, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, commons.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, commons.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(commons.ErrorResponse{
				Error: "You are not authorized to access this link",
			})
		default:
			h.log.Error("failed to get user link by short code", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
				Error: "Failed to retrieve short link",
			})
		}
	}

	return c.JSON(commons.SuccessResponse{
		Message: "Short link retrieved successfully",
		Data:    link,
	})

}

// CreateShortLink handles creation of a new short link
// @Godoc CreateShortLink
// @Summary Create a new short link
// @Description Create a new short link for the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param request body dto.CreateLinkRequest true "Create Link Request"
// @Success 201 {object} dto.SuccessResponse{data=dto.LinkResponse} "Short link created successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid request body or missing required fields"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links [post]
// @Security ApiKeyAuth
func (h *Handler) CreateShortLink(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Parse request body into DTO
	var req CreateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate required fields
	if req.OriginalURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Original URL is required",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	// Let the service layer handle the creation using the request DTO
	link, err := h.svr.CreateLinkFromDTO(c.Context(), userUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to create short link: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(commons.SuccessResponse{
		Message: "Short link created successfully",
		Data:    link,
	})
}

// GetUserLinks retrieves all user's short links
// @Godoc GetUserLinks
// @Summary Get all short links for the authenticated user
// @Description Retrieve all short links created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param limit query int false "Limit the number of results"
// @Param offset query int false "Offset for pagination"
// @Param search query string false "Search term for link titles or original URLs"
// @Param order_by query string false "Order by field" Enums(created_at, title, is_active, expired_at)
// @Param ascending query bool false "Order direction (true for ascending, false for descending)"
// @Param start_date query string false "Filter links created after this date (RFC3339 format)"
// @Param end_date query string false "Filter links created before this date (RFC3339 format)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.LinkResponse} "Short links retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links [get]
// @Security ApiKeyAuth
// GetUserLinks retrieves a paginated, filtered, and sorted list of links for the authenticated user.
func (h *Handler) GetUserLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		h.log.Warn("user_id not found in context or is not a string")
		return c.Status(fiber.StatusUnauthorized).JSON(commons.ErrorResponse{
			Error: "Unauthorized: Invalid user session",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	var req GetLinksRequest
	if err := c.QueryParser(&req); err != nil {
		h.log.Warn("Failed to parse query parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
		})
	}

	if req.Order == nil {
		defaultOrder := datastore.ShortlinkOrderColumnCreatedAt
		req.Order = &defaultOrder
	}

	links, pagination, err := h.svr.GetUserLinksWithCount(ctx, userUUID, req)
	if err != nil {
		h.log.Error("failed to get user links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to retrieve short links",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Short links retrieved successfully",
		Data: fiber.Map{
			"links":      links,
			"pagination": pagination,
		},
	})
}

// GetUserLinkByID retrieves a specific short link by ID
// @Godoc GetUserLinkByID
// @Summary Get a short link by ID for the authenticated user
// @Description Retrieve a specific short link created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.LinkResponse} "Short link retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID"
// @Failure 404 {object} dto.ErrorResponse "Short link not found"
// @Failure 403 {object} dto.ErrorResponse "Unauthorized access to this link"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/{id} [get]
// @Security ApiKeyAuth
func (h *Handler) GetUserLinkByID(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	link, err := h.svr.GetUserLinkByID(ctx, userUUID, linkUUID)
	if err != nil {
		switch {
		case errors.Is(err, commons.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, commons.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(commons.ErrorResponse{
				Error: "You are not authorized to access this link",
			})
		default:
			h.log.Error("failed to get user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
				Error: "Failed to retrieve short link",
			})
		}
	}

	return c.JSON(commons.SuccessResponse{
		Message: "Short link retrieved successfully",
		Data:    link,
	})
}

// UpdateLink updates an existing short link
// @Godoc UpdateLink
// @Summary Update a short link by ID for the authenticated user
// @Description Update a specific short link created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Param request body dto.UpdateLinkRequest true "Update Link Request"
// @Success 200 {object} dto.SuccessResponse{data=dto.LinkResponse} "Short link updated successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID or request body"
// @Failure 404 {object} dto.ErrorResponse "Short link not found"
// @Failure 403 {object} dto.ErrorResponse "Unauthorized access to this link"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/{id} [put]
// @Security ApiKeyAuth
func (h *Handler) UpdateLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	var req UpdateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	link, err := h.svr.UpdateUserLink(ctx, userUUID, linkUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to update short link: " + err.Error(),
		})
	}

	return c.JSON(commons.SuccessResponse{
		Message: "Short link updated successfully",
		Data:    link,
	})
}

// DeleteLink deletes a short link
// @Godoc DeleteLink
// @Summary Delete a short link by ID for the authenticated user
// @Description Delete a specific short link created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Success 204 "Short link deleted successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID"
// @Failure 404 {object} dto.ErrorResponse "Short link not found"
// @Failure 403 {object} dto.ErrorResponse "Unauthorized access to this link"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/{id} [delete]
// @Security ApiKeyAuth
func (h *Handler) DeleteLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	h.log.Println("userID:", userID, "linkID:", linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	if err := h.svr.DeleteUserLink(ctx, userUUID, linkUUID); err != nil {
		switch {
		case errors.Is(err, commons.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(commons.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, commons.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(commons.ErrorResponse{
				Error: "You are not authorized to delete this link",
			})
		default:
			h.log.Error("failed to delete user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
				Error: "Failed to delete short link",
			})
		}

	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ToggleLinkStatus activates or deactivates a short link
// @Godoc ToggleLinkStatus
// @Summary Toggle the status of a short link by ID for the authenticated user
// @Description Activate or deactivate a specific short link created by the authenticated user
// @Tags Short Links
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.LinkResponse} "Short link status toggled successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID"
// @Failure 404 {object} dto.ErrorResponse "Short link not found"
// @Failure 403 {object} dto.ErrorResponse "Unauthorized access to this link"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /api/v1/links/{id}/status [patch]
// @Security ApiKeyAuth
func (h *Handler) ToggleLinkStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	link, err := h.svr.ToggleUserLinkStatus(ctx, userUUID, linkUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to toggle link status: " + err.Error(),
		})
	}

	return c.JSON(commons.SuccessResponse{
		Message: "Short link status toggled successfully",
		Data:    link,
	})
}
