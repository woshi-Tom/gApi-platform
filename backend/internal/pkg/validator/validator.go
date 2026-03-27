package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
	emailRe  = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	tokenRe  = regexp.MustCompile(`^sk-ap-[a-f0-9]{32}$`)
)

func init() {
	// Register custom validators
	validate.RegisterValidation("email_format", validateEmail)
	validate.RegisterValidation("token_format", validateTokenKey)
	validate.RegisterValidation("password_strength", validatePasswordStrength)
}

// Validate validates a struct
func Validate(s interface{}) error {
	return validate.Struct(s)
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	return emailRe.MatchString(email)
}

// ValidateTokenKey validates API token key format
func ValidateTokenKey(key string) bool {
	return tokenRe.MatchString(key)
}

func validateEmail(fl validator.FieldLevel) bool {
	email := fl.Field().String()
	return emailRe.MatchString(email)
}

func validateTokenKey(fl validator.FieldLevel) bool {
	key := fl.Field().String()
	if key == "" {
		return true // optional field
	}
	return tokenRe.MatchString(key)
}

func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 8 {
		return false
	}
	// At least one letter and one number
	hasLetter := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasNumber := strings.ContainsAny(password, "0123456789")
	return hasLetter && hasNumber
}

// GetValidationErrors returns formatted validation errors
func GetValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				errors[field] = field + " is required"
			case "email":
				errors[field] = "invalid email format"
			case "min":
				errors[field] = field + " must be at least " + fe.Param() + " characters"
			case "max":
				errors[field] = field + " must be at most " + fe.Param() + " characters"
			default:
				errors[field] = field + " is invalid"
			}
		}
	}

	return errors
}
