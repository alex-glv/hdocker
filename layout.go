package main

import (
	"encoding/json"
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
