package generic

type SortDirection string

const (
	Asc  SortDirection = "ASC"
	Desc SortDirection = "DESC"
)

type Sort[T any] struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}

func NewSort[T any](field string, direction SortDirection) Sort[T] {
	return Sort[T]{
		Field:     field,
		Direction: direction,
	}
}
