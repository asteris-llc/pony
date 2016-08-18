package cli

import (
	"strconv"
)

func checkIsNumeric(input string) error {
	if _, err := strconv.Atoi(input); err != nil {
		return new(NumericError)
	}

	return nil
}

type NumericError struct {}

func (e *NumericError) Error() string { return "" }

func isNotNumeric(err error) bool {
	switch err.(type) {
	case *NumericError:
		return true
	}

	return false
}
