package response

import (
	"errors"
	"expire-share/internal/domain/errors/services/auth"
	"expire-share/internal/domain/errors/services/files"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type Response struct {
	Errors []string `json:"errors,omitempty"`
}

func Error(errorMessages ...string) Response {
	return Response{Errors: errorMessages}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var validationErrorsMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "min":
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s must be greater or equal than %s", err.Field(), err.Param()))
		case "required":
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is required", err.Field()))
		case "url":
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is not a URL", err.Field()))
		default:
			validationErrorsMessages = append(validationErrorsMessages, fmt.Sprintf("%s is not a valid value", err.Field()))
		}
	}

	return Response{Errors: validationErrorsMessages}
}

func RenderError(w http.ResponseWriter, r *http.Request, statusCode int, errorMessage string) {
	render.Status(r, statusCode)
	render.JSON(w, r, Error(errorMessage))
}

func RenderValidationError(w http.ResponseWriter, r *http.Request, errors validator.ValidationErrors) {
	render.Status(r, http.StatusUnprocessableEntity)
	render.JSON(w, r, ValidationError(errors))
}

func RenderFileServiceError(w http.ResponseWriter, r *http.Request, err error) bool {
	if errors.Is(err, files.ErrAliasNotFound) {
		RenderError(w, r,
			http.StatusNotFound,
			"file with current alias not found")
		return true
	}

	if errors.Is(err, files.ErrPasswordRequired) {
		RenderError(w, r,
			http.StatusUnauthorized,
			"password is required")
		return true
	}

	if errors.Is(err, files.ErrIncorrectPassword) {
		RenderError(w, r,
			http.StatusForbidden,
			"incorrect password")
		return true
	}

	if errors.Is(err, files.ErrFileSizeTooBig) {
		RenderError(w, r,
			http.StatusUnprocessableEntity,
			"file size is very big")
		return true
	}

	return false
}

func RenderUserServiceError(w http.ResponseWriter, r *http.Request, err error) bool {
	if errors.Is(err, auth.ErrUserAlreadyExists) {
		RenderError(w, r,
			http.StatusConflict,
			"user with this login already exists")
		return true
	}

	if errors.Is(err, auth.ErrInvalidPassword) || errors.Is(err, auth.ErrUserNotFound) {
		RenderError(w, r,
			http.StatusUnauthorized,
			"login or password is invalid")
		return true
	}

	return false
}
