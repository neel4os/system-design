package engine

import (
	"context"
)

type State map[string]any

// Node is the unit or work
type Node struct {
	// unique name of the node
	Name string
	// trigger is the channels whose version change wake up the node
	Triggers []string
	// Reads specifies the channels from which the node reads data
	Reads []string
	// Fn is the function that implements the node's logic
	Fn func(ctx context.Context, in State)([]Write, error)
}

type Write struct {
	// Write are the channels that the node writes to
	// note: writes are buffered until the next tick,
	// that means nodes write are not visible to other nodes until the next tick
	Channel string
	Value any
}

// task is scheduled execution of nodes in a tick
type Task struct {
	Node string
	In State
	SnapShot map[string]int64 // snapshot is channel version map which will eventually become seen
	Writes []Write
	Err error
}