package graph

import (
	"context"
	"fmt"
)

const (
	// special nodes
	StartNode = "start"
	EndNode   = "end"
)

// Always is the default condition used when AddEdge is called with a
// nil condition — the edge is unconditionally taken.
func Always[S any](s S) bool {
	return true
}

// NodeFunc is a unit of work, parameterized over the graph's state type.
// Every node in a given Graph[S] shares the same S — that's what lets a
// single Run() thread one state value through the whole traversal.
type NodeFunc[S any] func(ctx context.Context, s S) (newState S, err error)

// Condition decides whether an edge should be taken, given the state
// produced by the node it originates from.
type Condition[S any] func(s S) bool

type Node[S any] struct {
	Name string
	Run  NodeFunc[S]
}

type Edge[S any] struct {
	To        string
	Condition Condition[S]
}

type Graph[S any] struct {
	Nodes map[string]Node[S]
	Edges map[string][]Edge[S]
}

func NewGraph[S any]() *Graph[S] {
	return &Graph[S]{
		Nodes: make(map[string]Node[S]),
		Edges: make(map[string][]Edge[S]),
	}
}

func (g *Graph[S]) AddNode(node Node[S]) error {
	if _, exists := g.Nodes[node.Name]; exists {
		return fmt.Errorf("node with name %s already exists", node.Name)
	}
	g.Nodes[node.Name] = node
	return nil
}

func (g *Graph[S]) AddEdge(from string, to string, condition Condition[S]) error {
	if _, exists := g.Nodes[from]; !exists {
		return fmt.Errorf("node with name %s does not exist", from)
	}
	if _, exists := g.Nodes[to]; !exists {
		return fmt.Errorf("node with name %s does not exist", to)
	}
	if condition == nil {
		condition = Always[S]
	}
	g.Edges[from] = append(g.Edges[from], Edge[S]{
		To:        to,
		Condition: condition,
	})
	return nil
}

func (g *Graph[S]) Run(ctx context.Context, from string, initial S) (S, error) {
	if from == "" {
		from = StartNode
	}
	currentNode := from
	state := initial
	for {
		node, ok := g.Nodes[currentNode]
		if !ok {
			return state, fmt.Errorf("node with name %s does not exist", currentNode)
		}
		newState, err := node.Run(ctx, state)
		if err != nil {
			return newState, err
		}
		state = newState

		if currentNode == EndNode {
			return state, nil
		}

		nextNode, err := g.getNextNode(currentNode, state)
		if err != nil {
			return state, err
		}
		currentNode = nextNode
	}
}

func (g *Graph[S]) Validate() error {
	if len(g.Nodes) == 0 {
		return fmt.Errorf("graph does not have any nodes")
	}
	for name := range g.Nodes {
		if name == "" {
			return fmt.Errorf("node name cannot be empty")
		}
	}
	for name, node := range g.Nodes {
		if node.Run == nil {
			return fmt.Errorf("node with name %s does not have an executable function", name)
		}
	}
	for name := range g.Nodes {
		if name != EndNode {
			if _, exists := g.Edges[name]; !exists {
				return fmt.Errorf("node with name %s does not have any outgoing edges", name)
			}
		}
	}
	for from, edges := range g.Edges {
		for _, edge := range edges {
			if _, exists := g.Nodes[edge.To]; !exists {
				return fmt.Errorf("edge from %s to %s has a destination node that does not exist", from, edge.To)
			}
		}
	}
	incomingEdges := make(map[string]bool)
	for _, edges := range g.Edges {
		for _, edge := range edges {
			incomingEdges[edge.To] = true
		}
	}
	if _, exists := incomingEdges[StartNode]; exists {
		return fmt.Errorf("start node should not have any incoming edges")
	}
	for from := range g.Edges {
		if _, exists := g.Nodes[from]; !exists {
			return fmt.Errorf("edge from %s has a source node that does not exist", from)
		}
	}
	if _, exists := g.Edges[EndNode]; exists {
		return fmt.Errorf("end node should not have any outgoing edges")
	}
	if _, exists := g.Nodes[StartNode]; !exists {
		return fmt.Errorf("graph does not have a start node")
	}
	if _, exists := g.Nodes[EndNode]; !exists {
		return fmt.Errorf("graph does not have an end node")
	}
	return nil
}

func (g *Graph[S]) getNextNode(current string, state S) (string, error) {
	edges, exists := g.Edges[current]
	if !exists {
		return "", fmt.Errorf("no outgoing edges from node %s", current)
	}
	for _, edge := range edges {
		if edge.Condition(state) {
			return edge.To, nil
		}
	}
	return "", fmt.Errorf("no valid outgoing edge from node %s", current)
}