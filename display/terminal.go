package display

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/equalframework/cli-console-go/parser"
	"github.com/equalframework/cli-console-go/syscalls"

	"golang.org/x/exp/slices"
	"golang.org/x/term"
)

const VERSION = "v1.1"


func clamp(min,max,value int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

const CLEARANCE = 8

var Filter_type []string = []string{"ALL","ERROR","WARNING","INFO","DEBUG"}

var LogLevelColorSelected map[string]string = map[string]string{
	"INFO" : "\u001b[0;104m",
	"DEBUG" : "\u001b[0;102m",
	"WARNING" : "\u001b[1;103m",
	"ERROR"	: "\u001b[1;101m",
}

var LogLevelColor map[string]string = map[string]string{
	"ALL" : "\u001b[0;0m",
	"INFO" : "\u001b[0;34m",
	"DEBUG" : "\u001b[0;32m",
	"WARNING" : "\u001b[1;33m",
	"ERROR"	: "\u001b[1;31m",
}

var LogLevelDisp map[string]string = map[string]string{
	"INFO" : 	"ℹ️   INFO",
	"DEBUG" : 	"🪲  DEBUG",
	"WARNING" : "🚨WARNING",
	"ERROR"	: 	"❌  ERROR",
}

type Displayer struct {
	data 		 *map[int64]*map[string]*parser.LogLineArray
	Filename 	 string
	SelectedLine [3]int
	CurrentSkip	 [3]int
	W 			 int
	H 			 int
	SelectedThread *parser.LogLineArray
	SelectedLL	   *parser.LogLine
	ContextLevel	int
	FilterTypeIndex	int
}


func (d *Displayer) RefreshSize() (hasChanged bool) {
	width,height,err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Fprintln(os.Stderr,"Failed to get terminal size : "+err.Error())
		os.Exit(1)
	}
	ret := d.W != width || d.H != height
	d.W = width
	d.H = height
	return ret
}

func NewDisplayer(filename string) Displayer {
	return Displayer{
		data : new(map[int64]*map[string]*parser.LogLineArray) , 
		Filename: filename,
		SelectedLine: [3]int{0,0,0},
		SelectedThread: nil,
		ContextLevel: 0,
	}
}

func (d Displayer) Lenght() int {
	switch d.ContextLevel {
	case 0:
		sum := 0
		for _,item := range *d.data {
			for _,tem := range *item {
				if Filter_type[d.FilterTypeIndex] == "ALL" || tem.GetHighestLogLevel() == Filter_type[d.FilterTypeIndex] {
					sum ++
				}
			}
		}
		return sum
	case 1:
		return len(*d.SelectedThread)
	default:
		return 0
	}

}

