package cli

import (
	"regexp"
)

func checkNotEmpty(input string) error {
	m, err := regexp.MatchString("^[[:blank:]]*$", input)
	if err != nil {
		return err
	}

	if m {
		return new(EmptyError)
	}

	return nil
}

type EmptyError struct{}

func (e *EmptyError) Error() string { return "" }

func isEmpty(err error) bool {
	switch err.(type) {
	case *EmptyError:
		return true
	}

	return false
}
