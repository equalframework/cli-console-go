package parser

import (
	"encoding/json"
	"strconv"
)

var LogLevelOrderRelation map[string]int = map[string]int{
	"DEBUG" : 0,
	"INFO" : 1,
	"WARNING" : 2,
	"ERROR"	: 3,
}

type LogLineArray []*LogLine

type LogLine struct {
	ThreadID	string 	`json:"thread_id"`
	Time		string 	`json:"time"`
	Mtime		string	`json:"mtime"`
	Level		string	`json:"level"`
	Mode		string	`json:"mode"`
	Class		string	`json:"class"`
	Function	string	`json:"function"`
	File		string	`json:"file"`
	Line		int		`json:"line"`
	Message		string	`json:"message"`
	Stack		[]StackItem	`json:"stack"`
}

type StackItem struct {
	File 		string		`json:"file"`
	Function 	string		`json:"function"`
	Line 		int			`json:"line"`
	Args 		[]string	`json:"args"`
	Class 		string		`json:"class"`
	Type		string		`json:"type"`
}

func ParseLogLine(line []byte) *LogLine {
	ret := new(LogLine)
	json.Unmarshal(line,ret)
	return ret
}

func (lla LogLineArray) GetHighestLogLevel() string {
	level := "DEBUG"
	level_int := 0
	for _,item := range lla {
		val,ok := LogLevelOrderRelation[(*item).Level]
		if !ok {
			return (*item).Level
		}
		if val > level_int {
			level_int = val
			level = (*item).Level
		}
	}
	return level
}

func (lla LogLineArray) GetBigestMtime() int {
	n,err := strconv.Atoi(lla[0].Mtime)
	if err != nil {
		panic(err)
	}
	smallest := n
	for i := 1 ; i < len(lla) ; i++ {
		n,err = strconv.Atoi(lla[i].Mtime)
		if err != nil {
			panic(err)
		}
		if smallest < n {
			smallest = n
		}
	}
	return smallest
}

func (lla LogLineArray) GetLastMessage() string {
	if len(lla) > 0 {
		return lla[len(lla)-1].Message
	}
	return ""
}