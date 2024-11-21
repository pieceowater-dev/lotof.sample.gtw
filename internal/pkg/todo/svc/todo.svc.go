package svc

import "app/internal/graph/model"

type TodoService struct {
}

func NewTodoService() *TodoService {
	return &TodoService{}
}

func (s TodoService) Todos() ([]*model.Todo, error) {
	return []*model.Todo{
		{
			ID:       "1",
			Text:     "Test Todo",
			Category: model.TodoCategoryEnumPersonal,
			Done:     false,
		},
		{
			ID:       "2",
			Text:     "Work Todo",
			Category: model.TodoCategoryEnumWork,
			Done:     true,
		},
	}, nil
}
