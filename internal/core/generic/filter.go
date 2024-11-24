package generic

type Filter struct {
	Search string      `json:"search"`
	Sort   interface{} `json:"sort"`
	Pagination
}

func NewFilter(search string, sort interface{}, pagination Pagination) Filter {
	return Filter{
		Search:     search,
		Sort:       sort,
		Pagination: pagination,
	}
}
