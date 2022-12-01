package main

import (
	"flag"
	"fmt"
	"github.com/ChrisGora/semaphore"
	"runtime"
	"sync"

	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/sdl"
)

// main is the function called when starting Game of Life with 'go run .'
func main() {
	runtime.LockOSThread()
	var params gol.Params

	flag.IntVar(
		&params.Threads,
		"t",
		8,
		"Specify the number of worker threads to use. Defaults to 8.")

	flag.IntVar(
		&params.ImageWidth,
		"w",
		512,
		"Specify the width of the image. Defaults to 512.")

	flag.IntVar(
		&params.ImageHeight,
		"h",
		512,
		"Specify the height of the image. Defaults to 512.")

	flag.IntVar(
		&params.Turns,
		"turns",
		10000000000,
		"Specify the number of turns to process. Defaults to 10000000000.")

	noVis := flag.Bool(
		"noVis",
		false,
		"Disables the SDL window, so there is no visualisation during the tests.")

	flag.Parse()

	fmt.Println("Threads:", params.Threads)
	fmt.Println("Width:", params.ImageWidth)
	fmt.Println("Height:", params.ImageHeight)

	keyPresses := make(chan rune, 10)
	var events []gol.Event
	eventAvailable := semaphore.Init(1000, 0)
	spaceAvailable := semaphore.Init(1000, 1000)
	eventMutex := new(sync.Mutex)

	sharedEvents := gol.SharedData[gol.Event]{
		Value:      &events,
		Mutex:      eventMutex,
		ContentSem: eventAvailable,
		SpaceSem:   spaceAvailable,
	}

	go gol.Run(params, sharedEvents, keyPresses)
	if !(*noVis) {
		sdl.Run(params, events, eventAvailable, eventMutex, keyPresses)
	} else {
		complete := false
		for !complete {
			eventAvailable.Wait()
			eventMutex.Lock()
			event := events[0]
			if len(events) > 1 {
				events = events[1:]
			} else {
				events[0] = nil
			}
			eventMutex.Unlock()
			spaceAvailable.Post()

			switch event.(type) {
			case gol.FinalTurnComplete:
				complete = true
			}
		}
	}
}
