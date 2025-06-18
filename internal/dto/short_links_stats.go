package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateLinkStatRequest struct {
	IpAddress  *string `json:"ip_address"`
	UserAgent  *string `json:"user_agent"`
	Referrer   *string `json:"referrer"`
	Country    *string `json:"country"`
	DeviceType *string `json:"device_type"`
}

type StatsResponse struct {
	TotalUsers    int64 `json:"total_users"`
	TotalLinks    int64 `json:"total_links"`
	ActiveLinks   int64 `json:"active_links"`
	InactiveLinks int64 `json:"inactive_links"`
}
