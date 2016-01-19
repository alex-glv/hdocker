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

func draw() {
	// w, h := termbox.Size()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	endpoint := "unix:///var/run/docker.sock"
	client, _ := docker.NewClient(endpoint)
	cnt, _ := client.ListContainers(docker.ListContainersOptions{})
	rows := make([]layerdraw.TableRow, 0)
	for _, c := range cnt {
		rows = append(rows, layerdraw.TableRow{
			Row: []string{c.ID, c.Image, c.Status},
		})
	}
	cols := []string{"ID", "Image", "Status"}

	widths := []int{10, 40, 40}
	tbl := layerdraw.NewTable(cols, rows, widths)
	el := layerdraw.NewElement(0, 0, 40, 40, tbl)

	layer := layerdraw.NewLayer()
	layer.Add(el)

	layer.Draw()
	termbox.Flush()
}

func main() {
	event_queue := make(chan termbox.Event)
	// layers := make([]Layer, 5)
	go func() {
		for {
			event_queue <- termbox.PollEvent()
		}
	}()
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	draw()
loop:
	for {
		select {
		case ev := <-event_queue:
			if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
				break loop
			}
		default:
			draw()
			time.Sleep(100 * time.Millisecond)
		}
	}
}
