package generic

type Pagination struct {
	Page   int `json:"page"`
	Length int `json:"length"`
}

func NewPagination(page, length int) Pagination {
	return Pagination{
		Page:   page,
		Length: length,
	}
}
