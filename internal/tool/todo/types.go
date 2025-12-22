package todo

import (
	"fmt"

	"github.com/Cyclone1070/iav/internal/config"
)

// -- Contract Types --

// TodoStatus represents the status of a todo item.
type TodoStatus string

const (
	TodoStatusPending    TodoStatus = "pending"
	TodoStatusInProgress TodoStatus = "in_progress"
	TodoStatusCompleted  TodoStatus = "completed"
	TodoStatusCancelled  TodoStatus = "cancelled"
)

// TodoDTO is the wire format for todo items
type TodoDTO struct {
	Description string `json:"description"`
	Status      string `json:"status"`
}

// Todo is the domain entity for todo items
type Todo struct {
	description string
	status      TodoStatus
}

// Description returns the todo description.
func (t Todo) Description() string {
	return t.description
}

// Status returns the todo status.
func (t Todo) Status() TodoStatus {
	return t.status
}

// NewTodo creates a validated Todo domain entity.
func NewTodo(description string, status TodoStatus) (Todo, error) {
	if description == "" {
		return Todo{}, ErrEmptyDescription
	}
	return Todo{description: description, status: status}, nil
}

// ReadTodosDTO is the wire format for ReadTodos operation
type ReadTodosDTO struct{}

// ReadTodosRequest is the validated domain entity for ReadTodos operation
type ReadTodosRequest struct{}

// NewReadTodosRequest creates a validated ReadTodosRequest from a DTO
func NewReadTodosRequest(dto ReadTodosDTO, cfg *config.Config) (*ReadTodosRequest, error) {
	return &ReadTodosRequest{}, nil
}

// ReadTodosResponse contains the list of current todos.
type ReadTodosResponse struct {
	Todos []TodoDTO
}

// WriteTodosDTO is the wire format for WriteTodos operation
type WriteTodosDTO struct {
	Todos []TodoDTO `json:"todos"`
}

// WriteTodosRequest is the validated domain entity for WriteTodos operation
type WriteTodosRequest struct {
	todos []Todo
}

// Todos returns the list of todos
func (r *WriteTodosRequest) Todos() []Todo {
	return r.todos
}

// NewWriteTodosRequest creates a validated WriteTodosRequest from a DTO
func NewWriteTodosRequest(dto WriteTodosDTO, cfg *config.Config) (*WriteTodosRequest, error) {
	todos := make([]Todo, len(dto.Todos))
	for i, tDTO := range dto.Todos {
		status := TodoStatus(tDTO.Status)
		// Validate status
		switch status {
		case TodoStatusPending, TodoStatusInProgress, TodoStatusCompleted, TodoStatusCancelled:
			// Valid
		default:
			return nil, fmt.Errorf("todo[%d]: %w %q", i, ErrInvalidStatus, tDTO.Status)
		}

		// Validate description is not empty
		if tDTO.Description == "" {
			return nil, fmt.Errorf("todo[%d]: %w", i, ErrEmptyDescription)
		}

		var err error
		todos[i], err = NewTodo(tDTO.Description, status)
		if err != nil {
			return nil, fmt.Errorf("todo[%d]: %w", i, err)
		}
	}

	return &WriteTodosRequest{
		todos: todos,
	}, nil
}

// WriteTodosResponse contains the result of a WriteTodos operation.
type WriteTodosResponse struct {
	Count int
}
