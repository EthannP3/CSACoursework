package sdl

import (
	"fmt"
	"github.com/ChrisGora/semaphore"
	"github.com/veandco/go-sdl2/sdl"
	"sync"
	"uk.ac.bris.cs/gameoflife/gol"
)

func getEvent(h gol.SharedData[gol.Event]) gol.Event {
	event := gol.getData(h)
	return event
}

func Run(p gol.Params, events []gol.Event, eventSemaphore semaphore.Semaphore, eventMutex *sync.Mutex, keyPresses chan<- rune) {
	h := EventHandler{
		events:    events,
		mutex:     eventMutex,
		semaphore: eventSemaphore,
	}
	w := NewWindow(int32(p.ImageWidth), int32(p.ImageHeight))

sdlLoop:
	for {
		event := w.PollEvent()
		if event != nil {
			switch e := event.(type) {
			case *sdl.KeyboardEvent:
				switch e.Keysym.Sym {
				case sdl.K_p:
					keyPresses <- 'p'
				case sdl.K_s:
					keyPresses <- 's'
				case sdl.K_q:
					keyPresses <- 'q'
				case sdl.K_k:
					keyPresses <- 'k'
				}
			}
		}
		if h.semaphore.GetValue() >= 0 {
			gEvent, ok := getEvent(h)
			if gEvent == nil {
				fmt.Printf("event is nil\n")
			}

			if !ok {
				w.Destroy()
				break sdlLoop
			}
			switch e := gEvent.(type) {
			case gol.CellFlipped:
				w.FlipPixel(e.Cell.X, e.Cell.Y)
			case gol.TurnComplete:
				w.RenderFrame()
			case gol.FinalTurnComplete:
				w.Destroy()
				break sdlLoop
			default:
				if len(gEvent.String()) > 0 {
					fmt.Printf("Completed Turns %-8v%v\n", gEvent.GetCompletedTurns(), event)
				}
			}
		}

	}

}
