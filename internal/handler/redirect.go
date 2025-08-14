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
	"strings"
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

type CreateLinkStatRequest struct {
	IpAddress  *string
	UserAgent  *string
	Referrer   *string
	Country    *string
	DeviceType *string
}

type IPAPIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Country string `json:"country"`
	City    string `json:"city"`
	Mobile  bool   `json:"mobile"`
	Query   string `json:"query"`
}

var (
	ErrLinkNotFound  = errors.New("link not found")
	ErrLinkNotActive = errors.New("link is not active")
)

func (h *RedirectHandler) RedirectToOriginalURL(c *fiber.Ctx) error {
	ctx := c.Context()
	code := c.Params("code")
	if code == "" {
		return c.Status(fiber.StatusNotFound).SendString("Link not found")
	}

	originalURL, linkID, isActive, err := h.service.GetOriginalURL(ctx, code)
	if err != nil {
		switch {
		case errors.Is(err, ErrLinkNotFound):
			h.log.Warn("link not found", "code", code)
			return c.Status(fiber.StatusNotFound).SendString("Link not found")
		case errors.Is(err, ErrLinkNotActive):
			h.log.Warn("link is inactive", "code", code)
			return c.Status(fiber.StatusForbidden).SendString("Link is inactive")
		// Add other error cases here (Expired, ClickLimit, etc.)
		default:
			h.log.Println("unexpected error while retrieving original URL", "error", err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal app error")
		}
	}

	if !isActive {
		h.log.Warn("attempted to access inactive link", "code", code)
		return c.Status(fiber.StatusForbidden).SendString("Link is inactive")
	}

	ipAddress := c.Get("CF-Connecting-IP")
	if ipAddress == "" {
		// X-Forwarded-For can be a list of IPs. The first one is the original client.
		ips := strings.Split(c.Get("X-Forwarded-For"), ",")
		if len(ips) > 0 && ips[0] != "" {
			ipAddress = strings.TrimSpace(ips[0])
		}
	}
	if ipAddress == "" {
		ipAddress = c.IP()
	}

	userAgent := c.Get("User-Agent")
	referrer := c.Get("Referer")
	country := c.Get("CF-IPCountry")

	deviceType := "Desktop"
	if strings.Contains(strings.ToLower(userAgent), "mobile") {
		deviceType = "Mobile"
	} else if strings.Contains(strings.ToLower(userAgent), "tablet") {
		deviceType = "Tablet"
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		clickInfo := dto.CreateLinkStatRequest{
			IpAddress:  helper.StringToPtr(ipAddress),
			UserAgent:  helper.StringToPtr(userAgent),
			Referrer:   helper.StringToPtr(referrer),
			Country:    helper.StringToPtr(country),
			DeviceType: helper.StringToPtr(deviceType),
		}

		ipInfo, err := fetchIPInfo(ipAddress)
		if err == nil && ipInfo != nil {

			if clickInfo.Country == nil || *clickInfo.Country == "" {
				clickInfo.Country = helper.StringToPtr(ipInfo.Country)
			}
			if ipInfo.Mobile && (clickInfo.DeviceType == nil || *clickInfo.DeviceType != "Mobile") {
				clickInfo.DeviceType = helper.StringToPtr("Mobile")
			}
		}

		if err := h.service.RecordLinkStat(ctx, linkID, clickInfo); err != nil {
			h.log.Println("failed to record link stat", "link_id", linkID, "error", err)
		}
	}()

	return c.Redirect(originalURL, fiber.StatusFound)
}

func fetchIPInfo(ipAddress string) (*IPAPIResponse, error) {
	if ipAddress == "127.0.0.1" || ipAddress == "::1" || strings.HasPrefix(ipAddress, "172.") {
		return nil, errors.New("private IP address")
	}

	client := &http.Client{Timeout: 2 * time.Second}
	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,message,country,city,mobile,query", ipAddress)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ipInfo IPAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&ipInfo); err != nil {
		return nil, err
	}

	if ipInfo.Status != "success" {
		return nil, fmt.Errorf("ip-api error: %s", ipInfo.Message)
	}

	return &ipInfo, nil
}
