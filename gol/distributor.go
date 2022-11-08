package gol

import (
<<<<<<< HEAD
	"strconv"
=======
	"math"
>>>>>>> a1008acbc72f2fe76f574642ad1506c862fdf437
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

	world := make([][]int, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]int, p.ImageWidth)
	}

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

<<<<<<< HEAD
=======
					if sum == 2 || sum == 3 {
						newWorld[y][x] = 1
					} else {
						newWorld[y][x] = 0
					}

				} else { // this cell is dead

					if sum == 3 {
						newWorld[y][x] = 1
					} else {
						newWorld[y][x] = 0
					}

				}
			}
		}

	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

>>>>>>> a1008acbc72f2fe76f574642ad1506c862fdf437
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{turn, Quitting}
<<<<<<< HEAD

=======
>>>>>>> a1008acbc72f2fe76f574642ad1506c862fdf437
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
