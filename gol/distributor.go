package gol

import (
	"math"
	"strconv"
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

func worker(world func(x, y int) uint8, sY, eY, sX, eX int, out chan<- [][]uint8, events chan<- Event, p Params) {
	var newWorld [][]uint8
	for turn := 0; turn < p.Turns; turn++ {
		newWorld = gol(world, p.ImageHeight, p.ImageWidth, sY, eY, sX, eX, turn, events)
		world = makeImmutableWorld(newWorld)
	}
	out <- newWorld
}

func gol(world func(x, y int) uint8, height, width int, sY, eY, sX, eX int, turn int, events chan<- Event) [][]uint8 {

	h := eY - sY
	w := eX - sX

	newWorld := make([][]uint8, h)
	for i := 0; i < h; i++ {
		newWorld[i] = make([]uint8, w)
	}

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
					newWorld[y][x] = 255
				} else {
					newWorld[y][x] = 0
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
					newWorld[y][x] = 255
					events <- CellFlipped{
						CompletedTurns: turn,
						Cell: util.Cell{
							X: x,
							Y: y,
						},
					}
					//fmt.Println("new world ", x, y, " flipped to alive. Turn:", turn)
				} else {
					newWorld[y][x] = 0
				}

			}
		}
	}

	events <- TurnComplete{CompletedTurns: turn + 1}

	return newWorld
}

func makeImmutableWorld(w [][]uint8) func(x, y int) uint8 {
	return func(x, y int) uint8 {
		return w[y][x]
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

	out := make([]chan [][]uint8, p.Threads)
	for i := range out {
		out[i] = make(chan [][]uint8)
	}

	for w := 0; w < p.Threads; w++ {
		go worker(immutableWorld, w*(p.ImageHeight/p.Threads), (w+1)*(p.ImageHeight/p.Threads), 0, p.ImageWidth, out[w], c.events, p)
	}

	newWorld := make([][]uint8, p.ImageHeight)
	for x := 0; x < p.ImageHeight; x++ {
		newWorld[x] = make([]uint8, p.ImageWidth)
	}

	for w := 0; w < p.Threads; w++ {

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	var alive []util.Cell
	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			if newWorld[y][x] == 255 {
				alive = append(alive, util.Cell{X: x, Y: y})
			}
		}
	}

	c.events <- FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          alive,
	}

	c.ioCommand <- ioOutput
	c.ioFilename <- strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			c.ioOutput <- newWorld[y][x]
		}
	}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
