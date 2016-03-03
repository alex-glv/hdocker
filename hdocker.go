package main

import (
	"github.com/alex-glv/hdocker/layerdraw"
	"github.com/alex-glv/hdocker/selectables"
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"

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

type DockerContext struct {
	client   *docker.Client
	endpoint string
}

var dckCtx *DockerContext

func getDockerContext() *DockerContext {
	if dckCtx == nil {

		endpoint := "unix:///var/run/docker.sock"
		client, err := docker.NewClient(endpoint)
		if err != nil {
			panic("Can't establish docker connection")
		}
		dckCtx = &DockerContext{
			client:   client,
			endpoint: endpoint,
		}
	}

	return dckCtx
}

func getContainerIp(cid string) string {
	dckCtx := getDockerContext()
	inspect, _ := dckCtx.client.InspectContainer(cid)
	if inspect == nil {
		return ""
	}

	return inspect.NetworkSettings.IPAddress
}

func getRunningContainers() []docker.APIContainers {
	dckCtx := getDockerContext()
	cnt, _ := dckCtx.client.ListContainers(docker.ListContainersOptions{})

	return cnt
}

func updateTableRows(t *layerdraw.Table, lc *layerdraw.Container, cnt []docker.APIContainers, selCtx *selectables.SelectableContext) {
	foundIds := make(map[string]bool)
	nodes := selCtx.Nodes
	for _, c := range cnt {
		if _, e := nodes[c.ID]; !e {
			cNode := &selectables.Node{
				Hash:      c.ID,
				Container: c,
			}
			selectables.AddSelectableNode(cNode, selCtx)
		} else {
			nodes[c.ID].Container = c
		}
		foundIds[c.ID] = true
	}
	node := selCtx.Head

	for i := 0; i < len(nodes); i++ {
		lc.DeleteGroup(node.Hash)
		if _, e := foundIds[node.Hash]; !e {
			selectables.DeleteSelectableNode(node.Hash, selCtx)
		}
		node = node.Next
	}
	syncRows(t, lc, selCtx)
}

func syncRows(t *layerdraw.Table, lc *layerdraw.Container, selCtx *selectables.SelectableContext) {
	node := selCtx.Head
	totalNodes := len(selCtx.Nodes)

	for i := 0; i < totalNodes; i++ {
		lc.DeleteGroup(node.Hash)

		dockCont := node.Container.(docker.APIContainers)
		row := layerdraw.NewTableRow(dockCont.ID, dockCont.Image)
		if node == selCtx.CurrentSelection {
			row.Fg = termbox.ColorBlue
		}
		lc.AddTableRow(t, row, dockCont.ID)
		node = node.Next
	}
}

func drawContainersTable(width, height int) ([]string, []int) {
	cols := []string{"ID", "Image"}
	widths := []int{width / 2, width / 2}
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
	selCtx := selectables.New()
	layer := layerdraw.NewLayer()
	cols, widths := drawContainersTable(width/3, height)

	headerElement := layerdraw.NewContainer(0, 0, width/3, 1)
	table := headerElement.NewTableWithHeader(cols, widths)

	rowsElement := layerdraw.NewContainer(0, 1, width/3, height)

	cntInfoElement := layerdraw.NewContainer(width/3+2, 0, width/2*3, 10)
	cntInfoTable := cntInfoElement.NewTableWithHeader([]string{"IP", "Status"}, []int{width / 4, width / 4})

	layer.Add(headerElement)
	layer.Add(rowsElement)
	layer.Add(cntInfoElement)

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
					if len(selCtx.Nodes) == 0 {
						break
					}

					if selCtx.CurrentSelection != nil {
						if ev.Key == termbox.KeyArrowDown {
							selCtx.CurrentSelection = selCtx.CurrentSelection.Next
						} else {
							selCtx.CurrentSelection = selCtx.CurrentSelection.Prev
						}

					} else if selCtx.Head != nil {
						selCtx.CurrentSelection = selCtx.Head
					} else {
						panic("Head is missing! Where's my mind?")
					}

					syncRows(table, rowsElement, selCtx)
					cntInfoElement.DeleteGroup("cntInfo")
					cntInfoElement.AddTableRow(cntInfoTable, layerdraw.NewTableRow(getContainerIp(selCtx.CurrentSelection.Hash), ""), "cntInfo")
					layer.Draw()
					termbox.Flush()
				}

			}

		case cnt := <-containers_queue:
			updateTableRows(table, rowsElement, cnt, selCtx)
			layer.Draw()
			termbox.Flush()
		}
	}
}
