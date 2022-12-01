package gol

import (
	"github.com/ChrisGora/semaphore"
	"sync"
)

// Params provides the details of how to run the Game of Life and which image to load.
type Params struct {
	Turns       int
	Threads     int
	ImageWidth  int
	ImageHeight int
}

type SharedData[T any] struct {
	Value      *[]T
	Mutex      *sync.Mutex
	ContentSem semaphore.Semaphore
	SpaceSem   semaphore.Semaphore
}

// Run starts the processing of Game of Life. It should initialise channels and goroutines.
func Run(p Params, sharedEvents SharedData[Event], keyPresses <-chan rune) {

	//	TODO: Put the missing channels in here.
	ioCommand := SharedData[ioCommand]{
		Value:      &[]ioCommand{},
		Mutex:      &sync.Mutex{},
		ContentSem: semaphore.Init(10, 0),
		SpaceSem:   semaphore.Init(10, 10),
	}
	ioIdle := SharedData[bool]{
		Value:      &[]bool{},
		Mutex:      &sync.Mutex{},
		ContentSem: semaphore.Init(10, 0),
		SpaceSem:   semaphore.Init(10, 10),
	}
	ioFilename := SharedData[string]{
		Value:      &[]string{},
		Mutex:      &sync.Mutex{},
		ContentSem: semaphore.Init(10, 0),
		SpaceSem:   semaphore.Init(10, 10),
	}
	ioOutput := SharedData[uint8]{
		Value:      &[]uint8{},
		Mutex:      &sync.Mutex{},
		ContentSem: semaphore.Init(100, 0),
		SpaceSem:   semaphore.Init(100, 100),
	}
	ioInput := SharedData[uint8]{
		Value:      &[]uint8{},
		Mutex:      &sync.Mutex{},
		ContentSem: semaphore.Init(100, 0),
		SpaceSem:   semaphore.Init(100, 100),
	}

	ioChannels := ioChannels{
		command:  ioCommand,
		idle:     ioIdle,
		filename: ioFilename,
		output:   ioOutput,
		input:    ioInput,
	}
	go startIo(p, ioChannels)

	distributorChannels := distributorChannels{
		events:     sharedEvents,
		ioCommand:  ioCommand,
		ioIdle:     ioIdle,
		ioFilename: ioFilename,
		ioOutput:   ioOutput,
		ioInput:    ioInput,
	}
	distributor(p, distributorChannels, keyPresses)
}