func (d *Displayer) PrintResult() {
	count := 0
	show := ""
	if(d.W < 120 || d.H < 30) {
		syscalls.CallClear()
		println("Terminal size should be at least 120x30.\nPlease resize your terminal")
		return
	}
	if d.CurrentSkip[0] + d.H - CLEARANCE - 1 < d.SelectedLine[0] { 
		d.CurrentSkip[0] = d.SelectedLine[0] - d.H + CLEARANCE +1
	}
	if d.CurrentSkip[0] > d.SelectedLine[0] {
		d.CurrentSkip[0] = d.SelectedLine[0]
	}
	if d.CurrentSkip[0] < 0 {
		d.CurrentSkip[0] = 0
	}
	switch d.ContextLevel {
	case 0:
		l := d.Lenght()
		l1 := l - (d.CurrentSkip[0] + d.H - CLEARANCE - 1)
		if l1 < 0 {
			l1 = 0
		}
		l2 := l - d.CurrentSkip[0]
		if l2 > l {
			l2 = l
		}
		show += fmt.Sprintf("eQual Logger %v  | show %v%v\u001b[0;0m | lines %v-%v of %v\n",VERSION,LogLevelColor[Filter_type[d.FilterTypeIndex]],Filter_type[d.FilterTypeIndex],l1,l2,l)
		for i := 0 ; i < d.W ; i++ {
			show += fmt.Sprint("-")
		}
		show += fmt.Sprint("\n")
		show += fmt.Sprintln(" Level          Thread ID   Time                   Last log")
		show += fmt.Sprintln()
		count = d.PrintStep1(CLEARANCE,&show) 
		break
	case 1:
		l := d.Lenght()
		t,err := time.Parse(time.RFC3339,(*(*d.SelectedThread)[0]).Time)
		if err != nil {
			panic(err)
		}
		l1 := l - (d.CurrentSkip[0] + d.H - CLEARANCE - 1)
		if l1 < 0 {
			l1 = 0
		}
		l2 := l - d.CurrentSkip[0]
		if l2 > l {
			l2 = l
		}
		show += fmt.Sprintf("eQual Logger %v > %v@%v |  lines %v-%v of %v\n",VERSION,(*(*d.SelectedThread)[0]).ThreadID,strings.ReplaceAll(t.Format(time.DateTime)," ","_"),l1,l2,l)
		for i := 0 ; i < d.W ; i++ {
			show += fmt.Sprint("-")
		}
		show += fmt.Sprintln()
		show += fmt.Sprintln()
		show += fmt.Sprint("\n")
		count = d.PrintStep2(CLEARANCE,&show)
		break
	case 2:
		t,err := time.Parse(time.RFC3339,d.SelectedLL.Time)
		tf := strings.ReplaceAll(t.Format(time.DateTime)," ","_")
		if err != nil {
			panic(err)
		}
		line := strconv.Itoa(d.SelectedLL.Line)
		file := d.SelectedLL.File
		if len(file) > d.W-30-len(VERSION)-len(line)-len(tf)-len(d.SelectedLL.Function) {
			file = file[:d.W-33-len(VERSION)-len(line)-len(tf)-len(d.SelectedLL.Function)]+"..."
		}
		show += fmt.Sprintf("eQual Logger %v > %v@%v > %v@%v:%v\n",VERSION,d.SelectedLL.ThreadID,tf,d.SelectedLL.Function,file,line)
		for i := 0 ; i < d.W ; i++ {
			show += fmt.Sprint("-")
		}
		show += fmt.Sprint("\n")
		count = d.PrintStep3(CLEARANCE,&show)
	}
	for ;count < d.H-CLEARANCE;count++ {
		show += fmt.Sprintln()
	}
	show += fmt.Sprint("\n\u001b[0;107m\u001b[1;90mPAGEUP\u001b[0;0m Go to top \u001b[0;107m\u001b[1;90mPAGEDOWN\u001b[0;0m Go to BOTTOM \u001b[0;107m\u001b[1;90mENTER\u001b[0;0m Open context \u001b[0;107m\u001b[1;90mBKSP\u001b[0;0m Close context \u001b[0;107m\u001b[1;90m Q \u001b[0;0m Exit \u001b[0;107m\u001b[1;90mTAB\u001b[0;0m Change filter \u001b[0;107m\u001b[1;90m ^R \u001b[0;0m Reload File\n")
	syscalls.CallClear()
	fmt.Print(show)
}

func (d *Displayer) ReadContent() {
	(*d.data) = make(map[int64]*map[string]*parser.LogLineArray)
	f,err := os.Open(d.Filename)
	if err != nil {
		fmt.Fprintf(os.Stderr,"Failed to open %v : %v ",os.Args[1],err.Error())
		os.Exit(1)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if(len(scanner.Bytes()) <= 0) {
			continue
		}
		ll := parser.ParseLogLine(scanner.Bytes())
		t,err := time.Parse(time.RFC3339,ll.Time)
		if err != nil {
			panic(err)
		}
		_,ok := (*d.data)[t.Unix()]
		if !ok {
			(*d.data)[t.Unix()] = new(map[string]*parser.LogLineArray)
			*(*d.data)[t.Unix()] = make(map[string]*parser.LogLineArray)
		}
		_,ok = (*(*d.data)[t.Unix()])[ll.ThreadID]
		if !ok {
			(*(*d.data)[t.Unix()])[ll.ThreadID] = new(parser.LogLineArray)
		}
		(*(*(*d.data)[t.Unix()])[ll.ThreadID]) = append((*(*(*d.data)[t.Unix()])[ll.ThreadID]), ll)
	}
}

