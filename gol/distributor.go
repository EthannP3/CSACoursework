package gol

import (
	"fmt"
	"math"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	fmt.Println("started distributor")

	c.ioCommand <- ioInput

	world := make([][]uint8, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]uint8, p.ImageWidth)
		for j := 0; j < p.ImageWidth; j++ {
			world[i][j] = <-c.ioInput
		}
	}

	fmt.Println("loaded world")

	turn := 0

	// TODO: Execute all turns of the Game of Life.
	newWorld := make([][]int, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		newWorld[i] = make([]int, p.ImageWidth)
	}

	for ; turn < p.Turns; turn++ {

		for y := 0; y < p.ImageHeight; y++ {
			for x := 0; x < p.ImageWidth; x++ {
				sum := world[int(math.Mod(float64(y+p.ImageHeight-1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth-1), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight-1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight-1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth+1), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth-1), float64(p.ImageWidth)))] +
					//
					world[int(math.Mod(float64(y+p.ImageHeight), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth+1), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight+1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth-1), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight+1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth), float64(p.ImageWidth)))] +
					world[int(math.Mod(float64(y+p.ImageHeight+1), float64(p.ImageHeight)))][int(math.Mod(float64(x+p.ImageWidth+1), float64(p.ImageWidth)))]

				if world[y][x] == 255 { // this cell is alive
					if sum == 2 || sum == 3 {
						newWorld[y][x] = 255
					} else {
						newWorld[y][x] = 0
						c.events <- CellFlipped{}
					}

				} else { // this cell is dead

					if sum == 3 {
						newWorld[y][x] = 255
						c.events <- CellFlipped{}
					} else {
						newWorld[y][x] = 0
					}

				}
			}
		}

		c.events <- TurnComplete{}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	fmt.Println("finished game")
	c.events <- FinalTurnComplete{}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
