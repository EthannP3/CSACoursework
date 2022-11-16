package gol

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

func worker(world func(x, y int) uint8, sY, eY, sX, eX int, shared [][]uint8, events chan<- Event, p Params, wg *sync.WaitGroup, turn int) {
	gol(world, shared, p.ImageHeight, p.ImageWidth, sY, eY, sX, eX, turn, events)
	wg.Done()
}

func CountAlive(end chan bool) {
	for {
		select {
		case <-end:
			fmt.Println("Ended, stopped counting turns.")
			return
		default:
			time.Sleep(2 * time.Second)

			fmt.Println("Completed Turns: ", 1)
		}
	}
}

func gol(world func(x, y int) uint8, sharedWorld [][]uint8, height, width int, sY, eY, sX, eX int, turn int, events chan<- Event) {

	for y := sY; y < eY; y++ {
		for x := sX; x < eX; x++ {
			sum := world(int(math.Mod(float64(x+width-1), float64(width))), int(math.Mod(float64(y+height-1), float64(height))))/255 +
				world(int(math.Mod(float64(x+width), float64(width))), int(math.Mod(float64(y+height-1), float64(height))))/255 +
				world(int(math.Mod(float64(x+width+1), float64(width))), int(math.Mod(float64(y+height-1), float64(height))))/255 +
				world(int(math.Mod(float64(x+width-1), float64(width))), int(math.Mod(float64(y+height), float64(height))))/255 +
				world(int(math.Mod(float64(x+width+1), float64(width))), int(math.Mod(float64(y+height), float64(height))))/255 +
				world(int(math.Mod(float64(x+width-1), float64(width))), int(math.Mod(float64(y+height+1), float64(height))))/255 +
				world(int(math.Mod(float64(x+width), float64(width))), int(math.Mod(float64(y+height+1), float64(height))))/255 +
				world(int(math.Mod(float64(x+width+1), float64(width))), int(math.Mod(float64(y+height+1), float64(height))))/255

			if world(x, y) == 255 { // this cell is alive
				if sum == 2 || sum == 3 {
					sharedWorld[y][x] = 255
				} else {
					sharedWorld[y][x] = 0
					events <- CellFlipped{
						CompletedTurns: turn,
						Cell: util.Cell{
							X: x,
							Y: y,
						},
					}
					//fmt.Println("new world ", x, y, " flipped to dead. Turn:", turn)
				}

			} else { // this cell is dead
				if sum == 3 {
					sharedWorld[y][x] = 255
					events <- CellFlipped{
						CompletedTurns: turn,
						Cell: util.Cell{
							X: x,
							Y: y,
						},
					}
					//fmt.Println("new world ", x, y, " flipped to alive. Turn:", turn)
				} else {
					sharedWorld[y][x] = 0
				}

			}
		}
	}

	events <- TurnComplete{CompletedTurns: turn + 1}
}

func makeImmutableWorld(w [][]uint8) func(x, y int) uint8 {
	l := len(w)

	iW := make([][]uint8, l)
	for i := 0; i < l; i++ {
		iW[i] = make([]uint8, l)
	}

	for y := 0; y < l; y++ {
		for x := 0; x < l; x++ {
			iW[y][x] = w[y][x]
		}
	}

	return func(x, y int) uint8 {
		return iW[y][x]
	}
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	c.ioCommand <- ioInput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)

	world := make([][]uint8, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]uint8, p.ImageWidth)
		for j := 0; j < p.ImageWidth; j++ {
			world[i][j] = <-c.ioInput
		}
	}

	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	// TODO: Start workers and piece together GoL

	immutableWorld := makeImmutableWorld(world)

	sharedWorld := make([][]uint8, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		sharedWorld[i] = make([]uint8, p.ImageWidth)
	}
	exit := make(chan bool)
	wg := &sync.WaitGroup{}
	//remainder := int(math.Mod(float64(p.ImageHeight), float64(p.Threads)))
	go CountAlive(exit)
	for turn := 0; turn < p.Turns; turn++ {
		wg.Add(p.Threads)

		for w := 0; w < p.Threads-1; w++ {
			go worker(immutableWorld, w*(p.ImageHeight/p.Threads), (w+1)*(p.ImageHeight/p.Threads), 0, p.ImageWidth, sharedWorld, c.events, p, wg, turn)
		}
		go worker(immutableWorld, (p.Threads-1)*(p.ImageHeight/p.Threads), p.ImageHeight, 0, p.ImageWidth, sharedWorld, c.events, p, wg, turn)

		// block here until done
		wg.Wait()
		immutableWorld = makeImmutableWorld(sharedWorld)
	}

	if p.Turns == 0 {
		sharedWorld = world
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	var alive []util.Cell
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if sharedWorld[y][x] == 255 {
				alive = append(alive, util.Cell{X: x, Y: y})
			}
		}
	}
	exit <- true
	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          alive,
	}

	c.ioCommand <- ioOutput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- sharedWorld[y][x]
		}
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
