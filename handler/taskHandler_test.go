package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go_crud/service"
	"go_crud/store"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type FakeTaskService struct {
	CreateTaskFn func(service.CreateTaskInput) (*store.Task, error)
	GetTasksFn func(service.GetTasksInput) ([]*store.Task, error)
	GetTaskByIDFn func(service.GetTaskByIDInput) (*store.Task, error)
	UpdateTaskByIDFn func(service.UpdateTaskByIDInput) (*store.Task, error)
	CompleteAllTaskFn func(ctx context.Context) error
	DeleteTaskByIDFn func(service.DeleteTaskByIDInput) error
}

func (f *FakeTaskService) CreateTask(input service.CreateTaskInput) (*store.Task, error){
	return f.CreateTaskFn(input)
}

func (f *FakeTaskService) GetTasks(input service.GetTasksInput) ([]*store.Task, error){
	return  f.GetTasksFn(input)
}

func (f *FakeTaskService) GetTaskByID(input service.GetTaskByIDInput) (*store.Task, error){
	return f.GetTaskByIDFn(input)
}

func (f *FakeTaskService) UpdateTaskByID(input service.UpdateTaskByIDInput) (*store.Task, error){
	return f.UpdateTaskByIDFn(input)
}

func (f *FakeTaskService) CompleteAllTask(ctx context.Context) error{
	return f.CompleteAllTaskFn(ctx)
}

func (f *FakeTaskService) DeleteTaskByID(input service.DeleteTaskByIDInput) error{
	return f.DeleteTaskByIDFn(input)
}


