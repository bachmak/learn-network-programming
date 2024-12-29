package robotmaid

import (
	"context"
	hw "learn-network-programming/housework"
	hwproto "learn-network-programming/housework/v1"
	"sync"
)

// struct Rosie
type Rosie struct {
	// mutex to synchronize the state to support multiple clients
	mutex sync.Mutex
	// chores
	chores hw.Chores
	// unimplemented server (required by the RobotMaidServer interface)
	hwproto.UnimplementedRobotMaidServer
}

// func Add
func (r *Rosie) Add(
	_ context.Context,
	chores *hwproto.Chores,
) (
	*hwproto.Response,
	error,
) {
	// sync state access
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// add chores
	for _, chore := range chores.Chores {
		r.chores.Add((*hw.Chore)(chore))
	}

	// response ok
	response := hwproto.Response{
		Message: "ok",
	}

	return &response, nil
}

// func Complete
func (r *Rosie) Complete(
	_ context.Context,
	req *hwproto.CompleteRequest,
) (
	*hwproto.Response,
	error,
) {
	// sync state access
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// get chore index
	idx := int(req.GetChoreNumber())
	// mark chore as completed
	err := r.chores.SetComplete(idx, true)
	if err != nil {
		return nil, err
	}

	// response ok
	response := hwproto.Response{
		Message: "ok",
	}

	return &response, nil
}

// func List
func (r *Rosie) List(
	_ context.Context,
	_ *hwproto.Empty,
) (
	*hwproto.Chores,
	error,
) {
	// sync state access
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// cast chores
	chores := (*hwproto.Chores)(&r.chores)

	return chores, nil
}
