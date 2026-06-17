package store

import "errors"

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrCreateTask = errors.New("failed to create Task")
	ErrFetchTask = errors.New("failed to fetch task/tasks")
	ErrUpdateTask = errors.New("failed to update task/tasks")
	ErrDeleteTask = errors.New("failed to delete task/tasks")
	ErrCreateUser = errors.New("failed to create user")
	ErrFetchUser = errors.New("failed to fetch user")
	ErrUserNotFound = errors.New("user not found")
)