func TestCreateTask_Success(t *testing.T){
	fakeService := &FakeTaskService{
		CreateTaskFn: func(
			input service.CreateTaskInput,
		) (*store.Task, error) {

			return &store.Task{
				ID: 1,
				Title: input.Title,
				Done: false,
			}, nil
		},
	}

	handler := NewTaskHandler(fakeService)

	body := strings.NewReader(`{
		"title": "Learn Go"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/tasks", body)

	rr := httptest.NewRecorder()

	handler.CreateTask(rr, req)

	if rr.Code != http.StatusCreated{
		t.Fatalf("expected %d got %d", http.StatusCreated, rr.Code)
	}

	var got store.Task

	err := json.Unmarshal(rr.Body.Bytes(), &got)

	if err != nil{
		t.Fatal(err)
	}

	if got.Title != "Learn Go"{
		t.Fatalf("expected Learn Go but got %s", got.Title)
	}

}

func TestCreateTask_Failed(t *testing.T){
	fakeService := &FakeTaskService{
		CreateTaskFn: func(
			input service.CreateTaskInput,
		) (*store.Task, error) {

			return nil, errors.New("failed to create task")
		},
	}

	handler := NewTaskHandler(fakeService)

	body := strings.NewReader(`{
		"title": "Learn Go"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/tasks", body)

	rr := httptest.NewRecorder()

	handler.CreateTask(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("expected %d got %d", http.StatusInternalServerError, rr.Code)
	}

}

func TestCreateTask_InvalidTitle(t *testing.T){
	fakeService := &FakeTaskService{
		CreateTaskFn: func(
			input service.CreateTaskInput,
		) (*store.Task, error) {

			return nil, errors.New("failed to create task")
		},
	}

	handler := NewTaskHandler(fakeService)

	body := strings.NewReader(`{
		"title": "    "
	}`)

	req := httptest.NewRequest(http.MethodPost, "/tasks", body)

	rr := httptest.NewRecorder()

	handler.CreateTask(rr, req)

	if rr.Code != http.StatusBadRequest{
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}

}

func TestGetTasks_Success(t *testing.T){
	fakeService := &FakeTaskService{
		GetTasksFn: func(gti service.GetTasksInput) ([]*store.Task, error) {
			return []*store.Task{
				&store.Task{
					ID: 1,
					Title: "Hello Go",
					Done: false,
				},
			}, nil
		},
	}

	handler := NewTaskHandler(fakeService)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	handler.GetTasks(rr, req)

	want := []*store.Task{
		&store.Task{
			ID: 1,
			Title: "Hello Go",
			Done: false,
		},
	}

	var tasks []*store.Task

	if err := json.Unmarshal(rr.Body.Bytes(), &tasks); err != nil{
		t.Fatal(err)
	}

	if rr.Code != http.StatusOK{
		t.Fatalf("wanted success but got %d", rr.Code)
	}

	if !reflect.DeepEqual(tasks, want){
		t.Fatalf("wanted %v but got %v", want, tasks)
	}
}

func TestGetTasks_Invalid(t *testing.T){
	tests := []struct{
		name string
		url string
	}{
		{
			name: "invalid page value",
			url: "/tasks?page=abc",
		},
		{
			name: "invalid limit value",
			url: "/tasks?limit=abc",
		},
		{
			name: "invalid sort field",
			url: "/tasks?sort=salary",
		},
		{
			name: "invalid filter value",
			url: "/tasks?done=abc",
		},
		{
			name: "invalid order value",
			url: "/tasks?order=sideways",
		},
	}

	for _, tt := range tests{
		t.Run(tt.name, func (t *testing.T)  {
			handler := NewTaskHandler(&FakeTaskService{})
			req := httptest.NewRequest(
				http.MethodGet,
				tt.url,
				nil,
			)

			rr := httptest.NewRecorder()

			handler.GetTasks(rr, req)

			if rr.Code != http.StatusBadRequest{
				t.Fatalf("wanted %d but got %d", http.StatusBadRequest, rr.Code)
			}

		})
	}
}

func TestGetTasks_Failed(t *testing.T){
	fakeTaskService := &FakeTaskService{
		GetTasksFn: func(gti service.GetTasksInput) ([]*store.Task, error) {
			return nil, store.ErrFetchTask
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	handler := NewTaskHandler(fakeTaskService)

	handler.GetTasks(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("wanted %d but got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestGetTaskByID_Success(t *testing.T){
	
	fakeTaskService := &FakeTaskService{
		GetTaskByIDFn: func(gtbi service.GetTaskByIDInput) (*store.Task, error) {
			if gtbi.ID != 1{
				return nil, store.ErrTaskNotFound
			}

			return &store.Task{
				ID: 1,
				Title: "Learn Go",
				Done: false,
			}, nil
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK{
		t.Fatalf("wanted success but got %d", rr.Code)
	}

	want := store.Task{
		ID: 1,
		Title: "Learn Go",
		Done: false,
	}

	var got store.Task

	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil{
		t.Fatal(err)
	}



	if !reflect.DeepEqual(got, want){
		t.Fatalf("wanted %v but got %v", want, got)
	}
}

func TestGetTaskByID_TaskNotFound(t *testing.T){
	
	fakeTaskService := &FakeTaskService{
		GetTaskByIDFn: func(gtbi service.GetTaskByIDInput) (*store.Task, error) {
			if gtbi.ID != 1{
				return nil, store.ErrTaskNotFound
			}

			return &store.Task{
				ID: 1,
				Title: "Learn Go",
				Done: false,
			}, nil
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/2", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound{
		t.Fatalf("wanted %d but got %d", http.StatusNotFound, rr.Code)
	}

}

func TestGetTaskByID_InvalidPath(t *testing.T){
	
	handler := NewTaskHandler(&FakeTaskService{})

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/abc", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest{
		t.Fatalf("wanted %d but got %d", http.StatusBadRequest, rr.Code)
	}

}

func TestGetTaskByID_Failed(t *testing.T){
	
	fakeTaskService := &FakeTaskService{
		GetTaskByIDFn: func(gtbi service.GetTaskByIDInput) (*store.Task, error) {
			return nil, store.ErrFetchTask
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks/{id}", handler.GetTaskByID)

	req := httptest.NewRequest(http.MethodGet, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("wanted %d but got %d", http.StatusInternalServerError, rr.Code)
	}

}

func TestUpdateTaskByID_Success(t *testing.T){
	fakeTaskService := &FakeTaskService{
		UpdateTaskByIDFn: func(utbi service.UpdateTaskByIDInput) (*store.Task, error) {
			return &store.Task{
				ID: utbi.ID,
				Title: utbi.Title,
				Done: utbi.Done,
			}, nil
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTaskByID)

	body := strings.NewReader(`{
		"title": "Hello Go",
		"done": true
	}`)

	req := httptest.NewRequest(http.MethodPut, "/tasks/1", body)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK{
		t.Fatalf("wanted success but got %d", rr.Code)
	}

	want := store.Task{
		ID: 1,
		Title: "Hello Go",
		Done: true,
	}

	var got store.Task

	if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil{
		t.Fatal(err)
	}


	if !reflect.DeepEqual(got, want){
		t.Fatalf("wanted %v but got %v", want, got)
	}

}

func TestUpdateTaskByID_AlreadyUpdated(t *testing.T){
	fakeTaskService := &FakeTaskService{
		UpdateTaskByIDFn: func(utbi service.UpdateTaskByIDInput) (*store.Task, error) {
			return nil, service.ErrTaskAlreadyCompleted
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTaskByID)

	body := strings.NewReader(`{
		"title": "Hello Go",
		"done": false
	}`)

	req := httptest.NewRequest(http.MethodPut, "/tasks/1", body)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict{
		t.Fatalf("wanted %d but got %d", http.StatusConflict, rr.Code)
	}

}

func TestUpdateTaskByID_TaskNotFound(t *testing.T){

	fakeTaskService := &FakeTaskService{
		CreateTaskFn: func(
			input service.CreateTaskInput,
		) (*store.Task, error) {

			return &store.Task{
				ID: 1,
				Title: input.Title,
				Done: true,
			}, nil
		},
		UpdateTaskByIDFn: func(utbi service.UpdateTaskByIDInput) (*store.Task, error) {
			if utbi.ID != 1{
				return nil, store.ErrTaskNotFound
			}
			return &store.Task{
				ID: utbi.ID,
				Title: utbi.Title,
				Done: utbi.Done,
			}, nil
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTaskByID)

	body := strings.NewReader(`{
		"title": "Hello Go",
		"done": true
	}`)

	req := httptest.NewRequest(http.MethodPut, "/tasks/2", body)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound{
		t.Fatalf("wanted %d but got %d", http.StatusNotFound, rr.Code)
	}

}

func TestUpdateTaskByID_Failed(t *testing.T){
	fakeTaskService := &FakeTaskService{
		UpdateTaskByIDFn: func(utbi service.UpdateTaskByIDInput) (*store.Task, error) {
			return nil, store.ErrUpdateTask
		},
	}

	handler := NewTaskHandler(fakeTaskService)

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTaskByID)

	body := strings.NewReader(`{
		"title": "Hello Go",
		"done": false
	}`)

	req := httptest.NewRequest(http.MethodPut, "/tasks/1", body)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("wanted %d but got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestUpdateTaskByID_InvalidPath(t *testing.T){

	handler := NewTaskHandler(&FakeTaskService{})

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTaskByID)

	body := strings.NewReader(`{
		"title": "Hello Go",
		"done": false
	}`)

	req := httptest.NewRequest(http.MethodPut, "/tasks/abc", body)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest{
		t.Fatalf("wanted %d but got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestCompleteAllTask_Success(t *testing.T){
	fakeTaskService := &FakeTaskService{
		CompleteAllTaskFn: func(ctx context.Context) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/tasks/complete-all", nil)
	rr := httptest.NewRecorder()

	handler := NewTaskHandler(fakeTaskService)

	handler.CompleteAllTask(rr, req)

	if rr.Code != http.StatusNoContent{
		t.Fatalf("wanted %d but got %d", http.StatusNoContent, rr.Code)
	}
}

func TestCompleteAllTask_Failed(t *testing.T){
	fakeTaskService := &FakeTaskService{
		CompleteAllTaskFn: func(ctx context.Context) error {
			return errors.New("transaction failed")
		},
	}

	req := httptest.NewRequest(http.MethodPut, "/tasks/complete-all", nil)
	rr := httptest.NewRecorder()

	handler := NewTaskHandler(fakeTaskService)

	handler.CompleteAllTask(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("wanted %d but got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteTaskByID_Success(t *testing.T){
	fakeTaskService := &FakeTaskService{
		DeleteTaskByIDFn: func(dtbi service.DeleteTaskByIDInput) error {
			return nil
		},
	}

	handler := NewTaskHandler(fakeTaskService)
	
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent{
		t.Fatalf("wanted %d but got %d", http.StatusNoContent, rr.Code)
	}

}

func TestDeleteTaskByID_TaskNotFound(t *testing.T){
	fakeTaskService := &FakeTaskService{
		DeleteTaskByIDFn: func(dtbi service.DeleteTaskByIDInput) error {
			return store.ErrTaskNotFound
		},
	}

	handler := NewTaskHandler(fakeTaskService)
	
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound{
		t.Fatalf("wanted %d but got %d", http.StatusNotFound, rr.Code)
	}

}

func TestDeleteTaskByID_Failed(t *testing.T){
	fakeTaskService := &FakeTaskService{
		DeleteTaskByIDFn: func(dtbi service.DeleteTaskByIDInput) error {
			return store.ErrDeleteTask
		},
	}

	handler := NewTaskHandler(fakeTaskService)
	
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/1", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError{
		t.Fatalf("wanted %d but got %d", http.StatusInternalServerError, rr.Code)
	}

}

func TestDeleteTaskByID_InvalidPath(t *testing.T){

	handler := NewTaskHandler(&FakeTaskService{})
	
	mux := http.NewServeMux()
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/abc", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest{
		t.Fatalf("wanted %d but got %d", http.StatusBadRequest, rr.Code)
	}

}
