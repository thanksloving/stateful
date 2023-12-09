package stateful

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	State1 = DefaultState("State1")
	State2 = DefaultState("State2")
	State3 = DefaultState("State3")
	State4 = DefaultState("State4")
	State5 = DefaultState("State5")
)

type (
	TestStatefulObject struct {
		state     State
		TestValue int
	}

	TestParam struct {
		Amount int
	}
)

func (tsp *TestStatefulObject) GetID() string {
	return "_test_stateful_object"
}

func (tsp *TestStatefulObject) State() State {
	return tsp.state
}

func (tsp *TestStatefulObject) SetState(_ context.Context, state State, _ Params) error {
	tsp.state = state
	return nil
}

func (tsp *TestStatefulObject) FromState1ToState2(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func (tsp *TestStatefulObject) FromState2ToState3(_ context.Context, _ Stateful, params Params) error {
	para, _ := params.Get("param")
	testParam, _ := para.(TestParam)
	tsp.TestValue += testParam.Amount
	return nil
}

func (tsp *TestStatefulObject) FromState3ToState1(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func (tsp *TestStatefulObject) FromState2And3To4(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func (tsp *TestStatefulObject) FromState4ToState1(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func (tsp *TestStatefulObject) ErrorBehavior(_ context.Context, _ Stateful, _ Params) error {
	return errors.New("there was an error")
}

func (tsp TestStatefulObject) NotExistingTransition(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func NewTestStatefulObject() *TestStatefulObject {
	return &TestStatefulObject{state: State1}
}

func newStateMachine(testStatefulObject *TestStatefulObject) *stateMachineImpl {
	stateMachine := &stateMachineImpl{}
	stateMachine.AddTransition(NewTransition("1", States{State1}, State2, testStatefulObject.FromState1ToState2))
	stateMachine.AddTransition(NewTransition("2", States{State2}, State3, testStatefulObject.FromState2ToState3))
	stateMachine.AddTransition(NewTransition("3", States{State3}, State1, testStatefulObject.FromState3ToState1))
	stateMachine.AddTransition(NewTransition("4", States{State2, State3}, State4, testStatefulObject.FromState2And3To4))
	stateMachine.AddTransition(NewTransition("5", States{State4}, State5, testStatefulObject.FromState4ToState1))
	stateMachine.AddTransition(NewTransition("6", States{AllStates}, State1, testStatefulObject.ErrorBehavior))
	return stateMachine
}

func TestStateMachine_AddTransition(t *testing.T) {
	testStatefulObject := NewTestStatefulObject()

	stateMachine := stateMachineImpl{}
	stateMachine.AddTransition(NewTransition("1", States{State1}, State2, testStatefulObject.FromState1ToState2))
	stateMachine.AddTransition(NewTransition("2", States{State2, State4}, State3, testStatefulObject.ErrorBehavior))

	assert.ElementsMatch(
		t,
		States{State1},
		stateMachine.transitions[0].GetSourceStates(),
	)
	assert.Equal(
		t,
		State2,
		stateMachine.transitions[0].GetDestinationState(),
	)

	assert.ElementsMatch(
		t,
		States{State2, State4},
		stateMachine.transitions[1].GetSourceStates(),
	)
	assert.Equal(
		t,
		State3,
		stateMachine.transitions[1].GetDestinationState(),
	)
}

func TestStateMachine_GetAllStates(t *testing.T) {
	testStatefulObject := NewTestStatefulObject()
	stateMachine := stateMachineImpl{}
	stateMachine.AddTransition(NewTransition("1", States{State1}, State2, testStatefulObject.FromState1ToState2))
	stateMachine.AddTransition(NewTransition("2", States{State2, State4}, State3, testStatefulObject.ErrorBehavior))

	assert.ElementsMatch(
		t,
		States{
			State1,
			State2,
			State3,
			State4,
		},
		stateMachine.GetAllStates(),
	)
}

func TestStateMachine_Run(t *testing.T) {
	testStatefulObject := NewTestStatefulObject()
	stateMachine := newStateMachine(testStatefulObject)
	err := stateMachine.Run(
		context.Background(),
		testStatefulObject,
		State2,
		nil,
	)
	assert.NoError(t, err)
	assert.Equal(t, State2, testStatefulObject.State())

	err = stateMachine.Run(
		context.Background(),
		testStatefulObject,
		State3,
		DefaultParams{
			"param": TestParam{Amount: 2},
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, State3, testStatefulObject.State())
	assert.Equal(t, 2, testStatefulObject.TestValue)

	err = stateMachine.Run(
		context.Background(),
		testStatefulObject,
		State5,
		nil,
	)
	assert.Error(t, err)
	assert.Equal(
		t,
		reflect.TypeOf(&TransitionNotFoundError{}),
		reflect.TypeOf(err),
	)
	_ = testStatefulObject.SetState(context.Background(), State5, nil)
	err = stateMachine.Run(
		context.Background(),
		testStatefulObject,
		State1,
		nil,
	)
	assert.Error(t, err)
	assert.Equal(t, errors.New("there was an error"), err)
}

func TestStateMachine_GetAvailableTransitions(t *testing.T) {
	testStatefulObject := NewTestStatefulObject()
	stateMachine := newStateMachine(testStatefulObject)
	availableTransitions := stateMachine.GetAvailableTransitions(testStatefulObject)
	assert.Equal(
		t,
		2,
		len(availableTransitions),
	)
	allTransitions := stateMachine.GetAllStates()
	assert.Equal(
		t,
		5,
		len(allTransitions),
	)
}
