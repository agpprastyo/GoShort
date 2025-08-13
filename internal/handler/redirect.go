package handler

import (
	"GoShort/internal/dto"
	"GoShort/internal/service"
	"GoShort/pkg/helper"
	"GoShort/pkg/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

type RedirectHandler struct {
	service service.IRedirectService
	log     *logger.Logger
}

func NewRedirectHandler(service service.IRedirectService, log *logger.Logger) *RedirectHandler {
	return &RedirectHandler{
		service: service,
		log:     log,
	}
}

func (h *RedirectHandler) RedirectToOriginalURL(c *fiber.Ctx) error {
	ctx := c.Context()
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusNotFound).SendString("Link not found")
	}

	originalURL, linkID, isActive, err := h.service.GetOriginalURL(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrLinkNotFound):
			h.log.Warn("link not found", "code", code)
			return c.Status(fiber.StatusNotFound).SendString("Link not found")
		case errors.Is(err, service.ErrLinkNotActive):
			h.log.Warn("link is inactive", "code", code)
			return c.Status(fiber.StatusForbidden).SendString("Link is inactive")
		case errors.Is(err, service.ErrLinkExpired):
			h.log.Warn("link has expired", "code", code)
			return c.Status(fiber.StatusGone).SendString("Link has expired")
		case errors.Is(err, service.ErrClickLimitExceeded):
			h.log.Warn("click limit exceeded for link", "code", code)
			return c.Status(fiber.StatusTooManyRequests).SendString("Click limit exceeded")
		default:
			h.log.Error("unexpected error while retrieving original URL", "error", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal app error")
		}
	}

	if !isActive {
		h.log.Warn("attempted to access inactive link", "code", code)
		return c.Status(fiber.StatusForbidden).SendString("Link is inactive")
	}

	// Extract all values from Fiber context BEFORE starting the goroutine
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")
	country := c.Get("X-Country")
	deviceTypeInfo := c.Get("X-Device-Type")

	// Log the click
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		var clickInfo = dto.CreateLinkStatRequest{
			IpAddress:  helper.StringToPtr(ipAddress),
			UserAgent:  helper.StringToPtr(userAgent),
			Referrer:   helper.StringToPtr(referrer),
			Country:    helper.StringToPtr(country),
			DeviceType: helper.StringToPtr(deviceTypeInfo),
		}

		// Enrich with IP API information
		ipInfo, err := fetchIPInfo(ipAddress)
		if err == nil && ipInfo != nil {
			// Set country if not available
			if clickInfo.Country == nil || *clickInfo.Country == "" {
				clickInfo.Country = helper.StringToPtr(ipInfo.Country)
			}

			// Enhance device type info with mobile status
			deviceType := deviceTypeInfo
			if ipInfo.Mobile {
				if deviceType != "" {
					deviceType += " (Mobile)"
				} else {
					deviceType = "Mobile"
				}
				clickInfo.DeviceType = helper.StringToPtr(deviceType)
			}
		}

		if err := h.service.RecordLinkStat(ctx, linkID, clickInfo); err != nil {
			h.log.Error("failed to record link stat", "link_id", linkID, "error", err)
		}
	}()

	return c.Redirect(originalURL, fiber.StatusFound)
}

// fetchIPInfo gets location and device info from ip-api.com
func fetchIPInfo(ipAddress string) (*dto.IPAPIResponse, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,city,mobile,query", ipAddress)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ipInfo dto.IPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return nil, err
	}

	if ipInfo.Status != "success" {
		return nil, fmt.Errorf("ip-api error: %s", ipInfo.Message)
	}

	return &ipInfo, nil
}
