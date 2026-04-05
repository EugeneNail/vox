package validation

import (
	"fmt"
	"github.com/EugeneNail/vox/auth/internal/infrastructure/validation/rules"
)

// Validator applies field rulesets to input data.
type Validator struct {
	data     map[string]any
	rulesets map[string][]rules.Rule
}

// NewValidator constructs a validator for the provided data and rulesets.
func NewValidator(data map[string]any, rulesets map[string][]rules.Rule) *Validator {
	return &Validator{
		data:     data,
		rulesets: rulesets,
	}
}

// Validate executes rulesets and returns either validation violations or a rule execution error.
func (validator *Validator) Validate() error {
	validationError := NewError()

	for field, rls := range validator.rulesets {
	ruleLoop:
		for i, rule := range rls {
			message, err := rule(validator.data, field)
			if err != nil {
				return fmt.Errorf("applying %dth rule to field %q: %w", i, field, err)
			}

			if len(message) > 0 {
				validationError.AddViolation(field, message)
				break ruleLoop
			}
		}
	}

	if len(validationError.Violations()) > 0 {
		return validationError
	}

	return nil
}
