package domain

import "errors"

var (
	ErrEmptyUsername     = errors.New("username is required")
	ErrEmptyPassword     = errors.New("password is required")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

type User struct {
	Username string
	Password string
}

func NewUser(username, password string) (User, error) {
	if username == "" {
		return User{}, ErrEmptyUsername
	}
	if password == "" {
		return User{}, ErrEmptyPassword
	}

	return User{Username: username, Password: password}, nil
}