func (d *Displayer) SelectCurrent() {
	switch d.ContextLevel {
	case 0:
		i := 0
		keys := make([]int, len(*d.data))
		for k := range *d.data {
			keys[i] = int(k)
			i++
		}
		sort.Ints(keys)
		count := 0
		for i := len(keys)-1 ; i >= 0 ; i-- {
			arr := (*(*d.data)[int64(keys[i])])
			scanned_keys := []string{}
			for len(scanned_keys) < len(arr) {
				bestKey := ""
				bestMtime := 0
				for key,item := range arr {
					Mtime := item.GetBigestMtime()
					if slices.Contains(scanned_keys,key) {
						continue
					}
					if Filter_type[d.FilterTypeIndex] != "ALL" && Filter_type[d.FilterTypeIndex]!=item.GetHighestLogLevel() {
						scanned_keys = append(scanned_keys, key)
						continue
					}
					if(bestKey=="" || bestMtime < Mtime) {
						bestKey = key
						bestMtime = Mtime
					}
				}
				if len(bestKey) <= 0 {
					continue
				}
				if(count >= d.CurrentSkip[0]) {
					if(count-d.SelectedLine[0] == 0) {
						d.SelectedThread = arr[bestKey]
						d.ContextLevel = 1
						return
					}
				}
				scanned_keys = append(scanned_keys, bestKey)
				count ++
			}
		}
		return
	case 1 :
		for i,item := range *d.SelectedThread {
			if(len(*d.SelectedThread)-d.SelectedLine[1]-1 == i) {
				d.SelectedLL = item
				d.ContextLevel = 2
				return
			}
		}
	}
}

func (d *Displayer) PrintStep1(clearance int,show *string) int {
	i := 0
	keys := make([]int, len(*d.data))
	for k := range *d.data {
		keys[i] = int(k)
		i++
	}
	sort.Ints(keys)
	
	count := 0
	toPrint := []string{}
	for i := len(keys)-1 ; i >= 0 && count - d.CurrentSkip[0] < d.H - clearance ; i-- {
		arr := (*(*d.data)[int64(keys[i])])
		scanned_keys := []string{}
		for len(scanned_keys) < len(arr) && count - d.CurrentSkip[0] < d.H - clearance {
			bestKey := ""
			bestMtime := 0
			for key,item := range arr {
				Mtime := item.GetBigestMtime()
				if slices.Contains(scanned_keys,key) {
					continue
				}
				if Filter_type[d.FilterTypeIndex] != "ALL" && Filter_type[d.FilterTypeIndex]!=item.GetHighestLogLevel() {
					scanned_keys = append(scanned_keys, key)
					continue
				}
				if(bestKey=="" || bestMtime < Mtime) {
					bestKey = key
					bestMtime = Mtime
				}
			}
			if len(bestKey) <= 0 {
				continue
			}
			if(count >= d.CurrentSkip[0]) {
				msg  := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(arr[bestKey].GetLastMessage(),"\n",""),"\t","	"),"  "," "),"  "," ")
				msg = msg[:clamp(0,len(msg),d.W-52)]
				if(count-d.SelectedLine[0] == 0) {
					toPrint = append(toPrint,fmt.Sprintf(" %v[%v]    %v    %v    %v\u001b[0;0m \n",
						LogLevelColorSelected[arr[bestKey].GetHighestLogLevel()],
						LogLevelDisp[arr[bestKey].GetHighestLogLevel()],
						bestKey,
						time.Unix(int64(keys[i]),0).Format(time.DateTime),
						msg,
					))
				} else {
					toPrint = append(toPrint,fmt.Sprintf(" %v[%v]    %v    %v    %v\u001b[0;0m\n",
						LogLevelColor[arr[bestKey].GetHighestLogLevel()],
						LogLevelDisp[arr[bestKey].GetHighestLogLevel()],
						bestKey,
						time.Unix(int64(keys[i]),0).Format(time.DateTime),
						msg,
					))
				}
			}
			scanned_keys = append(scanned_keys, bestKey)
			count ++
		}
	}
	for i := len(toPrint)-1 ; i >= 0 ; i-- {
		*show += fmt.Sprint(toPrint[i])
	}
	return count - d.CurrentSkip[0]
}

