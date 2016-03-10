package main

import (
	// "fmt"
	// "github.com/alex-glv/hdocker/drawable" //
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"
	"strings"
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

func getContainerCmd(cid string) string {
	dckCtx := getDockerContext()

	inspect, _ := dckCtx.client.InspectContainer(cid)
	if inspect == nil {
		return ""
	}

	cmdArr := make([]string, 1)
	cmdArr[0] = inspect.Path
	cmdArr = append(cmdArr, inspect.Args...)
	return strings.Join(cmdArr, " ")

	// return inspect.
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
	_, height := termbox.Size()
	selCtx := NewSelectablesContext()
	layer := NewLayer()

	cols := []string{"ID", "Image"}
	widths := []int{10, 50}

	headerElement := NewContainer(0, 0, 60, 1)
	table := headerElement.NewTableWithHeader(cols, widths)

	rowsElement := NewContainer(0, 1, 60, height)

	ip := NewWordDef("", 50)
	cmd := NewWordDef("", 50)
	contInfo := NewContainer(62, 0, 50, 30)
	contInfo.Add(NewWordDef("IP", 50))
	contInfo.Add(LineBreak())
	contInfo.Add(ip)
	contInfo.Add(LineBreak())
	contInfo.Add(NewWordDef("Command", 50))
	contInfo.Add(LineBreak())
	contInfo.Add(cmd)

	layer.Add(headerElement)
	layer.Add(rowsElement)
	layer.Add(contInfo)

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
				}

				if ev.Key == termbox.KeyArrowDown || ev.Key == termbox.KeyArrowUp {
					if len(selCtx.Nodes) == 0 {
						ip.WordString = ""
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

					redrawRows(table, rowsElement, selCtx)
					ip.WordString = getContainerIp(selCtx.CurrentSelection.Hash)
					cmd.WordString = getContainerCmd(selCtx.CurrentSelection.Hash)

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
