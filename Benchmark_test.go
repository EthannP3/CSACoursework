package main

import (
	"fmt"
	"github.com/ChrisGora/semaphore"
	"os"
	"sync"
	"testing"
	"uk.ac.bris.cs/gameoflife/gol"
)

func putData[T any](shared gol.SharedData[T], d T) {
	shared.SpaceSem.Wait()
	shared.Mutex.Lock()
	*shared.Value = append(*shared.Value, d)
	shared.Mutex.Unlock()
	shared.ContentSem.Post()
}
func getData[T any](shared gol.SharedData[T]) T {
	shared.ContentSem.Wait()
	shared.Mutex.Lock()
	var d T
	d, *shared.Value = (*shared.Value)[0], (*shared.Value)[1:]
	shared.Mutex.Unlock()
	shared.SpaceSem.Post()
	return d
}

const benchLength = 200

func BenchmarkGol(b *testing.B) {

	for threads := 1; threads <= 16; {
		os.Stdout = nil // Disable all program output apart from benchmark results
		p := gol.Params{
			Turns:       benchLength,
			Threads:     threads,
			ImageWidth:  512,
			ImageHeight: 512,
		}
		name := fmt.Sprintf("%dx%dx%d-%d", p.ImageWidth, p.ImageHeight, p.Turns, p.Threads)
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				events := gol.SharedData[gol.Event]{
					Value:      &[]gol.Event{},
					Mutex:      &sync.Mutex{},
					ContentSem: semaphore.Init(1000, 0),
					SpaceSem:   semaphore.Init(1000, 1000),
				}
				go gol.Run(p, events, nil)
				complete := false
				for !complete {
					event := getData(events)
					switch event.(type) {
					case gol.FinalTurnComplete:
						complete = true
					}
				}
			}
		})
		//if threads == 16 {
		//	threads = threads * 2
		//}
		threads = threads + 1
	}
}
