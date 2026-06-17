package main

import (
	"database/sql"
	"fmt"
	"go_crud/auth"
	"go_crud/handler"
	"go_crud/service"
	"go_crud/store"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	_ "github.com/lib/pq"
)


func main() {

	fmt.Println("DB_HOST:", os.Getenv("DB_HOST"))
	fmt.Println("DB_PORT:", os.Getenv("DB_PORT"))
	fmt.Printf("DB_USER=%q\n", os.Getenv("DB_USER"))
	fmt.Printf("DB_PASSWORD=%q\n", os.Getenv("DB_PASSWORD"))

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	fmt.Printf("%s", connStr)

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"})
	

	taskStore := store.NewTaskStore(
		db,
	)

	redisStore := store.NewRedisStore(
		rdb,
	)

	taskService := service.NewTaskService(
		taskStore,
		redisStore,
	)

	taskHandler := handler.NewTaskHandler(
		taskService,
	)

	userStore := store.NewUserStore(db)

	userService := service.NewUserService(userStore)

	userHandler := handler.NewUserHandler(userService)





	r := chi.NewRouter()

	r.Route("/tasks", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/", taskHandler.GetTasks)
		r.Get("/{id}", taskHandler.GetTaskByID)
		r.Post("/", taskHandler.CreateTask)
		r.Put("/{id}", taskHandler.UpdateTaskByID)
		r.Put("/complete-all", taskHandler.CompleteAllTask)
		r.With(auth.AdminRoleMiddleware).Delete("/{id}", taskHandler.DeleteTask)
	})

	r.Get("/ping", taskHandler.HealthCheck)

	r.Post("/register", userHandler.RegisterUser)
	r.Post("/login", userHandler.LoginUser)

	r.With(auth.AuthMiddleware).Get("/profile", userHandler.Profile)

	


	if err := http.ListenAndServe(":8000", r); err != nil {
		panic(err)
	}
}
