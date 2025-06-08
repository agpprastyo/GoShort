// internal/handler/redirect.go
package handler

import (
	"GoShort/internal/service"
	"GoShort/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type RedirectHandler struct {
	service *service.RedirectService
	log     *logger.Logger
}

func NewRedirectHandler(service *service.RedirectService, log *logger.Logger) *RedirectHandler {
	return &RedirectHandler{
		service: service,
		log:     log,
	}
}

func (h *RedirectHandler) RedirectToOriginalURL(c *fiber.Ctx) error {
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusNotFound).SendString("Link not found")
	}

	originalURL, isActive, err := h.service.GetOriginalURL(c.Context(), code)
	if err != nil {
		h.log.Error("failed to get original URL", "error", err)
		return c.Status(fiber.StatusNotFound).SendString("Link not found")
	}

	if !isActive {
		return c.Status(fiber.StatusGone).SendString("This link has been deactivated")
	}

	return c.Redirect(originalURL, fiber.StatusFound)
}
