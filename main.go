package main

import (

	// "github.com/alex-glv/hdocker/drawable" //
	"bytes"
	"fmt"
	"html/template"
	"os"
	"time"

	"log"

	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"
)

var logFile, _ = os.Create("/tmp/log.out")
var logger = log.New(logFile, "", 0)

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

func inspectContainer(cid string, templ string) string {
	dckCtx := getDockerContext()
	inspect, _ := dckCtx.client.InspectContainer(cid)

	if inspect == nil {
		return ""
	}

	tpl, err := template.New("container").Parse(templ)

	if err != nil {
		panic(fmt.Sprintf("can't parse: %s", templ))
	}

	buf := new(bytes.Buffer)
	tpl.Execute(buf, inspect)

	return buf.String()

}

func killContainer(cid string) {
	dckCtx := getDockerContext()
	dckCtx.client.KillContainer(docker.KillContainerOptions{
		ID: cid,
	})
}

func getRunningContainers() []docker.APIContainers {
	dckCtx := getDockerContext()
	cnt, _ := dckCtx.client.ListContainers(docker.ListContainersOptions{})

	return cnt
}

func updateTableRows(t *Table, lc *Container, cnt []docker.APIContainers, selCtx *SelectableContext) {
	foundIds := make(map[string]bool)
	nodes := selCtx.Nodes
	for _, c := range cnt {
		if _, e := nodes[c.ID]; !e {
			cNode := &Node{
				Hash:      c.ID,
				Container: c,
			}
			AddSelectableNode(cNode, selCtx)
		} else {
			nodes[c.ID].Container = c
		}
		foundIds[c.ID] = true
	}
	node := selCtx.Head

	for i := 0; i < len(nodes); i++ {
		lc.DeleteGroup(node.Hash)
		if _, e := foundIds[node.Hash]; !e {
			DeleteSelectableNode(node.Hash, selCtx)
		}
		node = node.Next
	}
	logger.Println("new length:", len(selCtx.Nodes))
	redrawRows(t, lc, selCtx)
}

func redrawRows(t *Table, lc *Container, selCtx *SelectableContext) {
	node := selCtx.Head
	totalNodes := len(selCtx.Nodes)

	for i := 0; i < totalNodes; i++ {
		lc.DeleteGroup(node.Hash)

		dockCont := node.Container.(docker.APIContainers)
		row := NewTableRow(dockCont.ID, dockCont.Image)
		if node == selCtx.CurrentSelection {
			row.Fg = termbox.ColorBlue
		}
		lc.AddTableRow(t, row, dockCont.ID)
		node = node.Next
	}
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
	logger.Println("Canvas w:", width, ", h:", height)
	selCtx := NewSelectablesContext()
	layer := NewLayer()

	cols := []string{"ID", "Image"}
	widths := []int{(width * 1 / 3) * 1 / 4, (width * 1 / 3) * 3 / 4}

	headerElement := NewContainer(0, 0, (width * 1 / 3), 1)
	table := headerElement.NewTableWithHeader(cols, widths)
	rowsElement := NewContainer(0, 1, (width * 1 / 3), height)

	contInfo := NewContainer(width*1/3+1, 0, width*2/3, 30)
	columns := Createlayout(contInfo)

	// layer.Add(headerElement)
	// layer.Add(rowsElement)
	// layer.Add(contInfo)

	// experimental

	expTable := NewTable([]string{"Col1", "Col2", "Col3"}, []int{10, 10, 10})
	expTable.AddRow(NewTableRow([]string{"CCellR1", "CCellC2", "YCellY2"}...))
	expTable.AddRow(NewTableRow([]string{"RCellR1", "RCellR2", "XCellX2"}...))

	expContainer := NewContainer(0, 0, 60, 30)

	expContainer.Add(NewWordDef("Tbl:", 4))
	expContainer.Add(expTable)
	layer.Add(expContainer)

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
				if ev.Key == termbox.KeyCtrlX {
					killContainer(selCtx.CurrentSelection.Hash)
					break
				}

				if ev.Key == termbox.KeyArrowDown || ev.Key == termbox.KeyArrowUp {
					logger.Println("Arrow key pressed:", ev.Key)
					next := ev.Key == termbox.KeyArrowDown
					Advance(selCtx, next)
					redrawRows(table, rowsElement, selCtx)
					if selCtx.CurrentSelection == nil {
						break
					}
					for _, cl := range columns {
						cl.WordRef.WordString = inspectContainer(selCtx.CurrentSelection.Hash, cl.Data)
					}
					// layer.Draw()
					termbox.Flush()
				}

			}

		case cnt := <-containers_queue:
			updateTableRows(table, rowsElement, cnt, selCtx)
			// layer.Draw()
			termbox.Flush()
		}
	}
}
