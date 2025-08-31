package helper

// BuildPaginationInfo creates a standardized pagination response structure
func BuildPaginationInfo(itemCount int, itemQuery int, limit, offset int64) Pagination {
	return Pagination{
		Total:      itemCount,
		TotalQuery: itemQuery,
		Limit:      limit,
		Offset:     offset,
		HasMore:    int64(itemCount) > (offset + limit),
	}
}

type Pagination struct {
	Total      int   `json:"total"`
	TotalQuery int   `json:"total_query"`
	Limit      int64 `json:"limit"`
	Offset     int64 `json:"offset"`
	HasMore    bool  `json:"has_more"`
}
