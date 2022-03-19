package schema

// PaginatedResponse represents a unified paginated API response
type PaginatedResponse[T any] struct {
	Pagination *PaginationMetadata `json:"pagination"`
	Data       []T                 `json:"data"`
}

// PaginationMetadata represents the metadata present in a PaginatedResponse
type PaginationMetadata struct {
	Offset        uint64 `json:"offset"`
	Limit         uint64 `json:"limit"`
	TotalCount    uint64 `json:"total_count"`
	IncludedCount int    `json:"included_count"`
}

// BuildPaginatedResponse builds a unified paginated API response
func BuildPaginatedResponse[T any](offset, limit, totalCount uint64, data []T) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		Pagination: &PaginationMetadata{
			Offset:        offset,
			Limit:         limit,
			TotalCount:    totalCount,
			IncludedCount: len(data),
		},
		Data: data,
	}
}
