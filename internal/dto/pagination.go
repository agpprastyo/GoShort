// In internal/dto/pagination.go
package dto

type Pagination struct {
	Total   int   `json:"total"`
	Limit   int64 `json:"limit"`
	Offset  int64 `json:"offset"`
	HasMore bool  `json:"has_more"`
}