func (d *Displayer) PrintStep2(clearance int,show *string) int {
	count := 0
	for i,item := range *d.SelectedThread {
		msg  := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(item.Message,"\n",""),"\t","	"),"  "," "),"  "," ")
		msg = msg[:clamp(0,len(msg),d.W-7)]
		var color string
		if(len(*d.SelectedThread)-d.SelectedLine[1]-1 == i) {
			color = LogLevelColorSelected[item.Level]
		} else {
			color = LogLevelColor[item.Level]
		}
		*show += fmt.Sprintf("%v [%v]  %v  %v %v @ %v:%v\n  => %v\u001b[0;0m\n",
			color,
			LogLevelDisp[item.Level],
			item.Mode,
			item.Mtime,
			item.Function,
			item.File,
			item.Line,
			msg,
		)
		count += 2
	}
	return count
}

func (d *Displayer) PrintStep3(clearance int,show *string) int {
	t,err := time.Parse(time.RFC3339,d.SelectedLL.Time)
	tshow := ""
	counterer := -1
	if err != nil {
		panic(err)
	}
	tshow += fmt.Sprintf("Level : %v[%v]\u001b[0;0m\nThreadID : %v  Timestamp : %v::%v\n",
		LogLevelColor[d.SelectedLL.Level],
		LogLevelDisp[d.SelectedLL.Level],
		d.SelectedLL.ThreadID,
		t.Format(time.DateTime),
		d.SelectedLL.Mtime,
	)
	funct := d.SelectedLL.Function
	if(len(funct) > 0) {
		funct += " @"
	}
	*show += fmt.Sprintf("From \u001b[1;36m%v\u001b[0;0m executing \u001b[1;36m%v %v : %v\u001b[0;0m\n",
		d.SelectedLL.Mode,
		funct,
		d.SelectedLL.File,
		d.SelectedLL.Line,
	)
	if len(d.SelectedLL.Stack) > 0 {
		tshow += fmt.Sprintf("\nStack :\n")
		for _,item := range d.SelectedLL.Stack {
			tshow += fmt.Sprint(" => ")
			if len(item.Type) > 0 {
				tshow += fmt.Sprintf("\u001b[1;36m%v\u001b[0;0m",item.Type)
			}
			if len(item.Class) > 0 {
				tshow += fmt.Sprintf("\u001b[1;36m%v\u001b[0;0m ::",item.Class)
			}
			if len(item.Function) > 0 {
				tshow += fmt.Sprintf("\u001b[1;36m%v(",item.Function)
				tshow += fmt.Sprintf(")\u001b[0;0m @")
			}
			tshow += fmt.Sprintf(" \u001b[1;36m%v\u001b[0;0m:\u001b[1;36m%v\u001b[0;0m\n",item.File,item.Line)
		}
	}
	tshow += fmt.Sprintf("\nmessage :\n")
	tshow += fmt.Sprintln(d.SelectedLL.Message)
	for _,line := range strings.Split(d.SelectedLL.Message, "\n") {
		counterer += len(line)/(d.W)
	}
	*show += tshow
	return strings.Count(tshow,"\n") + counterer
}