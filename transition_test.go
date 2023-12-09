package stateful

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func transitiontestA(_ context.Context, _ Stateful, _ Params) error {
	return nil
}

func TestTransition_IsAllowedToRun(t *testing.T) {
	transitionRule := NewTransition(
		"",
		States{
			DefaultState("transitionTest_a"),
			DefaultState("transitionTest_b"),
		},
		DefaultState("transitionTest_d"), nil)
	assert.True(t, transitionRule.IsAllowedToRun(DefaultState("transitionTest_a")))
	assert.True(t, transitionRule.IsAllowedToRun(DefaultState("transitionTest_b")))
	assert.False(t, transitionRule.IsAllowedToRun(DefaultState("transitionTest_c")))
}

func TestTransition_IsAllowedToTransfer(t *testing.T) {
	transitionRule := NewTransition(
		"",
		States{
			DefaultState("transitionTest_a"),
			DefaultState("transitionTest_b"),
		},
		DefaultState("transitionTest_d"), nil)
	assert.False(t, transitionRule.IsAllowedToTransfer(DefaultState("transitionTest_c")))
	assert.True(t, transitionRule.IsAllowedToTransfer(DefaultState("transitionTest_d")))
}

func TestTransitionRules_Find(t *testing.T) {
	testObj := &TestStatefulObject{
		DefaultState("a"),
		1,
	}
	sm := &stateMachineImpl{}
	sm.AddTransition(NewTransition("a1", States{DefaultState("a")}, DefaultState("b"), transitiontestA))
	sm.AddTransition(NewTransition("a2", States{DefaultState("b")}, DefaultState("c"), transitiontestA))
	assert.Equal(
		t,
		"a1",
		sm.findTransition(DefaultState("a"), DefaultState("b")).GetID(),
	)
	_ = testObj.SetState(context.Background(), DefaultState("b"), nil)
	assert.Equal(
		t,
		nil,
		sm.findTransition(DefaultState("b"), DefaultState("a")),
	)
}
