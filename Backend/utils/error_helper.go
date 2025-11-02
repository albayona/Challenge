package utils

import "fmt"

// ErrorPanic panics if the error is not nil with a custom message
func ErrorPanic(err error, message string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", message, err))
	}
}
