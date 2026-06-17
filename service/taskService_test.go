package service

import (
	"context"
	"errors"
	"go_crud/store"
	"reflect"
	"testing"
	"time"
)

type FakeTaskStore struct{
	CreateTaskFn func(
	*store.CreateTaskParams,
	) (*store.Task, error)

	GetTaskFn func(
		*store.GetTasksParams,
	) ([]*store.Task, error)

	GetTaskByIDFn func(
		*store.GetTaskByIDParams,
	) (*store.Task, error)
	
	UpdateTaskFn func(
		*store.UpdateTaskByIDParams,
	) (*store.Task, error)

	CompleteAllTaskFn func(ctx context.Context) error

	DeleteTaskByIDFn func(
		*store.DeleteTaskByIDParams,
	) error
}


func (f *FakeTaskStore) CreateTask(
	p *store.CreateTaskParams,
) (*store.Task, error) {
	return f.CreateTaskFn(p)
}

func (f *FakeTaskStore) GetTasks(
	p *store.GetTasksParams,
) ([] *store.Task, error){
	return f.GetTaskFn(p)
}

func (f *FakeTaskStore) GetTaskByID(
	p *store.GetTaskByIDParams,
)(*store.Task, error){
	return f.GetTaskByIDFn(p)
}

func (f *FakeTaskStore) UpdateTaskByID(
	p *store.UpdateTaskByIDParams,
)(*store.Task, error){
	return f.UpdateTaskFn(p)
}

func (f *FakeTaskStore) CompleteAllTask(ctx context.Context) error{
	return f.CompleteAllTaskFn(ctx)
}

func (f *FakeTaskStore) DeleteTaskByID(
	p *store.DeleteTaskByIDParams,
) error{
	return f.DeleteTaskByIDFn(p)
}

type FakeCacheStore struct{
	GetFn func(ctx context.Context, key string, value any) (string, error)
	SetFn func(ctx context.Context, key string, value any, ttl time.Duration) error
	DeleteFn func(ctx context.Context, key string) error
}

func (f *FakeCacheStore) Get(ctx context.Context, key string, value any) (string, error){
	return f.GetFn(ctx, key, value)
}

func (f *FakeCacheStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error{
	return f.SetFn(ctx, key, value, ttl)
}

func (f *FakeCacheStore) Delete(ctx context.Context, key string) error{
	return f.DeleteFn(ctx, key)
}

func TestCreateTask_InvalidTitle(t *testing.T) {
	service := NewTaskService(&FakeTaskStore{}, &FakeCacheStore{})

	_, err := service.CreateTask(
		CreateTaskInput{
			Title: "",
		},
	)

	if !errors.Is(err, ErrInvalidTitle){
		t.Fatalf("want %v but got %v", ErrInvalidTitle, err)
	}
}

func TestCreateTask_Success(t *testing.T) {
	service := NewTaskService(&FakeTaskStore{
		CreateTaskFn: func(ctp *store.CreateTaskParams) (*store.Task, error) {
			return &store.Task{
				ID: 1,
				Title: ctp.Title,
				Done: false,
			}, nil
		},
	})

	want := &store.Task{
		ID: 1,
		Title: "Hello world",
		Done: false,
	}

	got, err := service.CreateTask(CreateTaskInput{Title: "Hello world"})

	if err != nil{
		t.Fatalf("expected Success but Got %v", err)
	}

	if !reflect.DeepEqual(want, got){
		t.Fatalf("test failed wanted %v but got %v", want, got)
	}
}

func TestGetTasks_Success(t *testing.T){

	want := []*store.Task{
		&store.Task{
			ID: 1,
			Title: "Hello World",
			Done: true,
		},
	}

	service := NewTaskService(
		&FakeTaskStore{
			GetTaskFn: func(gtp *store.GetTasksParams) ([]*store.Task, error) {
				return []*store.Task{
					&store.Task{
						ID: 1,
						Title: "Hello World",
						Done: true,
					},
				}, nil
			},
		},
	)

	got, err := service.GetTasks(GetTasksInput{
		Limit: 1,
		Offset: 1,
		Filter: "true",
		SortBy: "id",
		Order: "DESC",
	})

	if err != nil{
		t.Fatalf("expected success but got %v", err)
	}

	if !reflect.DeepEqual(want, got){
		t.Fatalf("test failed wanted %v but got %v", want, got)
	}


}

func TestGetTaskByID_Success(t *testing.T){
	want := &store.Task{
		ID: 1,
		Title: "Hello World",
		Done: true,
	}

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
					ID: 1,
					Title: "Hello World",
					Done: true,
				}, nil
			}

			return nil, store.ErrTaskNotFound

		},
	})

	got, err := service.GetTaskByID(GetTaskByIDInput{ID: 1})

	if err != nil{
		t.Fatalf("expected success but got %v", err)
	}

	if !reflect.DeepEqual(want, got){
		t.Fatalf("test failed wanted %v but got %v", want, got)
	}
}

