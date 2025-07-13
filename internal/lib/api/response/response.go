package response

import (
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
