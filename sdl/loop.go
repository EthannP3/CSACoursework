package sdl

import (
	"fmt"
	_ "github.com/ChrisGora/semaphore"
	"github.com/veandco/go-sdl2/sdl"
	_ "sync"
	"uk.ac.bris.cs/gameoflife/gol"
)

func getData[T any](shared gol.SharedData[T]) T {
	shared.ContentSem.Wait()
	shared.Mutex.Lock()
	var d T
	d, *shared.Value = (*shared.Value)[0], (*shared.Value)[1:]
	shared.Mutex.Unlock()
	shared.SpaceSem.Post()
	return d
}

func getEvent(h gol.SharedData[gol.Event]) gol.Event {
	event := getData(h)
	return event
}

func Run(p gol.Params, events gol.SharedData[gol.Event], keyPresses chan<- rune) {
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
		if events.ContentSem.GetValue() >= 0 {
			gEvent := getEvent(events)
			if gEvent == nil {
				fmt.Printf("event is nil\n")
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
