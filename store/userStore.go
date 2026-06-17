package store

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

type User struct {
	ID int `json:"user_id"`
	Username string `json:"username"`
	Password string	`json:"password"`
	Role     string `json:"role"`
}

type UserStore struct {
	db *sql.DB
}

type CreateUserParams struct{
	Username string
	Password string
	Role string
}

type GetUserByIDParams struct{
	ID int
}

type GetUserByUsernameParams struct{
	Username string
}

func NewUserStore(db *sql.DB) *UserStore{
	return &UserStore{
		db: db,
	}
}


func (store *UserStore) CreateUser(p CreateUserParams) (*User, error){

	query := `
	INSERT INTO users(username, password, role)
	VALUES($1, $2, $3)
	RETURNING id
	`

	user := User{Username: p.Username, Password: p.Password, Role: p.Role}

	if err := store.db.QueryRow(query, p.Username, p.Password, p.Role).Scan(&user.ID); err != nil{
		return nil, ErrCreateUser
	}

	return &user, nil

}

func (store *UserStore) GetUserByID(p GetUserByIDParams) (*User, error){
	query := `
	SELECT username, password, role
	FROM users
	WHERE id = $1
	`

	user := User{ID: p.ID}

	if err := store.db.QueryRow(query, p.ID).Scan(&user.Username, &user.Password, &user.Role); err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, ErrUserNotFound
		}
		return nil, ErrFetchUser
	}

	return &user, nil
}

func (store *UserStore) GetUserByUsername(p GetUserByUsernameParams) (*User, error){
	query := `
	SELECT id, password, role
	FROM users
	WHERE username = $1
	`

	user := User{Username: p.Username}

	if err := store.db.QueryRow(query, p.Username).Scan(&user.ID, &user.Password, &user.Role); err != nil{
		if errors.Is(err, sql.ErrNoRows){
			return nil, ErrUserNotFound
		}
		return nil, ErrFetchUser
	}

	return &user, nil
}