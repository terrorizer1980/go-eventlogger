package eventlogger

import (
	"errors"
	"fmt"
)

type NodeType int

const (
	_ NodeType = iota
	NodeTypeFilter
	NodeTypeFormatter
	NodeTypeSink
)

// A Node in a Graph
type Node interface {
	// Process does something with the Event: filter, redaction,
	// marshalling, persisting.
	Process(e *Event) (*Event, error)
	// Reopen is used to re-read any config stored externally
	// and to close and reopen files, e.g. for log rotation.
	Reopen() error
	Name() string
	Type() NodeType
}

// A LinkableNode is a Node that has downstream children.  Nodes
// that are *not* LinkableNodes are Leafs.
type LinkableNode interface {
	Node
	SetNext([]Node)
	Next() []Node
}

// LinkNodes is a convenience function that connects
// Nodes together into a linked list. All of the nodes except the
// last one must be LinkableNodes
func LinkNodes(nodes []Node) ([]Node, error) {
	num := len(nodes)
	if num < 2 {
		return nodes, nil
	}

	for i := 0; i < num-1; i++ {
		ln, ok := nodes[i].(LinkableNode)
		if !ok {
			return nil, errors.New("Node is not Linkable")
		}
		ln.SetNext([]Node{nodes[i+1]})
	}

	return nodes, nil
}

func LinkNodesAndSinks(inner, sinks []Node) ([]Node, error) {
	_, err := LinkNodes(inner)
	if err != nil {
		return nil, err
	}

	ln, ok := inner[len(inner)-1].(LinkableNode)
	if !ok {
		return nil, fmt.Errorf("last inner node not linkable")
	}

	ln.SetNext(sinks)

	return inner, nil
}
