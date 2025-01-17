package utils

// PaginatedResult represents a paginated result set, including the rows of data and pagination information.
type PaginatedResult[T any] struct {
	Rows []T `json:"rows"` // Rows of data in the current page.
	Info struct {
		Count int `json:"count"` // Total count of items.
	} `json:"info"` // Pagination information.
}

// NewPaginatedResult creates a new PaginatedResult instance with the provided rows and total count of items.
func NewPaginatedResult[T any](rows []T, count int) PaginatedResult[T] {
	return PaginatedResult[T]{
		Rows: rows,
		Info: struct {
			Count int `json:"count"`
		}(struct{ Count int }{Count: count}),
	}
}
