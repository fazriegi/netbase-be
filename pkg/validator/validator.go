package validator

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationErrResponse struct {
	FailedField string `json:"failed_field"`
	Tag         string `json:"tag"`
	TagValue    string `json:"tag_value"`
}

func ValidateRequest(data any) []ValidationErrResponse {
	var validationErrors []ValidationErrResponse

	validate := validator.New()
	validate.RegisterValidation("password", password) // register custom validator

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}

		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}
		return name
	})

	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var validationErr ValidationErrResponse
			validationErr.FailedField = err.Field()
			validationErr.Tag = err.Tag()
			validationErr.TagValue = err.Param()

			validationErrors = append(validationErrors, validationErr)
		}
	}

	return validationErrors
}

func password(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if password != "" {
		hasDigit := regexp.MustCompile(`\d`).MatchString(password)
		hasSpecialChar := regexp.MustCompile(`[^\w\s]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)

		return hasDigit && hasSpecialChar && hasLower && hasUpper && len(password) >= 8
	}

	return true
}
