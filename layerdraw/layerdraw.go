package layerdraw

import (
	"github.com/nsf/termbox-go"
)

type Layer struct {
	added    int
	Elements []*Element
}

type Element struct {
	X, Y, Width, Height int
	Contents            VisibleElement
}

type VisibleElement interface {
	matrix() []runeMatrix
}

type SelectableElement interface {
	selected()
}

type Word struct {
	WordString string
	X, Y       int
}

type Table struct {
	Cols      []string
	Rows      []TableRow
	ColWidths []int
}

type TableRow struct {
	Row []string
}

func (tr *TableRow) selected() {

}

func NewLayer() *Layer {
	els := make([]*Element, 0, 10)
	return &Layer{
		Elements: els,
	}
}

func NewWord(word string, x, y int) *Word {
	return &Word{
		WordString: word,
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

func NewElement(x, y, width, height int, contents VisibleElement) *Element {
	return &Element{
		X:        x,
		Y:        y,
		Width:    width,
		Height:   height,
		Contents: contents,
	}
}

func (l *Layer) Add(el *Element) {
	if len(l.Elements) == cap(l.Elements) {
		t := make([]*Element, len(l.Elements), (cap(l.Elements)+1)*2)
		copy(t, l.Elements)
		l.Elements = t
	}

	l.Elements = l.Elements[0 : len(l.Elements)+1]
	l.Elements[len(l.Elements)-1] = el

}

func (l *Layer) Draw() {
	for _, v := range l.Elements {
		runes := v.Contents.matrix()
		for _, e := range runes {
			termbox.SetCell(e.X+v.X,
				e.Y+v.Y,
				e.Char,
				e.Fg,
				e.Bg)
		}
	}
}

func (w Word) matrix() []runeMatrix {
	matrc := make([]runeMatrix, len(w.WordString))
	for i := 0; i < len(w.WordString); i++ {
		matrc[i] = NewRuneMatrix(w.X+i, w.Y, w.WordString[i], termbox.ColorDefault, termbox.ColorDefault)
	}
	return matrc
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

func (t Table) matrix() []runeMatrix {
	matrc := make([]runeMatrix, 5)
	c := 0
	for e := 0; e < len(t.Cols); e++ {
		width := t.ColWidths[e]
		for i := 0; i < width; i++ {
			chru := byte(' ')
			if i < len(t.Cols[e]) {
				chru = t.Cols[e][i]
			}
			matrc = append(matrc, NewRuneMatrix(c, 0, chru, termbox.ColorDefault, termbox.ColorDefault))
			c++
		}
		c++
		matrc = append(matrc, NewRuneMatrix(c, 0, '\t', termbox.ColorDefault, termbox.ColorDefault))

	}
	c = 0
	for m := 0; m < len(t.Rows); m++ {
		for n := 0; n < len(t.Rows[m].Row); n++ {
			width := t.ColWidths[n]
			for s := 0; s < width; s++ {
				chru := byte(' ')
				if s < len(t.Rows[m].Row[n]) {
					chru = t.Rows[m].Row[n][s]
				}
				matrc = append(matrc, NewRuneMatrix(c, m+1, chru, termbox.ColorDefault, termbox.ColorDefault))
				c++
			}
			c++
			matrc = append(matrc, NewRuneMatrix(c, m+1, '\t', termbox.ColorDefault, termbox.ColorDefault))
		}
		c = 0
	}
	return matrc
}

type runeMatrix struct {
	X, Y int
	Char rune
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}
