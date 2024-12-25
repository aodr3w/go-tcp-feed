package db

import "fmt"

type UserNotFound struct {
	UserName string
}

func (e *UserNotFound) Error() string {
	return fmt.Sprintf("user not found: %s", e.UserName)
}

func (e *UserNotFound) Is(target error) bool {
	_, ok := target.(*UserNotFound)
	return ok
}

type UserNameNotAvailable struct {
	UserName string
}

func (e *UserNameNotAvailable) Error() string {
	return fmt.Sprintf("Username already taken : %s", e.UserName)
}

func (e *UserNameNotAvailable) Is(target error) bool {
	_, ok := target.(*UserNameNotAvailable)
	return ok
}

var UserNotFoundError = UserNotFound{}
var UserNameNotAvailableError = UserNameNotAvailable{}
