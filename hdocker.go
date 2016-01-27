package main

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"

	"github.com/alex-glv/hdocker/layerdraw"
	// "github.com/alex-glv/hdocker/selectables"
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

func AddSelectableNode(groupNode *Node, nodes map[string]*Node) {
	_, exists := nodes[groupNode.Hash]
	if exists {
		return
	}
	nodes[groupNode.Hash] = groupNode
	if tail == nil {
		tail = groupNode
	} else {
		tail.Next = groupNode
		tail = groupNode
	}
}

func getRunningContainers() []docker.APIContainers {
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	return cnt
}

func DeleteSelectableNode(hash string) {
	if tail == nodes[hash] {
		tail = nodes[hash].Prev
	}
	delete(nodes, hash)

}
func updateTableRows(t *layerdraw.Table, lc *layerdraw.Container, cnt []docker.APIContainers) {
	foundIds := make(map[string]bool)
	for _, n := range nodes {
		lc.DeleteGroup(n.Hash)
	}

	for _, c := range cnt {
		if _, e := nodes[c.ID]; !e {
			cNode := &Node{
				Prev: tail,
				Hash: c.ID,
			}

			AddSelectableNode(cNode, nodes)
			row := layerdraw.NewTableRow(c.ID, c.Image, c.Status, c.Names[0])
			lc.AddTableRow(t, row, c.ID)

		}

		foundIds[c.ID] = true
	}
	for _, n := range nodes {
		if _, e := foundIds[n.Hash]; !e {
			DeleteSelectableNode(n.Hash)
			lc.DeleteGroup(n.Hash)
		}
	}
}

func drawContainersTable(width, height int) ([]string, []int) {
	cols := []string{"ID", "Image", "Created", "Name"}
	widths := []int{width / 6, width / 3, width / 4, width / 3}
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
	cols, widths := drawContainersTable(width-2, height-50)

	table := layerdraw.NewTable(cols, widths)
	headerElement := layerdraw.NewContainer(0, 0, width, 1)

	layer := layerdraw.NewLayer()
	layer.Add(headerElement)

	headerElement.AddTableHeader(table)
	headerElement.Draw()

	rowsElement := layerdraw.NewContainer(0, 1, width-2, height-50)
	updateTableRows(table, rowsElement, getRunningContainers())
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

			// case cnt := <-containers_queue:
			// 	updateTableRows(table, rowsElement, cnt)
			// 	// rowsElement.AddTableRows(table, rows)
			// 	rowsElement.Draw()
			// 	termbox.Flush()
		}
	}
}
