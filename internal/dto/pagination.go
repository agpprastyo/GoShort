package dto

type Pagination struct {
	Total      int   `json:"total"`
	TotalQuery int   `json:"total_query"`
	Limit      int64 `json:"limit"`
	Offset     int64 `json:"offset"`
	HasMore    bool  `json:"has_more"`
}
