package main

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"

	"github.com/alex-glv/hdocker/layerdraw"
	// "github.com/alex-glv/hdocker/selectables"
	"fmt"
	"time"
)

/*

  *------*
 / *------*
/_/ *------*
 /_/      /
  /______/

Layers

*/

type Node struct {
	Prev      *Node
	Next      *Node
	Container docker.APIContainers
	Hash      string
	Selected  bool
}

var nodes = make(map[string]*Node)
var tail *Node
var head *Node
var selectedHash string
var endpoint = "unix:///var/run/docker.sock"

func getRunningContainers() []docker.APIContainers {
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	return cnt
}

func AddSelectableNode(groupNode *Node, nodes map[string]*Node) {
	_, exists := nodes[groupNode.Hash]
	if exists {
		return
	}

	if len(nodes) == 0 {
		head = groupNode
		tail = groupNode
	}

	groupNode.Prev = tail
	groupNode.Next = head
	tail.Next = groupNode
	tail = groupNode
	nodes[groupNode.Hash] = groupNode
}

func DeleteSelectableNode(hash string, nodes map[string]*Node) {
	tbd, e := nodes[hash]
	if !e {
		return
	}
	if tail == tbd {
		tail = tbd.Prev
	}
	if head == tbd {
		head = tbd.Next
	}
	prev := tbd.Prev
	next := tbd.Next
	prev.Next = tbd.Next
	next.Prev = tbd.Prev

	delete(nodes, hash)

}
func updateTableRows(t *layerdraw.Table, lc *layerdraw.Container, cnt []docker.APIContainers, nodes map[string]*Node) {
	foundIds := make(map[string]bool)

	for _, c := range cnt {
		if _, e := nodes[c.ID]; !e {
			cNode := &Node{
				Hash:      c.ID,
				Container: c,
			}
			AddSelectableNode(cNode, nodes)
		} else {
			nodes[c.ID].Container = c
		}
		foundIds[c.ID] = true
	}
	node := head
	for i := 0; i < len(nodes); i++ {
		// fmt.Println(node.Hash)
		lc.DeleteGroup(node.Hash)
		if _, e := foundIds[node.Hash]; !e {
			DeleteSelectableNode(node.Hash, nodes)

		} else {
			row := layerdraw.NewTableRow(node.Selected, node.Container.ID, node.Container.Image, node.Container.Command, node.Container.Status, node.Container.Names[0])
			lc.AddTableRow(t, row, node.Container.ID)
		}
		node = node.Next
	}

}

func drawContainersTable(width, height int) ([]string, []int) {
	cols := []string{"ID", "Image", "Command", "Created", "Name"}
	widths := []int{width / 10, width / 4, width / 4, width / 4, width / 3}
	return cols, widths
}

func main() {
	event_queue := make(chan termbox.Event)
	containers_queue := make(chan []docker.APIContainers)

	go func() {
		for {
			cnt := getRunningContainers()
			containers_queue <- cnt
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go func() {
		for {
			event_queue <- termbox.PollEvent()
		}
	}()

	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	width, height := termbox.Size()
	cols, widths := drawContainersTable(width-2, height)

	table := layerdraw.NewTable(cols, widths)
	headerElement := layerdraw.NewContainer(0, 0, width, 1)

	layer := layerdraw.NewLayer()

	headerElement.AddTableHeader(table)

	rowsElement := layerdraw.NewContainer(0, 1, width-2, height-20)
	updateTableRows(table, rowsElement, getRunningContainers(), nodes)

	debugElement := layerdraw.NewContainer(0, height-2, width, 2)
	debugTable := layerdraw.NewTable([]string{"Selected"}, []int{width})
	debugElement.AddTableRow(debugTable, layerdraw.NewTableRow(false, "Test"), "selHash")

	layer.Add(headerElement)
	layer.Add(debugElement)
	layer.Add(rowsElement)

	layer.Draw()
	termbox.Flush()
	defer termbox.Close()
	// draw()
loop:
	for {
		select {
		case ev := <-event_queue:
			if ev.Type == termbox.EventKey {
				if ev.Key == termbox.KeyEsc {
					break loop
				}

				if ev.Key == termbox.KeyArrowDown || ev.Key == termbox.KeyArrowUp {
					if len(nodes) == 0 {
						break
					}
					var line string
					if selected, e := nodes[selectedHash]; e {
						if ev.Key == termbox.KeyArrowDown {
							selected.Next.Selected = true
							line = selected.Next.Hash
						} else {
							selected.Prev.Selected = true
							line = selected.Prev.Hash
						}
						selected.Selected = false

					} else if head != nil {
						selectedHash = head.Hash
						head.Selected = true
					}
					debugElement.DeleteGroup("selHash")
					debugElement.DeleteGroup("headTail")
					debugElement.AddTableRow(debugTable, layerdraw.NewTableRow(false, line), "selHash")
					debugElement.AddTableRow(debugTable, layerdraw.NewTableRow(false, fmt.Sprintf(head.Hash, " <> ", tail.Hash)), "headTail")
					updateTableRows(table, rowsElement, getRunningContainers(), nodes)
					layer.Draw()
					termbox.Flush()
				}

			}

		case cnt := <-containers_queue:
			updateTableRows(table, rowsElement, cnt, nodes)
			layer.Draw()
			termbox.Flush()
		}
	}
}
