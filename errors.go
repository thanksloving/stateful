package stateful

import (
	"errors"
	"fmt"
)

var (
	funcNotFound      = errors.New("not found transfer function")
	batchFuncNotFound = errors.New("not found batch transfer function")
)

type TransitionNotFoundError struct {
	source      State
	destination State
}

func NewTransitionNotFoundError(source, destination State) *TransitionNotFoundError {
	return &TransitionNotFoundError{
		source:      source,
		destination: destination,
	}
}

func (trnfe TransitionNotFoundError) Error() string {
	return fmt.Sprintf(
		"no transition found source %s destination %s",
		trnfe.source, trnfe.destination,
	)
}
