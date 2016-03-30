package main

import "github.com/nsf/termbox-go"

type Table struct {
	Cols       []string
	Rows       []*TableRow
	ColWidths  []int
	state      []RunePos
	Dimensions *Dimensions
}

type TableRow struct {
	Cells []string
	Fg    termbox.Attribute
	Bg    termbox.Attribute
	Group string
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
	return t.Dimensions
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

func NewTable(dm *Dimensions, cols []string, widths []int) *Table {
	return &Table{
		Cols:       cols,
		ColWidths:  widths,
		Rows:       make([]*TableRow, 0),
		Dimensions: dm,
	}
}
