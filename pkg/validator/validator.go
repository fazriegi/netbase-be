package validator

import (
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

var (
	digitRegex       = regexp.MustCompile(`\d`)
	specialCharRegex = regexp.MustCompile(`[^\w\s]`)
	lowerRegex       = regexp.MustCompile(`[a-z]`)
	upperRegex       = regexp.MustCompile(`[A-Z]`)
	tickerRegex      = regexp.MustCompile(`^[A-Za-z0-9.\-]{1,20}$`)
)

type ValidationErrResponse struct {
	FailedField string `json:"failed_field"`
	Tag         string `json:"tag"`
	TagValue    string `json:"tag_value"`
}

func ValidateRequest(data any) []ValidationErrResponse {
	var validationErrors []ValidationErrResponse

	validate := validator.New()
	validate.RegisterValidation("password", password)        // register custom validator
	validate.RegisterValidation("date", date)                // register custom validator
	validate.RegisterValidation("decimal", decimalValidator) // register custom validator
	validate.RegisterValidation("ticker", tickerValidator)   // register custom validator

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := fld.Tag.Get("json")
		if name == "-" {
			return ""
		}

		if idx := strings.Index(name, ","); idx != -1 {
			name = name[:idx]
		}

		if name == "" {
			name = fld.Tag.Get("query")
			if idx := strings.Index(name, ","); idx != -1 {
				name = name[:idx]
			}
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
		if len(password) < 8 || len(password) > 72 {
			return false
		}
		hasDigit := digitRegex.MatchString(password)
		hasSpecialChar := specialCharRegex.MatchString(password)
		hasLower := lowerRegex.MatchString(password)
		hasUpper := upperRegex.MatchString(password)

		return hasDigit && hasSpecialChar && hasLower && hasUpper
	}

	return true
}

func tickerValidator(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}
	return tickerRegex.MatchString(val)
}

func date(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if val == "" {
		return true
	}

	param := fl.Param()
	if param == "" {
		param = "YYYY-MM"
	}

	layout := param
	layout = strings.ReplaceAll(layout, "YYYY", "2006")
	layout = strings.ReplaceAll(layout, "MM", "01")
	layout = strings.ReplaceAll(layout, "DD", "02")

	t, err := time.Parse(layout, val)
	if err != nil {
		return false
	}

	if t.Year() < 1970 {
		return false
	}

	return true
}

func decimalValidator(fl validator.FieldLevel) bool {
	field := fl.Field()

	var d decimal.Decimal

	switch v := field.Interface().(type) {
	case decimal.Decimal:
		d = v
	case *decimal.Decimal:
		if v == nil {
			return true
		}
		d = *v
	case string:
		if v == "" {
			return true
		}
		parsed, err := decimal.NewFromString(v)
		if err != nil {
			return false
		}
		d = parsed
	default:
		if field.Kind() == reflect.String {
			str := field.String()
			if str == "" {
				return true
			}
			parsed, err := decimal.NewFromString(str)
			if err != nil {
				return false
			}
			d = parsed
		} else {
			return false
		}
	}

	param := fl.Param()
	switch param {
	case "gt=0", "positive":
		return d.GreaterThan(decimal.Zero)
	case "gte=0":
		return d.GreaterThanOrEqual(decimal.Zero)
	case "lt=0":
		return d.LessThan(decimal.Zero)
	case "lte=0":
		return d.LessThanOrEqual(decimal.Zero)
	default:
		return true
	}
}
