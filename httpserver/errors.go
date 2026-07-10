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
	return fmt.Sprintf("[BadRequest] - %s", e.Message)
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
	return http.StatusNotFound
}

type ErrForbidden struct {
	Message string `json:"message"`
}

func (e ErrForbidden) Error() string {
	return fmt.Sprintf("[Forbidden] - %s", e.Message)
}
func (e ErrForbidden) HttpCode() int {
	return http.StatusForbidden
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
	case ErrForbidden:
		httpStatuscode = (err.(ErrForbidden)).HttpCode()
	default:
		httpStatuscode = http.StatusInternalServerError
	}
	return httpStatuscode
}

func HttpStatusCodeToError(statusCode int) error {
	var err error
	err = nil
	switch statusCode {
	case http.StatusForbidden:
		err = ErrForbidden{}
	case http.StatusBadRequest:
		err = ErrNotFound{}
	case http.StatusInternalServerError:
		err = ErrInternalServerError{}
	case http.StatusUnauthorized:
		err = ErrUnauthorized{}
	default:
		err = nil
	}
	return err
}

type ErrInternalServerError struct {
	Message string `json:"message"`
}

func (e ErrInternalServerError) Error() string {
	return fmt.Sprintf("[Internal Error] - %s", e.Message)
}
func (e ErrInternalServerError) HttpCode() int {
	return http.StatusUnauthorized
}

type ErrRedirect struct {
	RedirectUrl string
}

func (e ErrRedirect) Error() string {
	return fmt.Sprintf("redirect %s", e.RedirectUrl)
}
func (e ErrRedirect) HttpCode() int {
	return http.StatusTemporaryRedirect
}

type ErrPaymentRequired struct {
	Message string `json:"message"`
}

func (e ErrPaymentRequired) Error() string {
	return fmt.Sprintf("[Payment required] - %s", e.Message)
}
func (e ErrPaymentRequired) HttpCode() int {
	return http.StatusPaymentRequired
}
