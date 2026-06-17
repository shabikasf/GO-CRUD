package service

import (
	"errors"
)


var (
	ErrInvalidTitle = errors.New("title is required")
	ErrIDRequired = errors.New("id is required")
	ErrInvalidID = errors.New("invalid id type")
	ErrTaskAlreadyCompleted = errors.New("task already completed")

	ErrInvalidUsername = errors.New("username is invalid")
	ErrUsernameLengthIsShort = errors.New("username length is short")
	ErrUsernameAlreadyExists = errors.New("username already exists")

	ErrInavlidPassword = errors.New("password is invalid")
	ErrPasswordLengthIsShort = errors.New("password length is short")

	ErrIncorrectUsernameOrPassword = errors.New("username or password is incorrect")


)