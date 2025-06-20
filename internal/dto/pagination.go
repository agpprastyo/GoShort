// In internal/dto/pagination.go
package dto

type Pagination struct {
	Total      int   `json:"total"`       // total all available data
	TotalQuery int   `json:"total_query"` // total after query/filter
	Limit      int64 `json:"limit"`
	Offset     int64 `json:"offset"`
	HasMore    bool  `json:"has_more"`
}
