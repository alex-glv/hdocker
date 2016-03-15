package main

import (

	// "github.com/alex-glv/hdocker/drawable" //
	"bytes"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/nsf/termbox-go"
	"html/template"
	"io/ioutil"
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

func createLayout(infoBox *Container) []Column {
	dat, err := ioutil.ReadFile("./sample.layout.json")
	if err != nil {
		panic(err)
	}
	columns := ParseLayout(dat)

	curFill := 0
	curValues := make([]int, 0, 0)

	for i, v := range columns {
		infoBox.Add(NewWordDef(v.Title, v.Width))
		infoBox.Add(Space())
		curValues = append(curValues, i)
		curFill = curFill + v.Width
		if curFill >= 100 || i == len(columns)-1 {
			infoBox.Add(LineBreak())
			for _, ci := range curValues {
				columns[ci].WordRef = NewWordDef(columns[ci].Data, columns[ci].Width)
				infoBox.Add(columns[ci].WordRef)
				infoBox.Add(Space())

			}
			curValues = curValues[0:0]
			curFill = 0
			infoBox.Add(LineBreak())
		}

	}

	return columns
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
	widths := []int{10, 40}

	headerElement := NewContainer(0, 0, 50, 1)
	table := headerElement.NewTableWithHeader(cols, widths)
	rowsElement := NewContainer(0, 1, 50, height)

	contInfo := NewContainer(52, 0, 100, 30)
	columns := createLayout(contInfo)

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
						// ip.WordString = ""
						// todo: nullify all layout fields
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

					for _, cl := range columns {
						cl.WordRef.WordString = inspectContainer(selCtx.CurrentSelection.Hash, cl.Data)
					}
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
