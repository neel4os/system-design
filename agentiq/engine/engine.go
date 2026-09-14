package engine

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
)

const defaultStepLimit = 100

var (
	ErrNodeNameEmpty       = errors.New("node name can not be empty")
	ErrNodeNameDuplicated  = errors.New("node name can not be duplicated")
	ErrNodeFunctionMissing = errors.New("node function is missing")
	ErrChannelNotFound     = errors.New("channel not found")
)

type Option func(*Engine)

type TickInfo struct {
	Step  int
	Tasks []string
	State State
}

type Engine struct {
	nodes     map[string]*Node
	channels  map[string]Channel
	versions  map[string]int64            // channel -> versions
	seen      map[string]map[string]int64 // node -> channel -> versions
	clock     int64
	completed int
	limit     int
	hook      func(TickInfo)
}

func WithStepLimit(limit int) Option {
	return func(e *Engine) {
		e.limit = limit
	}
}

func WithHook(hook func(TickInfo)) Option {
	return func(e *Engine) {
		e.hook = hook
	}
}

func NewEngine(nodes []*Node, channels map[string]Channel, opts ...Option) (*Engine, error) {
	e := &Engine{
		nodes:    make(map[string]*Node, len(nodes)),
		channels: channels,
		versions: make(map[string]int64, len(channels)),
		seen:     make(map[string]map[string]int64, len(nodes)),
		limit:    defaultStepLimit,
	}
	for _, opt := range opts {
		opt(e)
	}

	// validation
	for _, node := range nodes {
		// nodes name can not be empty
		if node.Name == "" {
			return nil, ErrNodeNameEmpty
		}
		// nodes name can not be duplicated
		if _, ok := e.nodes[node.Name]; ok {
			return nil, ErrNodeNameDuplicated
		}
		// node must contain functions
		if node.Fn == nil {
			return nil, ErrNodeFunctionMissing
		}
		// read and trigger must be a valid channels
		for _, ch := range slices.Concat(node.Triggers, node.Reads) {
			if _, ok := channels[ch]; !ok {
				return nil, fmt.Errorf("%w: node %q references %q", ErrChannelNotFound, node.Name, ch)
			}
		}
		e.nodes[node.Name] = node
		e.seen[node.Name] = make(map[string]int64, len(node.Triggers)+len(node.Reads))
	}
	return e, nil
}

func (e *Engine) Run(ctx context.Context, input State) (State, error) {
	err := e.seed(input)
	if err != nil {
		return input, err
	}
	for e.completed < e.limit {
		tasks := e.plan()
		if len(tasks) == 0 {
			return e.State(), nil
		}
	}
	return nil, nil
}

func (e *Engine) State() State {
	s := make(State, len(e.channels))
	for name, ch := range e.channels{
		s[name] = ch.Get()
	}
	return s
}

func (e *Engine) triggered(name string, n *Node) bool {
	// A node is triggered only when its seen and actual engine version is different
	// for all triggers of a node
	for _, t := range n.Triggers {
		if e.versions[t] > e.seen[name][t] {
			return true
		}
	}
	return false

}

func (e *Engine) plan() []*Task {
	var tasks []*Task
	//for all node names
	for _, name := range slices.Sorted(maps.Keys(e.nodes)) {
		n := e.nodes[name]
		if !e.triggered(name, n) {
			continue
		}
		in := make(State, len(n.Reads))
		for _, r := range n.Reads {
			in[r] = e.channels[r].Get()
		}
		tasks = append(tasks, &Task{
			Node:     name,
			In:       in,
			SnapShot: maps.Clone(e.versions),
		})
	}
	return tasks

}

func (e *Engine) seed(input State) error {
	// it does initiation
	// 1. for all keys in State it check if corresponding channels are there or not
	// 2. if present then it sent to apply for each channels
	// apply will update the first write and increase the clock and version
	for _, name := range slices.Sorted(maps.Keys(input)) {
		ch, ok := e.channels[name]
		if !ok {
			return errors.New("unknown channel" + name)
		}
		err := e.apply(name, ch, []any{input[name]})
		if err != nil {
			return nil
		}
	}
	return nil
}

func (e *Engine) apply(name string, ch Channel, writes []any) error {
	changed, err := ch.Updates(writes)
	if err != nil {
		return err
	}
	if changed {
		e.clock++
		e.versions[name] = e.clock
	}
	return nil
}
