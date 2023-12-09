# stateful
FSM
State machine

1: implement Stateful interface 
```
    type Stateful interface {
		// GetID Get the id of the object
		GetID() string

		// State return current state
		State() State

		// SetState change the state to the given state
		SetState(ctx context.Context, state State, params Params) error
	}
```
2: define the Transitions
```
    transitions := stateful.Transitions{
    		stateful.NewTransition("accept_ticket",
    			stateful.States{open},
    			accept,
    			func(ctx context.Context, ticket stateful.Stateful, params stateful.Params) error {
    				log.Infof("accept")
    				return nil
    			}),
    		stateful.NewTransition("transform_ticket",
    			stateful.States{accept},
    			close,
    			func(ctx context.Context, ticket stateful.Stateful, params stateful.Params) error {
    				log.Infof("close")
    				return nil
    			}),
    	}
```
3: Run the state machine to the new state
```
    obj := &models.Ticket{
                ID:     1,
                Status: open,
            },
    sm := stateful.NewStateMachine(
    		stateful.SetTransitions(transitions),
    )
    err := sm.Run(ctx, obj, accept, params)
```
4: Export the state graph
```
    // export the state machine with the transtitions
    sm.Graph("state.png")
```
