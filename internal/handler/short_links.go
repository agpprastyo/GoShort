package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/repository"
	"errors"

	"GoShort/internal/service"
	"GoShort/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
// @Router /short-links/short-code/{shortCode} [get]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) GetUserLinkByShortCode(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	shortCode := c.Params("shortCode")
	if shortCode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Short code is required",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	link, err := h.svr.GetUserLinkByShortCode(ctx, userUUID, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, service.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error: "You are not authorized to access this link",
			})
		default:
			h.log.Error("failed to get user link by short code", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error: "Failed to retrieve short link",
			})
		}
	}

	return c.JSON(dto.SuccessResponse{
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
// @Router /short-links [post]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) CreateShortLink(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	// Parse request body into DTO
	var req dto.CreateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	// Validate required fields
	if req.OriginalURL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Original URL is required",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	// Let the service layer handle the creation using the request DTO
	link, err := h.svr.CreateLinkFromDTO(c.Context(), userUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Failed to create short link: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(dto.SuccessResponse{
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
// @Router /short-links [get]
// @Security ApiKeyAuth
// GetUserLinks retrieves a paginated, filtered, and sorted list of links for the authenticated user.
func (h *ShortLinkHandler) GetUserLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	// 1. Akses `c.Locals` dengan aman untuk mendapatkan ID pengguna
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		h.log.Warn("user_id not found in context or is not a string")
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{
			Error: "Unauthorized: Invalid user session",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID format",
		})
	}

	// 2. Gunakan QueryParser untuk mengisi DTO secara otomatis dari query string
	var req dto.GetLinksRequest
	if err := c.QueryParser(&req); err != nil {
		h.log.Warn("Failed to parse query parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid query parameters: " + err.Error(),
		})
	}

	// 3. Atur nilai default jika tidak disediakan oleh pengguna
	// Jika parameter 'order' tidak ada di URL, req.Order akan nil.
	if req.Order == nil {
		// Tetapkan nilai default untuk sorting
		defaultOrder := repository.ShortlinkOrderColumnCreatedAt
		req.Order = &defaultOrder
	}

	// Nilai default untuk 'ascending' (false) biasanya ditangani di lapisan repository/database
	// Jika req.Ascending nil, itu bisa diartikan sebagai false.

	// 4. Panggil service dengan DTO yang sudah terisi
	links, pagination, err := h.svr.GetUserLinksWithCount(ctx, userUUID, req)
	if err != nil {
		h.log.Error("failed to get user links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Failed to retrieve short links",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.SuccessResponse{
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
// @Router /short-links/{id} [get]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) GetUserLinkByID(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	link, err := h.svr.GetUserLinkByID(ctx, userUUID, linkUUID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, service.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error: "You are not authorized to access this link",
			})
		default:
			h.log.Error("failed to get user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
				Error: "Failed to retrieve short link",
			})
		}
	}

	return c.JSON(dto.SuccessResponse{
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
// @Router /short-links/{id} [put]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) UpdateLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	var req dto.UpdateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	link, err := h.svr.UpdateUserLink(ctx, userUUID, linkUUID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Failed to update short link: " + err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
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
// @Router /short-links/{id} [delete]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) DeleteLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	h.log.Println("userID:", userID, "linkID:", linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	if err := h.svr.DeleteUserLink(ctx, userUUID, linkUUID); err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
				Error: "Short link not found",
			})
		case errors.Is(err, service.ErrUnauthorized):
			return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{
				Error: "You are not authorized to delete this link",
			})
		default:
			h.log.Error("failed to delete user link", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
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
// @Router /short-links/{id}/status [patch]
// @Security ApiKeyAuth
func (h *ShortLinkHandler) ToggleLinkStatus(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	link, err := h.svr.ToggleUserLinkStatus(ctx, userUUID, linkUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: "Failed to toggle link status: " + err.Error(),
		})
	}

	return c.JSON(dto.SuccessResponse{
		Message: "Short link status toggled successfully",
		Data:    link,
	})
}
