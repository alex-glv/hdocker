package selectables

import ()

type Node struct {
	Prev      *Node
	Next      *Node
	Container interface{}
	Hash      string
	Selected  bool
}

type SelectableContext struct {
	Nodes            map[string]*Node
	Head             *Node
	Tail             *Node
	CurrentSelection *Node
}

func New() *SelectableContext {
	return &SelectableContext{
		Nodes: make(map[string]*Node),
	}
}

func DeleteSelectableNode(hash string, context *SelectableContext) {
	nodes := context.Nodes
	tbd, e := nodes[hash]
	if !e {
		return
	}
	prev := nodes[hash].Prev
	next := nodes[hash].Next

	if context.Tail == tbd {
		context.Tail = tbd.Prev
	}
	if context.Head == tbd {
		context.Head = tbd.Next
	}
	prev.Next = next
	next.Prev = prev

	delete(nodes, hash)

}

func AddSelectableNode(groupNode *Node, context *SelectableContext) {
	nodes := context.Nodes
	if _, e := nodes[groupNode.Hash]; e {
		return
	}

	nodes[groupNode.Hash] = groupNode

	if len(nodes) == 1 {
		context.Head = groupNode
		context.Tail = groupNode
	}

	groupNode.Prev = context.Tail
	groupNode.Next = context.Head
	context.Head.Prev = groupNode
	context.Tail.Next = groupNode
	context.Tail = groupNode

	// fmt.Println(Tail == Head)
	// fmt.Println(Head.Prev == Head.Next)

}
