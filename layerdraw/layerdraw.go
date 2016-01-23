package layerdraw

import (
	"github.com/nsf/termbox-go"
)

var DynamicContainer = 2

type Layer struct {
	added      int
	Containers []Container
}

type Container struct {
	X, Y, Width, Height, Options int
	ContainerElements            []ContainerElement
}

type Drawable interface {
	Draw()
}

type ContainerElement interface {
	getMatrix() []RunePos
}

type SelectableElement interface {
	selected()
}

type Word struct {
	WordString string
	Width      int
	state      []RunePos
}

type LineBreak struct {
}

type Table struct {
	Cols      []string
	Rows      []TableRow
	ColWidths []int
	state     []RunePos
	Width     int
}

type RunePos struct {
	X, Y int
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

type TableRow []string

func NewLayer() *Layer {
	els := make([]Container, 0)
	return &Layer{
		Containers: els,
	}
}

func NewWord(word string, width int) *Word {
	return &Word{
		WordString: word,
		Width:      width,
	}
}

func NewLineBreak() *LineBreak {
	return &LineBreak{}
}

func NewTable(cols []string, rows []TableRow, widths []int, width int) *Table {

	tbl := &Table{
		Cols:      cols,
		Rows:      rows,
		ColWidths: widths,
		Width:     width,
	}

	return tbl
}

func NewContainer(x, y, width, height, options int, contents ...ContainerElement) Container {
	return Container{
		X:                 x,
		Y:                 y,
		Width:             width,
		Height:            height,
		ContainerElements: contents,
		Options:           options,
	}
}

func NewTableRow(fields ...string) TableRow {
	row := make(TableRow, 0)
	for _, v := range fields {
		row = append(row, v)
	}
	return row
}

func (l *Layer) Add(el Container) {
	l.Containers = append(l.Containers, el)

}
func (c *Container) Add(el ContainerElement) {
	c.ContainerElements = append(c.ContainerElements, el)
}

func (c *Container) Draw() {
	for x := 0; x < c.X; x++ { // cleanup
		for y := 0; y < c.Y; y++ {
			termbox.SetCell(c.X+x, c.Y+y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
	}
	last := NewRunePos(c.X, c.Y, 0, 0, 0)
	lineBreaks := 1
	for _, v := range c.ContainerElements {
		matrix := v.getMatrix()
		if len(matrix) == 0 {
			last = NewRunePos(c.X, c.Y+lineBreaks, ' ', 0, 0)
			lineBreaks = lineBreaks + 1
			continue
		}
		matrix = addConstant(matrix, last.X, last.Y)
		for _, e := range matrix {
			termbox.SetCell(e.X,
				e.Y,
				e.Char,
				e.Fg,
				e.Bg)
		}
		last = matrix[len(matrix)-1]
		last.X = last.X + 1 // last char position X + 1

	}
	if DynamicContainer&c.Options == DynamicContainer {
		c.ContainerElements = make([]ContainerElement, 0)
	}
}

func (l *Layer) Draw() {
	for _, v := range l.Containers {
		v.Draw()
	}
}

func (w *Word) getMatrix() []RunePos {
	matrix := make([]RunePos, w.Width)
	for i := 0; i < w.Width; i++ {
		chru := byte(' ')
		if i < len(w.WordString) {
			chru = w.WordString[i]
		}
		matrix[i] = NewRunePos(i, 0, chru, termbox.ColorDefault, termbox.ColorDefault)
	}
	return matrix
}

func (l *LineBreak) getMatrix() []RunePos {
	var matrix []RunePos
	return matrix
}

func NewRunePos(x, y int, ch byte, fg, bg termbox.Attribute) RunePos {
	return RunePos{
		X:    x,
		Y:    y,
		Char: rune(ch),
		Fg:   fg,
		Bg:   bg,
	}
}

func (c *Container) AddTable(cols []string, rows []TableRow, widths []int, width int) {
	for k, v := range cols {
		width := widths[k]
		c.Add(NewWord(v, width))
		c.Add(NewWord(" ", 1))

	}
	c.Add(NewLineBreak())

	for k, row := range rows {
		for _, cell := range row {
			width := widths[k]
			c.Add(NewWord(cell, width))
			c.Add(NewWord(" ", 1))
		}
		c.Add(NewLineBreak())
	}
	c.Add(NewLineBreak())
}

func appendRunePosMatrix(m1, m2 []RunePos) []RunePos {
	for _, v := range m2 {
		m1 = append(m1, v)
	}
	return m1
}

func appendRunePos(m []RunePos, p RunePos) []RunePos {
	m = append(m, p)
	return m
}

func addConstant(m []RunePos, x, y int) []RunePos {
	for k, v := range m {
		m[k].X = v.X + x
		m[k].Y = v.Y + y
	}
	return m
}
