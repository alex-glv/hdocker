package main

import (
	"github.com/nsf/termbox-go"
)

var DEFAULT_GROUP = "default"

type Layer struct {
	added      int
	Containers []LayerElement
}

type Container struct {
	X, Y, Width, Height int
	ContainerElements   []*ContainerElement
}

type Dimensions struct {
	X, Y, Width, Height int
}

type LayerElement interface {
	getElements() []*ContainerElement
	getDimensions() *Dimensions
}

type VisibleElement interface {
	getMatrix() []RunePos
}

type ContainerElement struct {
	runeMatrixPos []RunePos

	Options int
	Element VisibleElement
}

type Word struct {
	WordString string
	Width      int
	Fg         termbox.Attribute
	Bg         termbox.Attribute
}

type LineBreakType struct{}

type RunePos struct {
	X, Y int
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

type Table struct {
	Cols                []string
	Rows                []*TableRow
	ColWidths           []int
	state               []RunePos
	X, Y, Width, Height int
}

type TableRow struct {
	Cells []string
	Fg    termbox.Attribute
	Bg    termbox.Attribute
	Group string
}

func NewLayer() *Layer {
	els := make([]LayerElement, 0)
	return &Layer{
		Containers: els,
	}
}

func newContainerElement(el VisibleElement) *ContainerElement {
	return &ContainerElement{
		Element: el,
	}
}

func NewWord(word string, width int, fg termbox.Attribute, bg termbox.Attribute) *Word {
	return &Word{
		WordString: word,
		Width:      width,
		Fg:         fg,
		Bg:         bg,
	}
}

func NewWordDef(word string, width int) *Word {
	return NewWord(word, width, 0, 0)
}

func LineBreak() *LineBreakType {
	return &LineBreakType{}
}

func NewContainer(x, y, width, height int) *Container {
	return &Container{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,

		ContainerElements: make([]*ContainerElement, 0),
	}
}

func NewTableRow(fields ...string) *TableRow {
	row := &TableRow{
		Cells: fields,
	}

	return row
}

func (l *Layer) Add(el LayerElement) {
	l.Containers = append(l.Containers, el)

}
func (c *Container) Add(el VisibleElement) {
	cel := newContainerElement(el)
	c.ContainerElements = append(c.ContainerElements, cel)
}

// LayerElement implementation
func (c *Container) getDimensions() *Dimensions {
	return &Dimensions{
		X:      c.X,
		Y:      c.Y,
		Height: c.Height,
		Width:  c.Width,
	}
}

func (c *Container) getElements() []*ContainerElement {
	return c.ContainerElements
}

//

func (l *Layer) RecalculateRunes(c LayerElement) []*ContainerElement {
	vsbels := make([]*ContainerElement, 0)
	lineBreaksCount := 0
	dims := c.getDimensions()
	lastRune := NewRunePos(dims.X-1, dims.Y, ' ', 0, 0) //
	for _, v := range c.getElements() {
		matrix := v.Element.getMatrix()
		if len(matrix) == 0 {
			lineBreaksCount++
			lastRune = NewRunePos(dims.X-1, dims.Y+lineBreaksCount, ' ', 0, 0)

		} else {
			matrix = addConstant(matrix, lastRune.X+1, lastRune.Y)
			v.runeMatrixPos = matrix
			lastRune = matrix[len(matrix)-1]
		}
		vsbels = append(vsbels, v)
	}
	return vsbels
}

func (c *Container) EmptyRunePos() RunePos {
	return NewRunePos(c.X, c.Y, 0, 0, 0)
}

func (l *Layer) Draw() {

	for _, c := range l.Containers {
		dims := c.getDimensions()
		vsbels := l.RecalculateRunes(c)
		for x := 0; x < dims.Width; x++ { // cleanup
			for y := 0; y < dims.Height; y++ {
				termbox.SetCell(dims.X+x, dims.Y+y, ' ', 0, 0)
			}
		}
		for _, v := range vsbels {
			for _, e := range v.runeMatrixPos {

				if e.X <= dims.Width+dims.X && e.Y <= dims.Height+dims.Y {
					termbox.SetCell(
						e.X,
						e.Y,
						e.Char,
						e.Fg,
						e.Bg,
					)
				}
			}

		}
	}
}

func (w *Word) getMatrix() []RunePos {
	matrix := make([]RunePos, w.Width)
	for i := 0; i < w.Width; i++ {
		chru := byte(' ')
		if i < len(w.WordString) {
			chru = w.WordString[i]
		}
		matrix[i] = NewRunePos(i, 0, chru, w.Fg, w.Bg)
	}
	return matrix
}

func (l *LineBreakType) getMatrix() []RunePos {
	runePosMatrix := make([]RunePos, 0)
	return runePosMatrix
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
	return NewWordDef(" ", 1)
}

func (t *Table) AddRow(fields ...string) *TableRow {
	row := &TableRow{
		Cells: fields,
	}
	t.Rows = append(t.Rows, row)

	return row
}

func (t *Table) DeleteRow(group string) {
	var rowIndex int = -1
	if len(t.Rows) == 0 {
		return
	}
	for i, row := range t.Rows {
		if row.Group == group {
			rowIndex = i
			break
		}
	}

	if rowIndex == -1 {
		return
	}

	t.Rows = append(t.Rows[:rowIndex], t.Rows[rowIndex+1:]...)
}

func (t *Table) SetColWidth(widths []int) {
	if len(widths) != len(t.Cols) {
		panic("widths is not correct for the table")
	}
	t.ColWidths = widths
}

func (t *Table) genTable() []VisibleElement {
	elements := make([]VisibleElement, 0)
	for i, col := range t.Cols {
		elements = append(elements, NewWordDef(col, t.ColWidths[i]), Space())
	}

	elements = append(elements, LineBreak())

	for _, row := range t.Rows {
		for y, word := range row.Cells {
			elements = append(elements, NewWord(word, t.ColWidths[y], row.Fg, row.Bg), Space())
		}
		elements = append(elements, LineBreak())
	}

	return elements

}

// LayerElement implementation

func (t *Table) getDimensions() *Dimensions {
	return &Dimensions{
		X:      t.X,
		Y:      t.Y,
		Height: t.Height,
		Width:  t.Width,
	}
}

func (t *Table) getElements() []*ContainerElement {
	var cels []*ContainerElement
	for _, el := range t.genTable() {
		contel := newContainerElement(el)
		cels = append(cels, contel)
	}

	return cels
}

//

func NewTable(x, y, width, height int, cols []string, widths []int) *Table {
	return &Table{
		Cols:      cols,
		ColWidths: widths,
		Rows:      make([]*TableRow, 0),
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
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
