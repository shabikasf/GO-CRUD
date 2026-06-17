package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go_crud/store"
	"strings"
	"time"
)

type CacheStore interface{
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}



type TaskRepository interface {
	CreateTask(*store.CreateTaskParams) (*store.Task, error)
	GetTasks(*store.GetTasksParams) ([]*store.Task, error)
	GetTaskByID(*store.GetTaskByIDParams) (*store.Task, error)
	UpdateTaskByID(*store.UpdateTaskByIDParams) (*store.Task, error)
	DeleteTaskByID(*store.DeleteTaskByIDParams) error
	CompleteAllTask(context.Context) error
}



type TaskService struct{
	store TaskRepository
	cache CacheStore
}

func NewTaskService(store TaskRepository, cache CacheStore) *TaskService {
	return &TaskService{
		store: store,
		cache: cache,
	}
}

type CreateTaskInput struct{
	Ctx context.Context
	Title string
}

type GetTasksInput struct{
	Ctx context.Context
	Limit int
	Offset int
	Filter string
	SortBy string
	Order string
}


type GetTaskByIDInput struct{
	Ctx context.Context
	ID int
}

type UpdateTaskByIDInput struct{
	Ctx context.Context
	ID int
	Title string
	Done bool
}

type DeleteTaskByIDInput struct{
	Ctx context.Context
	ID int
}



func (service *TaskService) CreateTask(c CreateTaskInput) (*store.Task, error) {
	if c.Title == "" {
		return nil, ErrInvalidTitle
	}

	task, err := service.store.CreateTask(
		&store.CreateTaskParams{Title: c.Title, Ctx: c.Ctx},
	)

	if err != nil{
		return nil, err
	}

	return task, nil
}

func (service *TaskService) GetTasks(g GetTasksInput) ([]*store.Task, error) {
	tasks, err := service.store.GetTasks(
		&store.GetTasksParams{
			Limit: g.Limit,
			Offset: g.Offset,
			Filter: g.Filter,
			SortBy: g.SortBy,
			Order: g.Order,
			Ctx: g.Ctx,
		},
	)

	if err != nil{
		return nil, err
	}

	return tasks, nil
}

func (service *TaskService) GetTaskByID(g GetTaskByIDInput) (*store.Task, error){
	key := fmt.Sprintf("task:%d", g.ID)

	res, err := service.cache.Get(g.Ctx, key)

	if res != ""{
		var task *store.Task
		json.Unmarshal([]byte(res),  task)
		return task , nil
	}
	
	task, err := service.store.GetTaskByID(
		&store.GetTaskByIDParams{ID: g.ID, Ctx: g.Ctx},
	)

	if err != nil{
		return nil, err
	}


	err = service.cache.Set(g.Ctx, key, task, time.Duration(1)*time.Minute)

	return task, nil
}

func (service *TaskService) UpdateTaskByID(u UpdateTaskByIDInput) (*store.Task, error){
	
	if strings.TrimSpace(u.Title) == ""{
		return nil, ErrInvalidTitle
	}

	task, err := service.store.GetTaskByID(
		&store.GetTaskByIDParams{
			ID: u.ID,
		},
	)

	if err != nil{
		return nil, err
	}

	if task.Done{
		return nil, ErrTaskAlreadyCompleted
	}

	task, err = service.store.UpdateTaskByID(
		&store.UpdateTaskByIDParams{
			ID: u.ID,
			Title: u.Title,
			Done: u.Done,
			Ctx: u.Ctx,
		},
	)

	if err != nil{
		return nil, err
	}

	key := fmt.Sprintf("task:%d",u.ID)

	err = service.cache.Delete(u.Ctx, key)

	return task, nil
}

func (service *TaskService) CompleteAllTask(Ctx context.Context) error {
	return service.store.CompleteAllTask(Ctx)
}

func (service *TaskService) DeleteTaskByID(d DeleteTaskByIDInput) error{


	return service.store.DeleteTaskByID(
		&store.DeleteTaskByIDParams{
			ID: d.ID,
			Ctx: d.Ctx,
		},
	)

}