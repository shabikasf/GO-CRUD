package service

import (
	"errors"
	"go_crud/store"
	"strings"
	"golang.org/x/crypto/bcrypt"
)

const (
    MinUsernameLength = 8
    MinPasswordLength = 8
)

type UserRepository interface{
	GetUserByID(store.GetUserByIDParams) (*store.User, error)
	GetUserByUsername(store.GetUserByUsernameParams) (*store.User, error)
	CreateUser(store.CreateUserParams) (*store.User, error)
}

type UserLoginInput struct{
	Username string
	Password string
}

type UserRegisterInput struct{
	Username string
	Password string
	Role string
}

type UserProfileInput struct{
	ID int
}

type UserService struct{
	store UserRepository
}

func NewUserService(store UserRepository) *UserService{
	return &UserService{
		store: store,
	}
}

func (service *UserService) Register(in UserRegisterInput) (*store.User, error){
	if strings.TrimSpace(in.Username) == ""{
		return nil, ErrInvalidUsername
	}

	if strings.TrimSpace(in.Password) == ""{
		return nil, ErrInavlidPassword
	}

	if len(in.Username) < MinUsernameLength{
		return nil, ErrUsernameLengthIsShort
	}

	if len(in.Password) < MinPasswordLength{
		return nil, ErrPasswordLengthIsShort
	}

	_, err := service.store.GetUserByUsername(store.GetUserByUsernameParams{Username: in.Username})

	if err == nil{
		return nil, ErrUsernameAlreadyExists
	}

	if !errors.Is(err, store.ErrUserNotFound){
		return nil, err
	}

	hashed_password, err := bcrypt.GenerateFromPassword(
		[]byte(in.Password),
		bcrypt.DefaultCost,
	)

	if err != nil{
		return nil, errors.New("error hashing password")
	}

	user, err := service.store.CreateUser(store.CreateUserParams{Username: in.Username, Password: string(hashed_password), Role: in.Role})

	if err != nil{
		return nil, err
	}

	return user, nil

}

func (service *UserService) Login(in UserLoginInput) (*store.User,error){
	if strings.TrimSpace(in.Username) == ""{
		return nil, ErrInvalidUsername
	}

	if strings.TrimSpace(in.Password) == ""{
		return nil, ErrInavlidPassword
	}

	if len(in.Username) < MinUsernameLength{
		return nil, ErrUsernameLengthIsShort
	}

	if len(in.Password) < MinPasswordLength{
		return nil, ErrPasswordLengthIsShort
	}

	user, err := service.store.GetUserByUsername(store.GetUserByUsernameParams{Username: in.Username})

	if err != nil{
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password),
		[]byte(in.Password),
	)

	if err != nil{
		return nil, ErrIncorrectUsernameOrPassword
	}

	return user, nil

}

func (service *UserService) Profile(in UserProfileInput) (*store.User, error){
	user, err := service.store.GetUserByID(store.GetUserByIDParams{ID: in.ID})

	if err != nil{
		return nil, err
	}

	return user, nil
}

