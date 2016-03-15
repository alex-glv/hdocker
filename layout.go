package main

import (
	"encoding/json"
	"io/ioutil"
)

type Column struct {
	Width   int    `json: "Width"`
	Title   string `json: "Title"`
	Data    string `json: "Data"`
	WordRef *Word
}

func ParseLayout(jsonStr []byte) []Column {
	var dat []Column
	// fmt.Println(string(jsonStr))
	if err := json.Unmarshal(jsonStr, &dat); err != nil {
		panic(err)
	}
	return dat
}

func Createlayout(infoBox *Container) []Column {
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
