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
	TR       layerdraw.TableRow
}

type Nodes map[string]*Node

var nodes Nodes
var tail *Node
var endpoint = "unix:///var/run/docker.sock"
var rowNodes *Nodes

func AddSelectableNode(groupNode *Node) {
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

func pollContainers(c chan []docker.APIContainers) {
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	c <- cnt
}

func updateTableRows(cnt []docker.APIContainers) []*layerdraw.TableRow {
	rows := make([]*layerdraw.TableRow, 0)
	for _, c := range cnt {
		rows = append(rows, layerdraw.NewTableRow(c.ID, c.Image, c.Status, c.Names[0]))
	}
	return rows
}

func drawContainersTable(width, height int) ([]string, []int) {
	cols := []string{"ID", "Image", "Created", "Name"}
	widths := []int{width / 4, width / 4, width / 4, width / 4}
	return cols, widths
}

func draw() {
	// w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func main() {
	event_queue := make(chan termbox.Event)
	containers_queue := make(chan []docker.APIContainers)

	go func() {
		for {
			pollContainers(containers_queue)
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
	cols, widths := drawContainersTable(width-2, height-50)

	table := layerdraw.NewTable(cols, widths)
	headerElement := layerdraw.NewContainer(0, 0, width, 1)

	layer := layerdraw.NewLayer()
	layer.Add(headerElement)

	headerElement.AddTableHeader(table)
	headerElement.Draw()

	rowsElement := layerdraw.NewContainer(0, 1, width-2, height-50)

	// el.AddTable(cols, rows, widths)
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

			rows := updateTableRows(cnt)
			rowsElement.AddTableRows(table, rows)
			rowsElement.Draw()

			termbox.Flush()
		}
	}
}
