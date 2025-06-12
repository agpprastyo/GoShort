package dto

// IPAPIResponse represents the response from ip-api.com
type IPAPIResponse struct {
	Status  string `json:"status"`
	Country string `json:"country"`
	City    string `json:"city"`
	Mobile  bool   `json:"mobile"`
	Query   string `json:"query"`
	Message string `json:"message,omitempty"`
}
