package validators

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func Init() {
	Validate = validator.New()
	Validate.RegisterValidation("isbn", isbnValidator)
	Validate.RegisterValidation("shelf", shelfValidator)
}

func isbnValidator(fl validator.FieldLevel) bool {
	isbn := fl.Field().String()
	if isbn == "" {
		return true 
	}

	clean := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")
	return len(clean) == 10 || len(clean) == 13
}

func shelfValidator(fl validator.FieldLevel) bool {
	shelf := fl.Field().String()
	if shelf == "" {
		return true
	}
	match, _ := regexp.MatchString(`^[A-Z]-\d{1,3}$`, shelf)
	return match
}