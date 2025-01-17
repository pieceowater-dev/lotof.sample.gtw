package utils

// SortDirection represents the direction of sorting, either ascending or descending.
type SortDirection string

const (
	// Asc represents ascending sort direction.
	Asc SortDirection = "ASC"
	// Desc represents descending sort direction.
	Desc SortDirection = "DESC"
)

// Sort represents the sorting criteria for a query, including the field to sort by and the direction of sorting.
type Sort[T any] struct {
	Field     string        `json:"field"`     // Field to sort by.
	Direction SortDirection `json:"direction"` // Direction of sorting.
}

// NewSort creates a new Sort instance with the provided field and sort direction.
func NewSort[T any](field string, direction SortDirection) Sort[T] {
	return Sort[T]{
		Field:     field,
		Direction: direction,
	}
}