func TestGetTaskByID_TaskNotFound(t *testing.T){

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
					ID: 1,
					Title: "Hello World",
					Done: true,
				}, nil
			}

			return nil, store.ErrTaskNotFound

		},
	})

	_, err := service.GetTaskByID(GetTaskByIDInput{ID: 2})

	if err != nil && !errors.Is(err, store.ErrTaskNotFound){
		t.Fatalf("expected %v but got %v",store.ErrTaskNotFound, err)
	}

}

func TestUpdateTaskByID_Success(t *testing.T){
	want := &store.Task{
		ID: 1,
		Title: "Hello World",
		Done: true,
	}

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
						ID: 1,
						Title: "Hello World",
						Done: false,
					}, nil
			}

			return nil, store.ErrTaskNotFound
		},

		UpdateTaskFn: func(utbi *store.UpdateTaskByIDParams) (*store.Task, error) {
			return &store.Task{
					ID: utbi.ID,
					Title: utbi.Title,
					Done: utbi.Done,
				}, nil
		},
	})

	got, err := service.UpdateTaskByID(UpdateTaskByIDInput{ID: 1, Title: "Hello World", Done: true})

	if err != nil{
		t.Fatalf("expected success but got %v", err)
	}

	if !reflect.DeepEqual(got, want){
		t.Fatalf("wanted %v but got %v", want, got)
	}
}


func TestUpdateTaskByID_TaskNotFound(t *testing.T){
	want := store.ErrTaskNotFound

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
						ID: 1,
						Title: "Hello World",
						Done: false,
					}, nil
			}

			return nil, store.ErrTaskNotFound
		},

		UpdateTaskFn: func(utbi *store.UpdateTaskByIDParams) (*store.Task, error) {
			return &store.Task{
					ID: utbi.ID,
					Title: utbi.Title,
					Done: utbi.Done,
				}, nil
		},
	})

	_, err := service.UpdateTaskByID(UpdateTaskByIDInput{ID: 2, Title: "Hello World", Done: true})

	if err != nil && !errors.Is(err, want){
		t.Fatalf("expected %v but got %v", want, err)
	}
}

func TestUpdateTaskByID_InvalidTitle(t *testing.T){
	want := ErrInvalidTitle

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
						ID: 1,
						Title: "Hello World",
						Done: false,
					}, nil
			}

			return nil, store.ErrTaskNotFound
		},

		UpdateTaskFn: func(utbi *store.UpdateTaskByIDParams) (*store.Task, error) {
			return &store.Task{
					ID: utbi.ID,
					Title: utbi.Title,
					Done: utbi.Done,
				}, nil
		},
	})

	_, err := service.UpdateTaskByID(UpdateTaskByIDInput{ID: 1, Title: "", Done: true})

	if err != nil && !errors.Is(err, want){
		t.Fatalf("expected %v but got %v", want, err)
	}
}

