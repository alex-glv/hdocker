package layerdraw

import (
	// "fmt"
	"github.com/nsf/termbox-go"
)

var DynamicContainer = 0x2
var DEFAULT_GROUP = "default"

type Layer struct {
	added      int
	Containers []*Container
}

type Container struct {
	X, Y, Width, Height int
	LineBreaksCount     int
	LastRune            RunePos
	ContainerElements   map[string][]*ContainerElement
	groupsOrder         []string
	groupsIndices       map[string]int
	currentGroup        string
}

type Drawable interface {
	Draw()
}

type VisibleElement interface {
	getMatrix() []RunePos
}

type ContainerElement struct {
	RuneMatrixPos []RunePos
	Group         string
	Options       int
	Element       VisibleElement
}

type SelectableElement interface {
	getGroup() string
}

type Word struct {
	WordString string
	Width      int
}

type LineBreakType struct{}

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

type TableRow struct {
	Cells []string
}

func NewLayer() *Layer {
	els := make([]*Container, 0)
	return &Layer{
		Containers: els,
	}
}

func NewContainerElement(el VisibleElement) *ContainerElement {
	return &ContainerElement{
		Element: el,
	}
}

func NewWord(word string, width int) *Word {
	return &Word{
		WordString: word,
		Width:      width,
	}
}

func LineBreak() *LineBreakType {
	return &LineBreakType{}
}

func NewContainer(x, y, width, height int) *Container {
	return &Container{
		X:                 x,
		Y:                 y,
		LastRune:          NewRunePos(x, y, 0, 0, 0),
		LineBreaksCount:   0,
		Width:             width,
		Height:            height,
		currentGroup:      DEFAULT_GROUP,
		groupsOrder:       make([]string, 0),
		groupsIndices:     make(map[string]int),
		ContainerElements: make(map[string][]*ContainerElement),
	}
}

func UpdateTableRow(hash string, fields ...string) {

}

func NewTableRow(fields ...string) *TableRow {
	row := &TableRow{
		Cells: fields,
	}

	return row
}

func (l *Layer) Add(el *Container) {
	l.Containers = append(l.Containers, el)

}
func (c *Container) Add(el VisibleElement) {
	cel := NewContainerElement(el)
	cel.Group = c.currentGroup

	matrix := el.getMatrix()
	if len(matrix) == 0 {
		c.LastRune = NewRunePos(c.X, c.Y+c.LineBreaksCount, ' ', 0, 0)
		c.LineBreaksCount = c.LineBreaksCount + 1
	} else {
		matrix = addConstant(matrix, c.LastRune.X+1, c.LastRune.Y)
		cel.RuneMatrixPos = matrix
		c.LastRune = matrix[len(matrix)-1]
		_, exists := c.ContainerElements[c.currentGroup]
		if !exists {
			c.ContainerElements[c.currentGroup] = make([]*ContainerElement, 0)
		}
	}
	c.ContainerElements[c.currentGroup] = append(c.ContainerElements[c.currentGroup], cel)
}

func (c *Container) StartGroup(hash string) {
	c.currentGroup = hash
	c.groupsOrder = append(c.groupsOrder, hash)
	c.groupsIndices[hash] = len(c.groupsOrder) - 1
}

func (c *Container) StopGroup() {
	c.currentGroup = DEFAULT_GROUP
}

func (c *Container) Reset() {
	c.ContainerElements = make(map[string][]*ContainerElement)
	c.LastRune = NewRunePos(c.X, c.Y, ' ', 0, 0)
	c.LineBreaksCount = 0
}

func (c *Container) DeleteGroup(hash string) {
	if _, e := c.ContainerElements[hash]; e {
		delete(c.ContainerElements, hash)
	}
	i := c.groupsIndices[hash]
	c.groupsOrder = append(c.groupsOrder[0:i], c.groupsOrder[i+1:len(c.groupsOrder)]...)
	delete(c.groupsIndices, hash)
}

func (c *Container) RecalculateRunes() {
	c.LineBreaksCount = 0
	c.LastRune = NewRunePos(c.X, c.Y, ' ', 0, 0)
	for _, e := range c.groupsOrder {
		group := c.ContainerElements[e]
		for _, v := range group {
			matrix := v.Element.getMatrix()
			if len(matrix) == 0 {
				c.LastRune = NewRunePos(c.X, c.Y+c.LineBreaksCount, ' ', 0, 0)
				c.LineBreaksCount = c.LineBreaksCount + 1
			} else {
				matrix = addConstant(matrix, c.LastRune.X+1, c.LastRune.Y)
				v.RuneMatrixPos = matrix
				c.LastRune = matrix[len(matrix)-1]
			}
		}
	}
}

func (c *Container) Draw() {
	for x := 0; x < c.Width; x++ { // cleanup
		for y := 0; y < c.Height; y++ {
			termbox.SetCell(c.X+x, c.Y+y, 0, 0, 0)
		}
	}
	for _, group := range c.ContainerElements {
		for _, v := range group {
			for _, e := range v.RuneMatrixPos {
				termbox.SetCell(e.X,
					e.Y,
					e.Char,
					e.Fg,
					e.Bg)
			}
		}

	}
	c.Reset()

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

func (l *LineBreakType) getMatrix() []RunePos {
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

func Space() *Word {
	return NewWord(" ", 1)
}

type TableCols map[string]int

func NewTable(cols []string, widths []int) *Table {
	return &Table{
		Cols:      cols,
		ColWidths: widths,
	}
}

func (c *Container) AddTableHeader(t *Table) {
	var width int
	for k, v := range t.Cols {
		width = t.ColWidths[k]
		c.Add(NewWord(v, width))
		c.Add(Space())

	}
	c.Add(LineBreak())
}

func UpdateWord(w *Word, ws string, wl int) {
	w.WordString = ws
	w.Width = wl
}

func (c *Container) AddTableRows(t *Table, rows []*TableRow) {
	// var firstRowWord *Word
	// var lastRowWord *Word
	var width int
	// c.Add(NewWord(fmt.Sprintf("%d", len(rows)), 100))
	// c.Add(LineBreak())
	for _, row := range rows {
		for k, cell := range row.Cells {
			width = t.ColWidths[k]
			c.Add(NewWord(cell, width))
			c.Add(Space())
		}
		c.Add(LineBreak())
	}
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
