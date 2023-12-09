package stateful

import "context"

type (
	// Stateful is the core interface which should be implemented by all stateful structs.
	// If this interface is implemented by a struct it can be processed by the state machine
	Stateful interface {
		// GetID Get the id of the object
		GetID() string

		// State return current state
		State() State

		// SetState change the state to the given state
		SetState(ctx context.Context, state State, params Params) error
	}
)
