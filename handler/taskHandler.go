package handler

import (
	"context"
	"encoding/json"
	"errors"
	"go_crud/service"
	"go_crud/store"
	"net/http"
	"strconv"
	"strings"
)


type TaskService interface {

	CreateTask(service.CreateTaskInput) (*store.Task, error)
	GetTasks(service.GetTasksInput) ([]*store.Task, error)
	GetTaskByID(service.GetTaskByIDInput) (*store.Task, error)
	UpdateTaskByID(service.UpdateTaskByIDInput) (*store.Task, error)
	DeleteTaskByID(service.DeleteTaskByIDInput) error
	CompleteAllTask(ctx context.Context) error
}


type TaskHandler struct{
	service TaskService
}

type HealthCheck struct{
	Message string
}

type CreateTaskRequest struct{
	Title string `json:"title"`
}

type UpdateTaskByIDRequest struct{
	Title string `json:"title"`
	Done bool `json:"done"`
}

func NewTaskHandler(service TaskService) *TaskHandler{
	return &TaskHandler{
		service: service,
	}
}

func (h *TaskHandler) HealthCheck(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	message := HealthCheck{Message: "pong"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request){
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	doneStr := r.URL.Query().Get("done")
	sortBy := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	w.Header().Set("Content-Type", "application/json")

	page := 1
	limit := 10

	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)

		if err != nil {
			http.Error(w, "invalid page paramter", http.StatusBadRequest)
			return
		}

		page = parsedPage
	}

	if page < 1 {
		page = 1
	}

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)

		if err != nil {
			http.Error(w, "invalid limit parameter", http.StatusBadRequest)
			return
		}

		limit = parsedLimit

	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	if doneStr != ""{
		_, err := strconv.ParseBool(doneStr)

		if err != nil{
			http.Error(w, "invalid done value", http.StatusBadRequest)
			return
		}

	}

	if order == ""{
		order = "ASC"
	}

	

	allowedSortFields := map[string]string{
		"id":         "id",
		"title":      "title",
		"created_at": "created_at",
	}

	sortColumn := "id"

	if sortBy != "" {
		column, exists := allowedSortFields[sortBy]

		if !exists {
			http.Error(w, "invalid sort field", http.StatusBadRequest)
			return
		}

		sortColumn = column
	}

	order = strings.ToUpper(order)

	if order != "ASC" && order != "DESC"{
		http.Error(w, "invalid order type", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	tasks, err := h.service.GetTasks(
		service.GetTasksInput{
			Limit: limit,
			Offset: offset,
			Filter: doneStr,
			SortBy: sortColumn,
			Order: order,
			Ctx: ctx,
		},
	)

	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(tasks); err != nil{
		http.Error(w, ErrEncodingResponse.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *TaskHandler) GetTaskByID(w http.ResponseWriter, r *http.Request){
	idStr := r.PathValue("id")
	w.Header().Set("Content-Type", "application/json")

	if idStr == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	task, err := h.service.GetTaskByID(
		service.GetTaskByIDInput{
			ID: id,
			Ctx: ctx,
		},
	)

	if err != nil{
		if errors.Is(err, store.ErrTaskNotFound){
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil{
		http.Error(w, ErrEncodingResponse.Error(), http.StatusInternalServerError)
	}

}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var task CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, ErrDecodingResponse.Error(), http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(task.Title) == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	t, err := h.service.CreateTask(
		service.CreateTaskInput{
			Title: task.Title,
			Ctx: ctx,
		},
	)

	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(t); err != nil{
		http.Error(w, ErrEncodingResponse.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) UpdateTaskByID(w http.ResponseWriter, r *http.Request){
	defer r.Body.Close()
	
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}


	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")

	if idStr == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	var updatedTask UpdateTaskByIDRequest
	if err := json.NewDecoder(r.Body).Decode(&updatedTask); err != nil {
		http.Error(w, ErrDecodingResponse.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	task, err := h.service.UpdateTaskByID(
		service.UpdateTaskByIDInput{
			ID: id,
			Title: updatedTask.Title,
			Done: updatedTask.Done,
			Ctx: ctx,
		},
	)

	if err != nil{
		if errors.Is(err, store.ErrTaskNotFound){
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if errors.Is(err, service.ErrTaskAlreadyCompleted){
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(task); err != nil{
		http.Error(w, ErrEncodingResponse.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TaskHandler) CompleteAllTask(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPut {
		http.Error(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	err := h.service.CompleteAllTask(ctx)

	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request){
	
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	idStr := r.PathValue("id")

	if idStr == "" {
		http.Error(w, "id parameter is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)

	if err != nil {
		http.Error(w, "invalid id parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	err = h.service.DeleteTaskByID(
		service.DeleteTaskByIDInput{
			ID: id,
			Ctx: ctx,
		},
	)

	if err != nil{
		if errors.Is(err, store.ErrTaskNotFound){
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
