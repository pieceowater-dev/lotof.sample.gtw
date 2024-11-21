package ctrl

import (
	"app/internal/graph/model"
	"app/internal/pkg/todo/svc"
	"fmt"
)

type TodoController struct {
	todoService *svc.TodoService
}

func NewTodoController(service *svc.TodoService) *TodoController {
	return &TodoController{todoService: service}
}

func (c TodoController) Todos() ([]*model.Todo, error) {
	todos, err := c.todoService.Todos()
	if err != nil {
		fmt.Println("Error fetching todos:", err)
		return nil, err
	}
	return todos, nil
}
