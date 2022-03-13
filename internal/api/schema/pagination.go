package schema

import (
	"reflect"
)

// PaginatedResponse represents a unified paginated API response
type PaginatedResponse struct {
	Pagination *PaginationMetadata `json:"pagination"`
	Data       interface{}         `json:"data"`
}

// PaginationMetadata represents the metadata present in a PaginatedResponse
type PaginationMetadata struct {
	Offset        uint64 `json:"offset"`
	Limit         uint64 `json:"limit"`
	TotalCount    uint64 `json:"total_count"`
	IncludedCount uint64 `json:"included_count"`
}

// BuildPaginatedResponse builds a unified paginated API response
func BuildPaginatedResponse(offset, limit, totalCount uint64, data interface{}) *PaginatedResponse {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return &PaginatedResponse{
			Pagination: &PaginationMetadata{},
			Data:       []interface{}{},
		}
	}
	return &PaginatedResponse{
		Pagination: &PaginationMetadata{
			Offset:        offset,
			Limit:         limit,
			TotalCount:    totalCount,
			IncludedCount: uint64(val.Len()),
		},
		Data: data,
	}
}
