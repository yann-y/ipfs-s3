package gerror

import (
	"fmt"
)

type GError struct {
	Cause error
}

func NewGError(err error) error {
	return &GError{
		Cause: err,
	}
}

func (err *GError) Error() string {
	return fmt.Sprintf("%s", err.Cause)
}
