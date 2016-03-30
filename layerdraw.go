package main

import (
	"github.com/nsf/termbox-go"
)

type Layer struct {
	added      int
	Containers []LayerElement
	buff       []RunePos
}

type Container struct {
	Dimensions        *Dimensions
	ContainerElements []*ContainerElement
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

func NewContainer(dm *Dimensions) *Container {
	return &Container{
		Dimensions:        dm,
		ContainerElements: make([]*ContainerElement, 0),
	}
}

func (c *Container) Add(el VisibleElement) {
	cel := newContainerElement(el)
	c.ContainerElements = append(c.ContainerElements, cel)
}

// LayerElement implementation
func (c *Container) getDimensions() *Dimensions {
	return c.Dimensions
}

func (c *Container) getElements() []*ContainerElement {
	return c.ContainerElements
}

//

func (c *Container) EmptyRunePos() RunePos {
	return NewRunePos(c.Dimensions.X, c.Dimensions.Y, 0, 0, 0)
}

func (l *Layer) Add(el LayerElement) {
	l.Containers = append(l.Containers, el)

}

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

func (l *Layer) Flush() {
	for _, v := range l.buff {
		termbox.SetCell(v.X, v.Y, v.Char, v.Fg, v.Bg)
	}
	termbox.Flush()
}

func (l *Layer) GetBuff() []RunePos {
	return l.buff
}

func (l *Layer) Draw() {
	l.buff = make([]RunePos, 0)
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
					l.buff = append(l.buff, e)
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
