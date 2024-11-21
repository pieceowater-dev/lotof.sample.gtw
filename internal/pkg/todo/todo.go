package todo

import (
	"app/internal/pkg/todo/ctrl"
	"app/internal/pkg/todo/svc"
)

type Module struct {
	API *ctrl.TodoController
}

func NewTodoModule() Module {
	service := svc.NewTodoService()
	controller := ctrl.NewTodoController(service)
	return Module{
		API: controller,
	}
}
