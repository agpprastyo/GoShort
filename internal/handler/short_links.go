package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ShortLinkHandler struct {
	service *service.ShortLinkService
}

func NewShortLinkHandler(service *service.ShortLinkService) *ShortLinkHandler {
	return &ShortLinkHandler{
		service: service,
	}
}

// CreateShortLink handles creation of a new short link
func (h *ShortLinkHandler) CreateShortLink(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	var req dto.CreateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	link, err := h.service.CreateLink(c.Context(), userUUID, req)
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

	links, err := h.service.GetUserLinks(ctx, userUUID, 0, 0)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(links)
}

// UpdateLink updates an existing short link
func (h *ShortLinkHandler) UpdateLink(c *fiber.Ctx) error {
	ctx := c.Context()
	userID := c.Locals("user_id").(string)
	linkID := c.Params("id")

	var req dto.UpdateLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
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
