package auth

import (
	"errors"
	"regexp"
	"unicode"

	"github.com/go-playground/validator/v10"
)

type SignUpValidator struct {
	validator *validator.Validate
}

func NewSignUpValidator() *SignUpValidator {
	v := validator.New()
	return &SignUpValidator{validator: v}
}

func PasswordValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(password) >= 7 {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}

		if hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial {
			return true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

type SignUpValidationErrors struct {
	NameError     string `json:"name"`
	LastNameError string `json:"last_name"`
	PhoneError    string `json:"phone"`
	EmailError    string `json:"email"`
	PasswordError string `json:"password"`
}

func isValidName(fl validator.FieldLevel) bool {
	// Expresión regular para validar nombres
	nameRegex := regexp.MustCompile(`^[a-zA-Z]+(([',. -][a-zA-Z ])?[a-zA-Z]*)*$`)
	return nameRegex.MatchString(fl.Field().String())
}

func (cv *SignUpValidator) ValidateSignUp(form *SignUpPayload) (SignUpValidationErrors, error) {
	cv.validator.RegisterValidation("password", PasswordValidation)
	cv.validator.RegisterValidation("validname", isValidName)

	err := cv.validator.Struct(form)
	if err != nil {
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

		customErrors := SignUpValidationErrors{
			NameError:     errorsMap["Name"],
			LastNameError: errorsMap["LastName"],
			PhoneError:    errorsMap["Phone"],
			EmailError:    errorsMap["Email"],
			PasswordError: errorsMap["Password"],
		}

		return customErrors, errors.New("validation errors")
	}

	return SignUpValidationErrors{}, nil
}
