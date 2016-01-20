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

var endpoint = "unix:///var/run/docker.sock"

func pollContainers(c chan []docker.APIContainers) {
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	c <- cnt
}

func redrawContainers(cnt []docker.APIContainers) []layerdraw.TableRow {
	rows := make([]layerdraw.TableRow, 0)
	for _, c := range cnt {
		rows = append(rows, layerdraw.TableRow{
			Row: []string{c.ID, c.Image, c.Status, c.Names[0]},
		})
	}
	return rows
}

func drawContainersTable(width, height int, l *layerdraw.Layer) *layerdraw.Table {
	cols := []string{"ID", "Image", "Status", "Name"}
	widths := []int{width / 4, width / 4, width / 4, width / 4}
	return layerdraw.NewTable(cols, make([]layerdraw.TableRow, 0), widths)
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
	layer := layerdraw.NewLayer()
	tbl := drawContainersTable(width, height, layer)
	el := layerdraw.NewElement(0, 0, width, height-50, tbl)
	layer.Add(el)
	layer.Draw()
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
			tbl.Rows = redrawContainers(cnt)
			layer.Draw()
			termbox.Flush()
		}
	}
}
