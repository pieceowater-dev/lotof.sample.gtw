package utils

// Pagination represents the pagination details for a query, including the page number and the length of items per page.
type Pagination struct {
	Page   int `json:"page"`   // Page number for pagination.
	Length int `json:"length"` // Number of items per page.
}

// NewPagination creates a new Pagination instance with the provided page number and length of items per page.
func NewPagination(page, length int) Pagination {
	return Pagination{
		Page:   page,
		Length: length,
	}
}
