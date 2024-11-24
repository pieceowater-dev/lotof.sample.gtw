package generic

type PaginatedResult[T any] struct {
	Rows []T `json:"rows"`
	Info struct {
		Count int `json:"count"`
	} `json:"info"`
}

func NewPaginatedResult[T any](rows []T, count int) PaginatedResult[T] {
	return PaginatedResult[T]{
		Rows: rows,
		Info: struct {
			Count int `json:"count"`
		}(struct{ Count int }{Count: count}),
	}
}
