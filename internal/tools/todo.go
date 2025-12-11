package tools

import (
	"context"
	"fmt"
	"sync"

	"github.com/Cyclone1070/iav/internal/tools/model"
)

// InMemoryTodoStore implements model.TodoStore using an in-memory slice.
type InMemoryTodoStore struct {
	todos []model.Todo
	mu    sync.RWMutex
}

// NewInMemoryTodoStore creates a new instance of InMemoryTodoStore.
func NewInMemoryTodoStore() *InMemoryTodoStore {
	return &InMemoryTodoStore{
		todos: make([]model.Todo, 0),
	}
}

// Read returns the current list of todos.
func (s *InMemoryTodoStore) Read() ([]model.Todo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to ensure isolation
	result := make([]model.Todo, len(s.todos))
	copy(result, s.todos)
	return result, nil
}

// Write replaces the current list of todos with the provided list.
func (s *InMemoryTodoStore) Write(todos []model.Todo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy to ensure isolation
	s.todos = make([]model.Todo, len(todos))
	copy(s.todos, todos)
	return nil
}

// ReadTodos retrieves all todos from the in-memory store.
// Returns an empty list if no todos exist.
func ReadTodos(ctx context.Context, wCtx *model.WorkspaceContext, req model.ReadTodosRequest) (*model.ReadTodosResponse, error) {
	if wCtx.TodoStore == nil {
		return nil, fmt.Errorf("todo store not configured")
	}

	todos, err := wCtx.TodoStore.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read todos: %w", err)
	}

	return &model.ReadTodosResponse{
		Todos: todos,
	}, nil
}

// WriteTodos replaces all todos in the in-memory store.
// This is an atomic operation that completely replaces the todo list.
func WriteTodos(ctx context.Context, wCtx *model.WorkspaceContext, req model.WriteTodosRequest) (*model.WriteTodosResponse, error) {
	if wCtx.TodoStore == nil {
		return nil, fmt.Errorf("todo store not configured")
	}

	if err := wCtx.TodoStore.Write(req.Todos); err != nil {
		return nil, fmt.Errorf("failed to write todos: %w", err)
	}

	return &model.WriteTodosResponse{
		Count: len(req.Todos),
	}, nil
}
