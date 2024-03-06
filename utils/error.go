package utils

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/juliotorresmoreno/specialist-talk-api/logger"
)

var log = logger.SetupLogger()

func ParseErrors(err error) map[string]string {
	log.Error("Error validating user input", err)
	errorsMap := make(map[string]string)

	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		tag := err.Tag()

		switch tag {
		case "required":
			errorsMap[field] = "This field is required!"
		case "email":
			errorsMap[field] = "Invalid email format!"
		case "phone":
			errorsMap[field] = "Invalid phone number!"
		case "pattern":
			errorsMap[field] = "Password does not meet requirements!"
		default:
			errorsMap[field] = "Invalid field!"
		}
	}
	return errorsMap
}

type HttpResponse struct {
	Status int
	Obj    HttpError
}

type HttpError struct {
	Message string `json:"message"`
}

var StatusInternalServerError = &HttpResponse{
	Status: http.StatusInternalServerError,
	Obj:    HttpError{Message: "Internal Server Error"},
}

var StatusUnauthorized = &HttpResponse{
	Status: http.StatusUnauthorized,
	Obj:    HttpError{Message: "Unauthorized"},
}

var StatusBadRequest = &HttpResponse{
	Status: http.StatusBadRequest,
	Obj:    HttpError{Message: "Bad Request"},
}

var StatusNotFound = &HttpResponse{
	Status: http.StatusNotFound,
	Obj:    HttpError{Message: "Not Found"},
}

func (e *HttpResponse) Error() string {
	return e.Obj.Message
}

func Response(c *gin.Context, payload interface{}) {
	if payload == nil {
		c.Status(http.StatusNoContent)
		c.Abort()
		return
	}

	payloadValue := reflect.ValueOf(payload)

	if payloadValue.Kind() == reflect.Ptr {
		payloadValue = reflect.Indirect(payloadValue)
	}

	statusField := payloadValue.FieldByName("Status")
	objField := payloadValue.FieldByName("Obj")

	if !statusField.IsValid() || !statusField.CanInterface() || !objField.IsValid() || !objField.CanInterface() {
		c.JSON(200, payload)
		return
	}

	statusValue := statusField.Interface().(int)
	objValue := objField.Interface()

	c.Header("Content-Type", "application/json")
	c.JSON(statusValue, objValue)
}
