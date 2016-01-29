package main

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"

	"github.com/alex-glv/hdocker/layerdraw"
	// "github.com/alex-glv/hdocker/selectables"
	// "fmt"
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
	Prev     *Node
	Next     *Node
	Hash     string
	Selected int
}

var nodes = make(map[string]*Node)
var tail *Node
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
	nodes[groupNode.Hash] = groupNode
	if tail == nil {
		tail = groupNode
	} else {
		groupNode.Prev = tail
		tail.Next = groupNode
		tail = groupNode
	}
}

func DeleteSelectableNode(hash string, nodes map[string]*Node) {
	tbd, e := nodes[hash]
	if !e {
		return
	}
	if tail == nodes[hash] {
		tail = nodes[hash].Prev
	}
	prev := tbd.Prev
	next := tbd.Next
	if prev != nil && next == nil {
		prev.Next = next
	}
	if next != nil && prev == nil {
		next.Prev = prev
	}
	if next != nil && prev != nil {
		prev.Next = next
		next.Prev = prev
	}

	delete(nodes, hash)

}
func updateTableRows(t *layerdraw.Table, lc *layerdraw.Container, cnt []docker.APIContainers, nodes map[string]*Node) {
	foundIds := make(map[string]bool)

	for _, c := range cnt {
		lc.DeleteGroup(c.ID)
		if _, e := nodes[c.ID]; !e {
			cNode := &Node{
				Prev: tail,
				Hash: c.ID,
			}
			AddSelectableNode(cNode, nodes)
		}
		foundIds[c.ID] = true
		row := layerdraw.NewTableRow(c.ID, c.Image, c.Command, c.Status, c.Names[0])
		lc.AddTableRow(t, row, c.ID)
	}
	for _, n := range nodes {
		if _, e := foundIds[n.Hash]; !e {
			DeleteSelectableNode(n.Hash, nodes)
			lc.DeleteGroup(n.Hash)
		}
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
			time.Sleep(1000 * time.Millisecond)
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
	layer.Add(headerElement)

	headerElement.AddTableHeader(table)
	headerElement.Draw()

	rowsElement := layerdraw.NewContainer(0, 1, width-2, height)
	updateTableRows(table, rowsElement, getRunningContainers(), nodes)
	rowsElement.Draw()
	termbox.Flush()
	defer termbox.Close()
	// draw()
loop:
	for {
		select {
		case ev := <-event_queue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				break loop
			}

		case cnt := <-containers_queue:
			updateTableRows(table, rowsElement, cnt, nodes)
			rowsElement.Draw()
			termbox.Flush()
		}
	}
}
