package admin

import (
	"GoShort/internal/commons"
	"GoShort/internal/datastore"
	"GoShort/internal/shortlink"
	"GoShort/pkg/logger"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	adminService IService
	log          *logger.Logger
	Validator    *validator.Validate
}

func NewHandler(adminService IService, log *logger.Logger, validator *validator.Validate) IHandler {
	return &Handler{
		adminService: adminService,
		log:          log,
		Validator:    validator,
	}
}

type IHandler interface {
	GetSystemStats(c *fiber.Ctx) error
	ListAllLinks(c *fiber.Ctx) error
	GetLink(c *fiber.Ctx) error
	ListUserLinks(c *fiber.Ctx) error
	ToggleLinkStatus(c *fiber.Ctx) error
}

// GetSystemStats retrieves system statistics
// @Godoc GetSystemStats
// @Summary Get system statistics
// @Description Retrieve system statistics including total users, links, and clicks
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} dto.SuccessResponse "System stats retrieved successfully"
// @Failure 403 {object} dto.ErrorResponse "Forbidden - Admin access required"
// @Failure 500 {object} dto.ErrorResponse "Failed to retrieve system stats"
// @Router /api/v1/admin/stats [get]
// @Security ApiKeyAuth
func (h *Handler) GetSystemStats(c *fiber.Ctx) error {
	stats, err := h.adminService.GetStats(c.Context())
	if err != nil {
		h.log.Error("failed to get system stats", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to retrieve system stats",
		})
	}
	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "System stats retrieved successfully",
		Data:    stats,
	})
}

// ListAllLinks lists all short links in the system
// @Godoc ListAllLinks
// @Summary List all short links
// @Description Retrieve a paginated list of all short links
// @Tags admin
// @Accept json
// @Produce json
// @Param limit query int false "Number of links to return per page"
// @Param offset query int false "Offset for pagination"
// @Param search query string false "Search term to filter links by title or URL"
// @Param order_by query string false "Order by field" Enums(created_at, title, is_active, expired_at)
// @Param ascending query bool false "Order direction (true for ascending, false for descending)"
// @Param start_date query string false "Start date for filtering links (RFC3339 format)"
// @Param end_date query string false "End date for filtering links (RFC3339 format)"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.LinkResponse}  "Links retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} dto.ErrorResponse "Failed to retrieve links"
// @Router /api/v1/admin/links [get]
func (h *Handler) ListAllLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	var req shortlink.GetLinksRequest

	if err := c.BodyParser(req); err != nil {
		h.log.Error("failed to parse request body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid request body",
		})
	}

	if req.Order == nil {
		defaultOrder := datastore.ShortlinkOrderColumnCreatedAt
		req.Order = &defaultOrder
	}

	links, pagination, err := h.adminService.ListAllLinks(ctx, req)
	if err != nil {
		h.log.Error("failed to list all links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to retrieve links",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Links retrieved successfully",
		Data: fiber.Map{
			"links":      links,
			"pagination": pagination,
		},
	})

}

// GetLink retrieves a specific short link by ID
// @Godoc GetLink
// @Summary Get a short link by ID
// @Description Retrieve a short link by its unique ID
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Success 200 {object} dto.SuccessResponse "Link retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID"
// @Failure 404 {object} dto.ErrorResponse "Link not found"
// @Failure 500 {object} dto.ErrorResponse "Failed to retrieve link"
// @Router /api/v1/admin/links/{id} [get]
func (h *Handler) GetLink(c *fiber.Ctx) error {
	ctx := c.Context()

	linkID := c.Params("id")

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		h.log.Error("invalid link ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	link, err := h.adminService.GetLinkByID(ctx, linkUUID)
	if err != nil {
		h.log.Error("failed to get link", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to retrieve link",
		})
	}

	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "Link retrieved successfully",
		Data:    link,
	})

}

// ListUserLinks lists all short links for a specific user
// @Godoc ListUserLinks
// @Summary List all short links for a user
// @Description Retrieve a paginated list of all short links for a specific user
// @Tags admin
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Param limit query int false "Number of links to return per page"
// @Param offset query int false "Offset for pagination"
// @Param search query string false "Search term to filter links by title or URL"
// @Param order_by query string false "Order by field" Enums(created_at, title, is_active, expired_at)
// @Param ascending query bool false "Order direction (true for ascending, false for descending)"
// @Param start_date query string false "Start date for filtering links (RFC3339 format)"
// @Param end_date query string false "End date for filtering links (RFC3339 format)"
// @Success 200 {object} dto.SuccessResponse "User links retrieved successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid user ID or query parameters"
// @Failure 404 {object} dto.ErrorResponse "User not found"
// @Failure 500 {object} dto.ErrorResponse "Failed to retrieve user links"
// @Router /api/v1/admin/users/{userId}/links [get]
func (h *Handler) ListUserLinks(c *fiber.Ctx) error {
	ctx := c.Context()

	userID := c.Params("userId")

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.log.Error("invalid user ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid user ID",
		})
	}

	var req shortlink.GetLinksRequest

	if err := c.QueryParser(&req); err != nil {
		h.log.Error("failed to parse query parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid query parameters",
		})
	}

	if req.Order == nil {
		// Default order by created_at if not specified
		defaultOrder := datastore.ShortlinkOrderColumnCreatedAt
		req.Order = &defaultOrder
	}

	userLinks, pagination, err := h.adminService.ListUserLinks(ctx, userUUID, req)
	if err != nil {
		h.log.Error("failed to list user links", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to retrieve user links",
		})

	}
	return c.Status(fiber.StatusOK).JSON(commons.SuccessResponse{
		Message: "User links retrieved successfully",
		Data: fiber.Map{
			"links":      userLinks,
			"pagination": pagination,
		},
	})
}

// ToggleLinkStatus toggles the active status of a short link
// @Godoc ToggleLinkStatus
// @Summary Toggle the active status of a short link
// @Description Activate or deactivate a short link by its ID
// @Tags admin
// @Accept json
// @Produce json
// @Param id path string true "Short link ID"
// @Success 204 "Link status toggled successfully"
// @Failure 400 {object} dto.ErrorResponse "Invalid link ID"
// @Failure 500 {object} dto.ErrorResponse "Failed to toggle link status"
// @Router /api/v1/admin/links/{id}/status [patch]
func (h *Handler) ToggleLinkStatus(c *fiber.Ctx) error {
	ctx := c.Context()

	linkID := c.Params("id")

	linkUUID, err := uuid.Parse(linkID)
	if err != nil {
		h.log.Error("invalid link ID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(commons.ErrorResponse{
			Error: "Invalid link ID",
		})
	}

	err = h.adminService.ToggleLinkStatus(ctx, linkUUID)
	if err != nil {
		h.log.Error("failed to toggle link status", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(commons.ErrorResponse{
			Error: "Failed to toggle link status",
		})
	}

	return c.Status(fiber.StatusNoContent).JSON(commons.SuccessResponse{
		Message: "Link status toggled successfully",
	})

}
