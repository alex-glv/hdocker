package selectables

type Node struct {
	Prev      *Node
	Next      *Node
	Container interface{}
	Hash      string
	Selected  bool
}

var Nodes = make(map[string]*Node)
var Tail *Node
var Head *Node
var Selectedhash string

func DeleteSelectableNode(hash string, nodes map[string]*Node) {
	tbd, e := nodes[hash]
	if !e {
		return
	}
	if Tail == tbd {
		Tail = tbd.Prev
	}
	if Head == tbd {
		Head = tbd.Next
	}
	prev := tbd.Prev
	next := tbd.Next
	prev.Next = tbd.Next
	next.Prev = tbd.Prev

	delete(nodes, hash)

}

func AddSelectableNode(groupNode *Node, nodes map[string]*Node) {
	_, exists := nodes[groupNode.Hash]
	if exists {
		return
	}

	if len(nodes) == 0 {
		Head = groupNode
		Tail = groupNode
	}

	groupNode.Prev = Tail
	groupNode.Next = Head
	Head.Prev = groupNode
	Tail.Next = groupNode
	Tail = groupNode
	nodes[groupNode.Hash] = groupNode
}
