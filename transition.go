package stateful

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type (
	// Transition represents the transition function which will be executed if the order is in the proper state
	// and there is a valid transitionRule in the state machine
	Transition interface {
		GetID() string
		GetSourceStates() States
		GetDestinationState() State
		IsAllowedToRun(state State) bool
		IsAllowedToTransfer(state State) bool
		Transfer(ctx context.Context, statefulObject Stateful, params Params) error
		BatchTransfer(ctx context.Context, statefulObjects []Stateful, params Params) error
	}

	defaultTransition struct {
		id          string
		sources     States
		destination State
		fn          func(ctx context.Context, obj Stateful, params Params) error
		batchFn     func(ctx context.Context, objs []Stateful, params Params) error
	}

	// Transitions are a slice of Transition
	Transitions []Transition
)

func NewTransition(id string, sources States, destination State,
	fn func(context.Context, Stateful, Params) error,
	options ...TransitionOption) Transition {
	transition := &defaultTransition{
		id:          id,
		sources:     sources,
		destination: destination,
		fn:          fn,
	}
	for _, option := range options {
		option(transition)
	}
	return transition
}

type TransitionOption func(*defaultTransition)

func BatchTransfer(batchFn func(ctx context.Context, objs []Stateful, params Params) error) TransitionOption {
	return func(transition *defaultTransition) {
		transition.batchFn = batchFn
	}
}

func (tr *defaultTransition) GetID() string {
	return tr.id
}

func (tr *defaultTransition) GetSourceStates() States {
	return tr.sources
}

func (tr *defaultTransition) GetDestinationState() State {
	return tr.destination
}

func (tr *defaultTransition) IsAllowedToRun(state State) bool {
	return tr.sources.Contains(state) || tr.sources.HasWildCard()
}

func (tr *defaultTransition) IsAllowedToTransfer(state State) bool {
	// 目标状态不允许通配的情况
	return tr.destination.String() == state.String() //|| tr.destinationState.IsWildCard()
}

func (tr *defaultTransition) Transfer(ctx context.Context, statefulObj Stateful, params Params) error {
	if tr.fn != nil {
		return tr.fn(ctx, statefulObj, params)
	}
	if tr.batchFn != nil {
		return tr.BatchTransfer(ctx, []Stateful{statefulObj}, params)
	}
	return funcNotFound
}

func (tr *defaultTransition) BatchTransfer(ctx context.Context, statefulList []Stateful, params Params) error {
	if tr.batchFn != nil {
		return tr.batchFn(ctx, statefulList, params)
	}
	if tr.fn != nil {
		for _, statefulObj := range statefulList {
			if err := tr.Transfer(ctx, statefulObj, params); err != nil {
				return err
			}
		}
	}
	return batchFuncNotFound
}

func (ts Transitions) Contains(t Transition) bool {
	for _, currentTransition := range ts {
		if currentTransition.GetID() == t.GetID() {
			return true
		}
	}
	return false
}

func (ts Transitions) getAllStates() States {
	var states States
	keys := make(map[State]struct{})

	for _, transition := range ts {
		for _, state := range append(transition.GetSourceStates(), transition.GetDestinationState()) {
			if _, ok := keys[state]; !ok {
				keys[state] = struct{}{}
				if !state.IsWildCard() {
					states = append(states, state)
				}
			}
		}
	}
	return states
}

func (ts Transitions) Export(outfile string) error {
	dot := `digraph StateMachine {
	rankdir=LR
	node[width=1 fixedsize=true shape=circle style=filled ]
	
	`
	for _, transition := range ts {
		allStates := ts.getAllStates()
		for _, state := range transition.GetSourceStates() {
			if !state.IsWildCard() {
				link := fmt.Sprintf(`%s -> %s [label="%s"]`, state, transition.GetDestinationState(), transition.GetID())
				dot = dot + "\r\n" + link
			} else {
				for _, st := range allStates {
					link := fmt.Sprintf(`%s -> %s [label="%s"]`, st, transition.GetDestinationState(), transition.GetID())
					dot = dot + "\r\n" + link
				}
			}
		}
	}
	dot = dot + "\r\n}"
	cmd := fmt.Sprintf("dot -o%s -T%s -K%s -s%s %s", outfile, "png", "dot", "72", "-Gsize=10,5 -Gdpi=200")

	return system(cmd, dot)
}

func system(c string, dot string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command(`cmd`, `/C`, c)
	} else {
		cmd = exec.Command(`/bin/sh`, `-c`, c)
	}
	cmd.Stdin = strings.NewReader(dot)
	return cmd.Run()
}
