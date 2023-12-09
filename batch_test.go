package stateful

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBatchTransfer(t *testing.T) {
	list := []Stateful{
		&TestStatefulObject{
			state: State1,
		},
	}
	sm := NewStateMachine(Transitions{
		NewTransition("a0", []State{State2}, State3, nil, BatchTransfer(func(ctx context.Context, stateful []Stateful, params Params) error {
			params.Set("success", true)
			return nil
		})),
		NewTransition("a1", []State{State1}, State2, nil, BatchTransfer(func(ctx context.Context, objs []Stateful, params Params) error {
			return nil
		})),
	})
	err := sm.BatchRun(context.Background(), list, State2, nil)
	assert.NoError(t, err)

	assert.Equal(t, State2, list[0].State())

	list = append(list, &TestStatefulObject{state: State1})
	err = sm.BatchRun(context.Background(), list, State2, nil)

	assert.Errorf(t, err, "the stateful object state is not the same. want State2 got State1")

	list[1].SetState(context.Background(), State2, nil)
	var params = make(DefaultParams)
	err = sm.BatchRun(context.Background(), list, State3, params)

	assert.NoError(t, err)
	assert.Equal(t, list[0].State(), State3)
	assert.Equal(t, list[1].State(), State3)
	res, ok := params.Get("success")
	assert.True(t, ok)
	assert.True(t, res.(bool))

}