func TestUpdateTaskByID_TaskCompleted(t *testing.T){
	want := ErrTaskAlreadyCompleted

	service := NewTaskService(&FakeTaskStore{
		GetTaskByIDFn: func(gtbi *store.GetTaskByIDParams) (*store.Task, error) {
			if gtbi.ID == 1{
				return &store.Task{
						ID: 1,
						Title: "Hello World",
						Done: true,
					}, nil
			}

			return nil, store.ErrTaskNotFound
		},

		UpdateTaskFn: func(utbi *store.UpdateTaskByIDParams) (*store.Task, error) {
			return &store.Task{
					ID: utbi.ID,
					Title: utbi.Title,
					Done: utbi.Done,
				}, nil
		},
	})

	_, err := service.UpdateTaskByID(UpdateTaskByIDInput{ID: 1, Title: "Hello World", Done: true})

	if err != nil && !errors.Is(err, want){
		t.Fatalf("expected %v but got %v", want, err)
	}
}

func TestCompleteAllTask_Success(t *testing.T) {
	var want error = nil

	service := NewTaskService(&FakeTaskStore{
		CompleteAllTaskFn: func(ctx context.Context) error {
			return nil
		},
	})

	ctx := context.Background()

	got := service.CompleteAllTask(ctx)

	if !errors.Is(got, want){
		t.Fatalf("expected %v but got %v", want, got)
	}
}

func TestCompleteAllTask_Failed(t *testing.T) {

	want := errors.New("transaction failed")

	service := NewTaskService(&FakeTaskStore{
		CompleteAllTaskFn: func(ctx context.Context) error {
			return errors.New("transaction failed")
		},
	})

	ctx := context.Background()

	got := service.CompleteAllTask(ctx)

	if got == nil{
		t.Fatalf("expected %v but got %v", want, got)
	}
}

func TestCompleteAllTask_TaskNotFound(t *testing.T) {

	want := store.ErrTaskNotFound

	service := NewTaskService(&FakeTaskStore{
		CompleteAllTaskFn: func(ctx context.Context) error {
			return store.ErrTaskNotFound
		},
	})


	ctx := context.Background()
	got := service.CompleteAllTask(ctx)

	if got == nil{
		t.Fatalf("expected %v but got %v", want, got)
	}
}

func TestDeleteTaskByID_Success(t *testing.T) {
	var want error = nil

	service := NewTaskService(&FakeTaskStore{
		DeleteTaskByIDFn: func(dtbi *store.DeleteTaskByIDParams) error {
			if dtbi.ID == 1{
				return nil
			}

			return store.ErrTaskNotFound
		},
	})

	got := service.DeleteTaskByID(DeleteTaskByIDInput{ID: 1})

	if !errors.Is(got, want){
		t.Fatalf("expected %v but got %v", want, got)
	}
}

func TestDeleteTaskByID_Failed(t *testing.T) {
	var ErrDeleteTask = errors.New("failed to delete task")
	want := ErrDeleteTask
	service := NewTaskService(&FakeTaskStore{
		DeleteTaskByIDFn: func(dtbi *store.DeleteTaskByIDParams) error {
			return ErrDeleteTask
		},
	})

	got := service.DeleteTaskByID(DeleteTaskByIDInput{ID: 1})

	if !errors.Is(got, want){
		t.Fatalf("expected %v but got %v", want, got)
	}
}

func TestDeleteTaskByID_TaskNotFound(t *testing.T) {
	want := store.ErrTaskNotFound

	service := NewTaskService(&FakeTaskStore{
		DeleteTaskByIDFn: func(dtbi *store.DeleteTaskByIDParams) error {
			if dtbi.ID == 1{
				return nil
			}

			return store.ErrTaskNotFound
		},
	})

	got := service.DeleteTaskByID(DeleteTaskByIDInput{ID: 2})

	if !errors.Is(got, want){
		t.Fatalf("expected %v but got %v", want, got)
	}
}
