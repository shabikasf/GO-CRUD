package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


type TaskStore struct {
	db *sql.DB
}

type CreateTaskParams struct{
	Ctx context.Context
	Title string
}

type GetTasksParams struct{
	Ctx context.Context
	Limit int
	Offset int
	Filter string
	SortBy string
	Order string
}

type GetTaskByIDParams struct{
	Ctx context.Context
	ID int
}

type UpdateTaskByIDParams struct{
	Ctx context.Context
	ID int
	Title string
	Done bool
}

type DeleteTaskByIDParams struct{
	Ctx context.Context
	ID int
}

func NewTaskStore(db *sql.DB) *TaskStore{
	return &TaskStore{db: db}
}

func (store *TaskStore) CreateTask(c *CreateTaskParams) (*Task, error){
	query := `
	INSERT INTO tasks(title, done, created_at, updated_at) 
	VALUES($1, $2, NOW(), NOW())
	RETURNING id, created_at, updated_at
	`

	task := Task{Title: c.Title, Done: false}

	if err := store.db.QueryRowContext(
		c.Ctx,
		query,
		task.Title,
		task.Done,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt); err != nil{
		return nil, ErrCreateTask
	}

	return &task, nil
}

func (store *TaskStore) GetTasks(g *GetTasksParams) ([]*Task, error){
	query := `
	SELECT id, title, done, created_at, updated_at
	FROM tasks
	`

	args := []interface{}{}
	placeHolderIndex := 1

	if g.Filter != ""{
		query += fmt.Sprintf(" WHERE done = $%d", placeHolderIndex)
		args = append(args, g.Filter)
		placeHolderIndex++
	}

	args = append(args, g.Limit, g.Offset)

	query += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", g.SortBy, g.Order, placeHolderIndex, placeHolderIndex+1)

	tasks := []*Task{}

	rows, err := store.db.QueryContext(g.Ctx, query, args...)

	if err != nil{
		return nil, ErrFetchTask
	}

	defer rows.Close()

	for rows.Next(){
		var task Task

		err := rows.Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt, &task.UpdatedAt)

		if err != nil{
			return nil, errors.New("failed to scan task")
		}

		tasks = append(tasks, &task)
	}

	if err := rows.Err(); err != nil{
		return nil, errors.New("row iteration error")
	}


	return tasks, nil

}

func (store *TaskStore) GetTaskByID(g *GetTaskByIDParams) (*Task, error){
	query := `
	SELECT id, title, done, created_at, updated_at
	FROM tasks
	WHERE id = $1
	`

	var task Task

	err := store.db.QueryRowContext(g.Ctx, query, g.ID).Scan(&task.ID, &task.Title, &task.Done, &task.CreatedAt, &task.UpdatedAt)

	if err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, errors.New("no task found")
		}

		return nil, ErrFetchTask
	}

	return &task, nil
}

func (store *TaskStore) UpdateTaskByID(u *UpdateTaskByIDParams)(*Task, error){
	query := `
	UPDATE tasks
	SET title = $1, done = $2
	WHERE id = $3
	`

	result, err := store.db.ExecContext(u.Ctx, query, u.Title, u.Done, u.ID)

	if err != nil{
		return  nil, ErrUpdateTask
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil{
		return nil, errors.New("failed to inspect task")
	}

	if rowsAffected == 0{
		return nil, ErrTaskNotFound
	}

	task, err := store.GetTaskByID(&GetTaskByIDParams{ID: u.ID})

	return task, err
}

func (store *TaskStore) CompleteAllTask(Ctx context.Context) error{
	query := `
	UPDATE tasks
	SET done = true,
		updated_at = NOW()
	`

	tx, err := store.db.Begin()

	if err != nil{
		return errors.New("error creating transaction")
	}

	defer tx.Rollback()

	result, err := tx.ExecContext(Ctx, query)

	if err != nil{
		return ErrUpdateTask
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil{
		return errors.New("failed to inspect task")
	}

	if rowsAffected == 0{
		return ErrTaskNotFound
	}

	if err := tx.Commit(); err != nil{
		return errors.New("error commiting transaction")
	}

	return nil


}

func (store *TaskStore) DeleteTaskByID(d *DeleteTaskByIDParams) error{
	query := `
	DELETE FROM tasks WHERE id = $1
	`

	result, err := store.db.ExecContext(d.Ctx, query, d.ID)

	if err != nil{
		return ErrDeleteTask
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil{
		return errors.New("failed to inspect task")
	}

	if rowsAffected == 0{
		return ErrTaskNotFound
	}

	return nil
}
