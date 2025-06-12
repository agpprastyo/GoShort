package dto

type CreateLinkStatRequest struct {
	IpAddress  *string `json:"ip_address"`
	UserAgent  *string `json:"user_agent"`
	Referrer   *string `json:"referrer"`
	Country    *string `json:"country"`
	DeviceType *string `json:"device_type"`
}
