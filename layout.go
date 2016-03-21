package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var logFile, _ = os.Create("/tmp/log.out")
var logger = log.New(logFile, "", 0)

type Column struct {
	Width   float32 `json: "Width"`
	Title   string  `json: "Title"`
	Data    string  `json: "Data"`
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
	logger.Println("Init CreateLayout with: W:", infoBox.Width, ", Start:", infoBox.X)
	columns := ParseLayout(dat)
	curFill := float32(0)
	curValues := make([]int, 0, 0)
	logger.Println("Columns:")
	logger.Println(columns)
	for i, v := range columns {
		logger.Println("Adding:", v.Title, ", w:", int(v.Width*float32(infoBox.Width)))
		infoBox.Add(NewWordDef(v.Title, int(v.Width*float32(infoBox.Width))))
		infoBox.Add(Space())
		curValues = append(curValues, i)
		curFill = curFill + v.Width
		logger.Println("Fill:", curFill)
		if curFill == 1 || i == len(columns)-1 {
			infoBox.Add(LineBreak())
			for _, ci := range curValues {
				logger.Println("Dumping:", columns[ci].Data, ", w:", int(columns[ci].Width*float32(infoBox.Width)))
				// logger.Println(ci)

				columns[ci].WordRef = NewWordDef(columns[ci].Data, int(columns[ci].Width*float32(infoBox.Width)))
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
