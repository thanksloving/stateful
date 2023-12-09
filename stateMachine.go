package stateful

import (
	"context"
	"errors"
	"fmt"
)

type (
	// StateMachine handles the state of the StatefulObject
	StateMachine interface {
		AddTransition(Transition)
		GetTransitions() Transitions
		GetAllStates() States
		Run(context.Context, Stateful, State, Params) error
		BatchRun(context.Context, []Stateful, State, Params) error
		GetAvailableTransitions(Stateful) Transitions
		Graph(outfile string) error
	}

	stateMachineImpl struct {
		transitions Transitions
	}

	// Params give the transitions parameters
	Params interface {
		Get(key string) (interface{}, bool)
		Set(key string, value interface{}) error
	}

	Option func(*stateMachineImpl)

	DefaultParams map[string]interface{}
)

var _ StateMachine = (*stateMachineImpl)(nil)

func NewStateMachine(transitions Transitions, options ...Option) *stateMachineImpl {
	sm := &stateMachineImpl{
		transitions: transitions,
	}
	for _, option := range options {
		option(sm)
	}
	return sm
}

func AddTransitions(transitions Transitions) Option {
	return func(machine *stateMachineImpl) {
		machine.transitions = append(machine.transitions, transitions...)
	}
}

// AddTransition adds a transition to the state machine.
func (sm *stateMachineImpl) AddTransition(t Transition) {
	sm.transitions = append(
		sm.transitions,
		t,
	)
}

// GetTransitions returns all transitions in the state machine
func (sm *stateMachineImpl) GetTransitions() Transitions {
	return sm.transitions
}

// GetAllStates returns all known and possible states by the state machine
func (sm *stateMachineImpl) GetAllStates() States {
	if len(sm.transitions) == 0 {
		return nil
	}
	return sm.transitions.getAllStates()
}

// Run runs the state machine with the given transition.
// If the transition
func (sm *stateMachineImpl) Run(
	ctx context.Context,
	stateful Stateful,
	newState State,
	params Params,
) error {
	f := NewFuture(func() error {
		currentState := stateful.State()
		transition := sm.findTransition(stateful.State(), newState)
		if transition == nil {
			return NewTransitionNotFoundError(currentState, newState)
		}
		// do transfer to the new state
		if err := transition.Transfer(ctx, stateful, params); err != nil {
			return err
		}
		// set the state
		if err := stateful.SetState(ctx, newState, params); err != nil {
			return err
		}
		return nil
	})
	return f.Get(ctx)
}

// Run runs the state machine with the given transition.
// If the transition
func (sm *stateMachineImpl) BatchRun(
	ctx context.Context,
	statefulList []Stateful,
	newState State,
	params Params,
) error {
	f := NewFuture(func() error {
		if len(statefulList) == 0 {
			return nil
		}
		var currentState = statefulList[0].State()
		for _, statefulObj := range statefulList {
			if statefulObj.State().String() != currentState.String() {
				return fmt.Errorf("the stateful object state is not the same. want %v got %v", currentState, statefulObj.State())
			}
		}
		transition := sm.findTransition(currentState, newState)
		if transition == nil {
			return NewTransitionNotFoundError(currentState, newState)
		}
		// do transfer to the new state
		if err := transition.BatchTransfer(ctx, statefulList, params); err != nil {
			return err
		}
		// set the state
		for _, statefulObj := range statefulList {
			if err := statefulObj.SetState(ctx, newState, params); err != nil {
				return err
			}
		}
		return nil
	})
	return f.Get(ctx)
}

// findTransition find the transition to the newState. return the first one
func (sm *stateMachineImpl) findTransition(currentState, newState State) Transition {
	for _, transition := range sm.transitions {
		// judge current state
		if !transition.IsAllowedToRun(currentState) {
			continue
		}
		// judge target state
		if transition.IsAllowedToTransfer(newState) {
			return transition
		}
	}
	return nil
}

// GetAvailableTransitions get all available transitions
func (sm *stateMachineImpl) GetAvailableTransitions(stateful Stateful) Transitions {
	var transitions Transitions
	for _, transition := range sm.transitions {
		if transition.IsAllowedToRun(stateful.State()) {
			transitions = append(transitions, transition)
		}
	}
	return transitions
}

// Graph exports the state diagram into a file.
func (sm *stateMachineImpl) Graph(outfile string) error {
	if len(sm.transitions) == 0 {
		return errors.New("can't find any transition")
	}
	return sm.transitions.Export(outfile)
}

func (p DefaultParams) Get(key string) (interface{}, bool) {
	if p == nil {
		return nil, false
	}
	val, ok := p[key]
	return val, ok
}

func (p DefaultParams) Set(key string, value interface{}) error {
	if p == nil {
		return nil
	}
	p[key] = value
	return nil
}
