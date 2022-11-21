package httpserver

import (
	"fmt"
	"net/http"
)

type ErrUnauthorized struct {
	Message string `json:"message"`
}

func (e ErrUnauthorized) Error() string {
	return fmt.Sprintf("[Unauthorized] - %s", e.Message)
}
func (e ErrUnauthorized) HttpCode() int {
	return http.StatusUnauthorized
}

type ErrBadRequest struct {
	Message string `json:"message"`
}

func (e ErrBadRequest) Error() string {
	return fmt.Sprintf("[Unauthorized] - %s", e.Message)
}
func (e ErrBadRequest) HttpCode() int {
	return http.StatusBadRequest
}

type ErrNotFound struct {
	Message string `json:"message"`
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("[not found] - %s", e.Message)
}
func (e ErrNotFound) HttpCode() int {
	return http.StatusBadRequest
}

func errorToHttpStatuscode(err error) int {
	httpStatuscode := 500
	switch err.(type) {
	case ErrUnauthorized:
		httpStatuscode = (err.(ErrUnauthorized)).HttpCode()
	case ErrBadRequest:
		httpStatuscode = (err.(ErrBadRequest)).HttpCode()
	case ErrNotFound:
		httpStatuscode = (err.(ErrNotFound)).HttpCode()
	default:
		httpStatuscode = http.StatusInternalServerError
	}
	return httpStatuscode
}
