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

func updateTableRows(t *Table, cnt []docker.APIContainers, selCtx *SelectableContext) {
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
		t.DeleteRow(node.Hash) //
		if _, e := foundIds[node.Hash]; !e {
			DeleteSelectableNode(node.Hash, selCtx)
		}
		node = node.Next
	}

	redrawRows(t, selCtx)
}

func redrawRows(t *Table, selCtx *SelectableContext) {
	node := selCtx.Head
	totalNodes := len(selCtx.Nodes)
	var row *TableRow
	for i := 0; i < totalNodes; i++ {
		t.DeleteRow(node.Hash)
		dockCont := node.Container.(docker.APIContainers)
		row = t.AddRow(dockCont.ID, dockCont.Image)

		row.Group = dockCont.ID
		if node == selCtx.CurrentSelection {
			row.Fg = termbox.ColorBlue
		}
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

	selCtx := NewSelectablesContext()
	layer := NewLayer()
	tableColWidths := []int{(width * 1 / 3) * 1 / 4, (width * 1 / 3) * 3 / 4}
	t := NewTable(0, 0, (width * 1 / 3), height,
		[]string{"ID", "Image"},
		tableColWidths,
	)

	contInfo := NewContainer(width*1/3+1, 0, width*2/3, 30)
	columns := Createlayout(contInfo)

	layer.Add(t)
	layer.Add(contInfo)

	defer termbox.Close()
	// draw()
loop:
	for {
		select {
		case ev := <-event_queue:
			if ev.Type == termbox.EventResize {
				width, height = termbox.Size()
				t.Width = (width * 1 / 3)
				t.SetColWidth(tableColWidths)
				contInfo.X = width*1/3 + 1
				contInfo.Width = width * 2 / 3
				contInfo.ContainerElements = make([]*ContainerElement, 0)
				columns = Createlayout(contInfo)

			} else if ev.Type == termbox.EventKey {

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
					redrawRows(t, selCtx)
					if selCtx.CurrentSelection == nil {
						break
					}
					for _, cl := range columns {
						cl.WordRef.WordString = inspectContainer(selCtx.CurrentSelection.Hash, cl.Data)
					}
					layer.Draw()
					termbox.Flush()
				}

			}

		case cnt := <-containers_queue:
			updateTableRows(t, cnt, selCtx)
			layer.Draw()
			termbox.Flush()
		}
	}
}
