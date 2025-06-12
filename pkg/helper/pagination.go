// In pkg/helper/pagination.go
package helper

import "GoShort/internal/dto"

// BuildPaginationInfo creates a standardized pagination response structure
func BuildPaginationInfo(itemCount int, limit, offset int64) dto.Pagination {
	return dto.Pagination{
		Total:   itemCount,
		Limit:   limit,
		Offset:  offset,
		HasMore: int64(itemCount) > (offset + limit),
	}
}
