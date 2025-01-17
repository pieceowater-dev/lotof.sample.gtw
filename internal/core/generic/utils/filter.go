package utils

// Filter represents the filtering criteria for a query, including search term, sorting, and pagination.
type Filter struct {
	Search     string      `json:"search"` // Search term for filtering results.
	Sort       interface{} `json:"sort"`   // Sorting criteria.
	Pagination             // Pagination details.
}

// NewFilter creates a new Filter instance with the provided search term, sorting criteria, and pagination details.
func NewFilter(search string, sort interface{}, pagination Pagination) Filter {
	return Filter{
		Search:     search,
		Sort:       sort,
		Pagination: pagination,
	}
}
