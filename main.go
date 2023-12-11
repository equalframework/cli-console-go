package main

import (
	"fmt"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"golang.org/x/term"

	"github.com/equalframework/cli-console-go/display"
	"github.com/equalframework/cli-console-go/syscalls"
)

func main() {

	fd := int(os.Stdin.Fd())

	fmt.Println(fd)

	if !term.IsTerminal(fd) {
		fmt.Println("Please launch this cli in a terminal.")
	}
	
	syscalls.Init()

	//fmt.Println(width,height)
	if(len(os.Args) < 2) {
		fmt.Fprintf(os.Stderr,"You need to give the equal log file in cli parameter.")
		os.Exit(1)
	}
	disper := display.NewDisplayer(os.Args[1])
	disper.ReadContent()
	keysEvents,err := keyboard.GetKeys(10)
	if err != nil {
		panic(err)
	}
	defer keyboard.Close()
	var event_fired,has_been_reloaded bool
	disper.PrintResult()
	for true {
		event_fired = false
		has_been_reloaded = false
		for len(keysEvents) > 0 {
			event := <-keysEvents
			if event.Err != nil {
				panic(nil)
			}
			//fmt.Println(event)
			if(event.Key == keyboard.KeyArrowUp) {
				if(disper.SelectedLine[disper.ContextLevel] < disper.Lenght()-1) {
					disper.SelectedLine[disper.ContextLevel] ++
					event_fired = true
					continue
				}
			}
			if(event.Key == keyboard.KeyArrowDown) {
				if(disper.SelectedLine[disper.ContextLevel] > 0) {
					disper.SelectedLine[disper.ContextLevel] --
					event_fired = true
					continue
				}
			}
			if(event.Key == keyboard.KeyEsc || event.Key == 3 || event.Rune == 'q') {
				syscalls.CallClear()
				os.Exit(0)
			}
			if(event.Key == keyboard.KeyEnter) {
				if disper.ContextLevel < 2 {
					fmt.Println(disper.ContextLevel)
					disper.SelectCurrent()
					fmt.Println(disper.ContextLevel)
					disper.CurrentSkip[disper.ContextLevel] = 0
					disper.SelectedLine[disper.ContextLevel] = 0
					event_fired = true
					continue
				}
			}
			if(event.Key == keyboard.KeyBackspace2) {
				if disper.ContextLevel > 0 {
					disper.ContextLevel --
					event_fired = true
					continue
				}
			}
			if(event.Key == 65519) { //TOP
				disper.SelectedLine[disper.ContextLevel] = disper.Lenght() - 1
				event_fired = true
				continue
			} 
			if(event.Key == 65518) { //BOTTOM
				disper.SelectedLine[disper.ContextLevel] = 0
				event_fired = true
				continue
			} 
			if(event.Key ==keyboard.KeyTab) {
				disper.FilterTypeIndex = (disper.FilterTypeIndex + 1)%len(display.Filter_type)
				disper.CurrentSkip[disper.ContextLevel] = 0
				event_fired = true
				continue
			}
			if(event.Key == 18) {
				if(disper.ContextLevel != 0) {
					continue
				}
				disper.ReadContent()
				disper.ContextLevel = 0
				disper.SelectedLine[0] = 0
				disper.SelectedLine[1] = 0
				disper.SelectedLine[2] = 0
				event_fired = true
				has_been_reloaded = true
				continue
			}
			
		}
		
		if(event_fired || disper.RefreshSize()) {
			disper.PrintResult()
			if has_been_reloaded {
				fmt.Print("RELOADED !")
			}
			fmt.Println()
		}
		time.Sleep(30*time.Millisecond)
	}
	
}