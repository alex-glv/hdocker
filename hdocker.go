package main

import (
	"github.com/alex-glv/hdocker/layerdraw"
	"github.com/alex-glv/hdocker/selectables"
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"
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

var endpoint = "unix:///var/run/docker.sock"

func getRunningContainers() []docker.APIContainers {
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	return cnt
}

func updateTableRows(t *layerdraw.Table, lc *layerdraw.Container, cnt []docker.APIContainers, nodes map[string]*selectables.Node) {
	foundIds := make(map[string]bool)

	for _, c := range cnt {
		if _, e := nodes[c.ID]; !e {
			cNode := &selectables.Node{
				Hash:      c.ID,
				Container: c,
			}
			selectables.AddSelectableNode(cNode, nodes)
		} else {
			nodes[c.ID].Container = c
		}
		foundIds[c.ID] = true
	}
	node := selectables.Head
	for i := 0; i < len(selectables.Nodes); i++ {
		lc.DeleteGroup(node.Hash)
		if _, e := foundIds[node.Hash]; !e {
			selectables.DeleteSelectableNode(node.Hash, selectables.Nodes)

		} else {
			dockCont := node.Container.(docker.APIContainers)
			row := layerdraw.NewTableRow(node.Selected, dockCont.ID, dockCont.Image, dockCont.Command, dockCont.Status, dockCont.Names[0])
			lc.AddTableRow(t, row, dockCont.ID)
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

	headerElement.AddTableHeader(table)

	rowsElement := layerdraw.NewContainer(0, 1, width-2, height)
	updateTableRows(table, rowsElement, getRunningContainers(), selectables.Nodes)

	layer.Add(headerElement)
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
					if len(selectables.Nodes) == 0 {
						break
					}
					if selected, e := selectables.Nodes[selectables.Selectedhash]; e {
						if ev.Key == termbox.KeyArrowDown {
							selected.Next.Selected = true
							selectables.Selectedhash = selected.Next.Hash
						} else {
							selected.Prev.Selected = true
							selectables.Selectedhash = selected.Prev.Hash
						}
						selected.Selected = false

					} else if selectables.Head != nil {
						selectables.Selectedhash = selectables.Head.Hash
						selectables.Head.Selected = true
					}
					updateTableRows(table, rowsElement, getRunningContainers(), selectables.Nodes)
					layer.Draw()
					termbox.Flush()
				}

			}

		case cnt := <-containers_queue:
			updateTableRows(table, rowsElement, cnt, selectables.Nodes)
			layer.Draw()
			termbox.Flush()
		}
	}
}
