package layerdraw

import (
	"github.com/nsf/termbox-go"
)

type Layer struct {
	added    int
	Elements []Element
}

type Element struct {
	X, Y, Width, Height int
	Contents            VisibleElement
}

type VisibleElement interface {
	getMatrix() []runeMatrix
	preserveState()
	cleanup() []runeMatrix
}

type SelectableElement interface {
	selected()
}

type Word struct {
	WordString  string
	Width, X, Y int
	state       []runeMatrix
}

type Table struct {
	Cols      []string
	Rows      []TableRow
	ColWidths []int
	state     []runeMatrix
}

type TableRow struct {
	Row []string
	Id  string
}

func (tr *TableRow) selected() {

}

func NewLayer() *Layer {
	els := make([]Element, 0, 10)
	return &Layer{
		Elements: els,
	}
}

func NewWord(word string, width, x, y int) Word {
	return Word{
		WordString: word,
		Width:      width,
		X:          x,
		Y:          y,
	}
}

func NewTable(cols []string, rows []TableRow, widths []int) *Table {
	return &Table{
		Cols:      cols,
		Rows:      rows,
		ColWidths: widths,
	}
}

func NewElement(x, y, width, height int, contents VisibleElement) Element {
	return Element{
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Contents: contents,
	}
}

func (l *Layer) Add(el Element) {
	if len(l.Elements) == cap(l.Elements) {
		t := make([]Element, len(l.Elements), (cap(l.Elements)+1)*2)
		copy(t, l.Elements)
		l.Elements = t
	}

	l.Elements = l.Elements[0 : len(l.Elements)+1]
	l.Elements[len(l.Elements)-1] = el

}

func (l *Layer) Draw() {
	for _, v := range l.Elements {
		runes := v.Contents.getMatrix()
		clean := v.Contents.cleanup()
		for _, e := range clean {
			termbox.SetCell(e.X+v.X, e.Y+v.Y, ' ', termbox.ColorDefault, termbox.ColorDefault)
		}
		v.Contents.preserveState()
		for _, e := range runes {
			termbox.SetCell(e.X+v.X,
				e.Y+v.Y,
				e.Char,
				e.Fg,
				e.Bg)
		}
	}
}

func (w *Word) getMatrix() []runeMatrix {
	matrix := make([]runeMatrix, w.Width)
	for i := 0; i < w.Width; i++ {
		chru := byte(' ')
		if i < len(w.WordString) {
			chru = w.WordString[i]
		}
		matrix[i] = NewRuneMatrix(w.X+i, w.Y, chru, termbox.ColorDefault, termbox.ColorDefault)
	}
	return matrix
}

func appendMatrix(m1, m2 []runeMatrix) []runeMatrix {
	for _, v := range m2 {
		m1 = append(m1, v)
	}
	return m1
}

func (t *Table) getMatrix() []runeMatrix {
	// clear previous rows

	matrix := make([]runeMatrix, 5)
	c := 0
	for e := 0; e < len(t.Cols); e++ {
		width := t.ColWidths[e]
		w := NewWord(t.Cols[e], width, c, 0)
		c = c + width
		matrix = appendMatrix(matrix, w.getMatrix())
		matrix = append(matrix, NewRuneMatrix(c, 0, '\t', termbox.ColorDefault, termbox.ColorDefault))
		c++
	}
	c = 0
	for m := 0; m < len(t.Rows); m++ {
		for n := 0; n < len(t.Rows[m].Row); n++ {
			width := t.ColWidths[n]
			w := NewWord(t.Rows[m].Row[n], width, c, m+1)
			c = c + width
			matrix = appendMatrix(matrix, w.getMatrix())
			matrix = append(matrix, NewRuneMatrix(c, m+1, '\t', termbox.ColorDefault, termbox.ColorDefault))
			c++
		}
		c = 0
	}
	return matrix
}

func (t *Table) preserveState() {
	t.state = t.getMatrix()
}

func (t *Table) cleanup() []runeMatrix {
	matrix := make([]runeMatrix, 0)
	for _, v := range t.state {
		if v.Y == 0 {
			continue
		}
		matrix = append(matrix, NewRuneMatrix(v.X, v.Y, ' ', 0, 0))

	}
	return matrix
}

func NewRuneMatrix(x, y int, ch byte, fg, bg termbox.Attribute) runeMatrix {
	return runeMatrix{
		X:    x,
		Y:    y,
		Char: rune(ch),
		Fg:   fg,
		Bg:   bg,
	}
}

type runeMatrix struct {
	X, Y int
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}
