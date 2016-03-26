package main

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

func NewSelectablesContext() *SelectableContext {
	return &SelectableContext{
		Nodes: make(map[string]*Node),
	}
}

func Advance(context *SelectableContext, next bool) {
	if !canAdvance(context) {
		return
	}
	if context.CurrentSelection != nil {
		logger.Println("Current selection: ", context.CurrentSelection.Hash)
		if next {
			logger.Println("Selecting .Next")
			context.CurrentSelection = context.CurrentSelection.Next

		} else {
			logger.Println("Selecting .Prev")
			context.CurrentSelection = context.CurrentSelection.Prev
		}
	} else if context.Head != nil {
		logger.Println("Selecting .Head")
		context.CurrentSelection = context.Head
	} else {
		panic("Head is missing! Where's my mind?")
	}
}

func canAdvance(context *SelectableContext) bool {
	if len(context.Nodes) == 0 {
		// ip.WordString = ""
		// todo: nullify all layout fields
		logger.Println("len(selCtx.Nodes) == 0; break")
		return false
	}
	return true
}

func DeleteSelectableNode(hash string, context *SelectableContext) {
	if context.CurrentSelection != nil && context.CurrentSelection.Hash == hash {
		context.CurrentSelection = nil
	}
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
}
