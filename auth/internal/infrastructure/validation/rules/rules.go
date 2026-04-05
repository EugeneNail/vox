package rules

import (
	"fmt"
	"reflect"
	"regexp"
	"unicode"
)

// Rule validates a field within the full data payload.
type Rule func(data map[string]any, field string) (string, error)

// LettersPattern allows only Latin and Cyrillic letters.
const LettersPattern = `^[\p{Latin}\p{Cyrillic}]+$`

// AlphaNumericPattern allows only Latin and Cyrillic letters with integer digits.
const AlphaNumericPattern = `^[\p{Latin}\p{Cyrillic}\d]+$`

// SlugPattern allows slug-like text with Latin and Cyrillic letters, integer digits, hyphens, and dots.
const SlugPattern = `^[\p{Latin}\p{Cyrillic}\d.-]+$`

// SlugWithSpacesPattern allows slug-like text with Latin and Cyrillic letters, integer digits, hyphens, dots, and spaces.
const SlugWithSpacesPattern = `^[\p{Latin}\p{Cyrillic}\d. -]+$`

// SentencePattern allows sentence-like text with letters, digits, punctuation, and spaces.
const SentencePattern = `^[\p{L}\p{N}\p{P}\p{Zs}]+$`

// EmailPattern allows a typical email address with a local part, @, and a domain part with a dot-separated suffix.
const EmailPattern = `^[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}$`

// Required validates that the field exists, is not nil, and differs from its type zero value.
func Required() Rule {
	return func(data map[string]any, field string) (string, error) {
		value, exists := data[field]
		if !exists || isNil(value) || isZeroValue(value) {
			return "is required", nil
		}

		return "", nil
	}
}

// Min validates that the field value or size is at least the provided minimum.
func Min(minimum int) Rule {
	return func(data map[string]any, field string) (string, error) {
		value, exists := data[field]
		if !exists {
			return "", nil
		}

		current, err := extractMeasurableValue(value)
		if err != nil {
			return "", fmt.Errorf("measuring minimum for field %q: %w", field, err)
		}

		if current < float64(minimum) {
			return fmt.Sprintf("must be at least %d", minimum), nil
		}

		return "", nil
	}
}

// Max validates that the field value or size does not exceed the provided maximum.
func Max(maximum int) Rule {
	return func(data map[string]any, field string) (string, error) {
		value, exists := data[field]
		if !exists {
			return "", nil
		}

		current, err := extractMeasurableValue(value)
		if err != nil {
			return "", fmt.Errorf("measuring maximum for field %q: %w", field, err)
		}

		if current > float64(maximum) {
			return fmt.Sprintf("must be at most %d", maximum), nil
		}

		return "", nil
	}
}

// Regex validates that the field string matches the provided regular expression mask.
func Regex(mask string) Rule {
	pattern, err := regexp.Compile(mask)

	return func(data map[string]any, field string) (string, error) {
		if err != nil {
			return "", fmt.Errorf("compiling regex mask %q: %w", mask, err)
		}

		value, exists := data[field]
		if !exists {
			return "", nil
		}

		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("matching regex for field %q: type %T is not supported", field, value)
		}

		if !pattern.MatchString(text) {
			return "has invalid format", nil
		}

		return "", nil
	}
}

// Same validates that the field value matches the value of another field.
func Same(otherField string) Rule {
	return func(data map[string]any, field string) (string, error) {
		value, exists := data[field]
		if !exists {
			return "", nil
		}

		otherValue, otherExists := data[otherField]
		if !otherExists {
			return "", fmt.Errorf("reading field %q for comparison with field %q: field does not exist", otherField, field)
		}

		if value != otherValue {
			return fmt.Sprintf("must match %s", otherField), nil
		}

		return "", nil
	}
}

// Password validates that the field contains at least one digit, one letter,
// one lowercase letter, one uppercase letter, and one special symbol.
func Password() Rule {
	return func(data map[string]any, field string) (string, error) {
		value, exists := data[field]
		if !exists {
			return "", nil
		}

		text, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("reading password for field %q: type %T is not supported", field, value)
		}

		var hasDigit bool
		var hasLetter bool
		var hasLower bool
		var hasUpper bool
		var hasSpecial bool

		for _, symbol := range text {
			switch {
			case unicode.IsDigit(symbol):
				hasDigit = true
			case unicode.IsLetter(symbol):
				hasLetter = true
				if unicode.IsLower(symbol) {
					hasLower = true
				}
				if unicode.IsUpper(symbol) {
					hasUpper = true
				}
			default:
				hasSpecial = true
			}
		}

		switch {
		case !hasDigit:
			return "must contain at least one digit", nil
		case !hasLetter:
			return "must contain at least one letter", nil
		case !hasLower:
			return "must contain at least one lowercase letter", nil
		case !hasUpper:
			return "must contain at least one uppercase letter", nil
		case !hasSpecial:
			return "must contain at least one special symbol", nil
		default:
			return "", nil
		}
	}
}

func extractMeasurableValue(value any) (float64, error) {
	switch typedValue := value.(type) {
	case int:
		return float64(typedValue), nil
	case int8:
		return float64(typedValue), nil
	case int16:
		return float64(typedValue), nil
	case int32:
		return float64(typedValue), nil
	case int64:
		return float64(typedValue), nil
	case uint:
		return float64(typedValue), nil
	case uint8:
		return float64(typedValue), nil
	case uint16:
		return float64(typedValue), nil
	case uint32:
		return float64(typedValue), nil
	case uint64:
		return float64(typedValue), nil
	case uintptr:
		return float64(typedValue), nil
	case float32:
		return float64(typedValue), nil
	case float64:
		return typedValue, nil
	case string:
		return float64(len(typedValue)), nil
	}

	reflectedValue := reflect.ValueOf(value)
	if reflectedValue.Kind() == reflect.Array || reflectedValue.Kind() == reflect.Slice || reflectedValue.Kind() == reflect.Map {
		return float64(reflectedValue.Len()), nil
	}

	return 0, fmt.Errorf("type %T is not supported", value)
}

func isNil(value any) bool {
	if value == nil {
		return true
	}

	reflectedValue := reflect.ValueOf(value)
	switch reflectedValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflectedValue.IsNil()
	default:
		return false
	}
}

func isZeroValue(value any) bool {
	if value == nil {
		return true
	}

	switch typedValue := value.(type) {
	case bool:
		return !typedValue
	case string:
		return typedValue == ""
	case int:
		return typedValue == 0
	case int8:
		return typedValue == 0
	case int16:
		return typedValue == 0
	case int32:
		return typedValue == 0
	case int64:
		return typedValue == 0
	case uint:
		return typedValue == 0
	case uint8:
		return typedValue == 0
	case uint16:
		return typedValue == 0
	case uint32:
		return typedValue == 0
	case uint64:
		return typedValue == 0
	case uintptr:
		return typedValue == 0
	case float32:
		return typedValue == 0
	case float64:
		return typedValue == 0
	}

	return reflect.ValueOf(value).IsZero()
}
