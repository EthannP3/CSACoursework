package gol

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     SharedData[Event]
	ioCommand  SharedData[ioCommand]
	ioIdle     SharedData[bool]
	ioFilename SharedData[string]
	ioOutput   SharedData[uint8]
	ioInput    SharedData[uint8]
}

func worker(world func(x, y int) uint8, sY, eY, sX, eX int, shared [][]uint8, c distributorChannels, p Params, wg *sync.WaitGroup, turn int) {
	gol(world, shared, p.ImageHeight, p.ImageWidth, sY, eY, sX, eX, turn, c)
	wg.Done()
}

func putData[T any](shared SharedData[T], d T) {
	shared.SpaceSem.Wait()
	shared.Mutex.Lock()
	*shared.Value = append(*shared.Value, d)
	shared.Mutex.Unlock()
	shared.ContentSem.Post()
}
func getData[T any](shared SharedData[T]) T {
	shared.ContentSem.Wait()
	shared.Mutex.Lock()
	var d T
	d, *shared.Value = (*shared.Value)[0], (*shared.Value)[1:]
	shared.Mutex.Unlock()
	shared.SpaceSem.Post()
	return d
}

func handleEvent(events SharedData[Event], e Event) {
	putData(events, e)
}

func getAlive(world func(x, y int) uint8, dim int) int {
	var total = 0
	for y := 0; y < dim; y++ {
		for x := 0; x < dim; x++ {
			if world(x, y) == 255 {
				total += 1
			}
		}
	}
	return total

}

func gol(world func(x, y int) uint8, sharedWorld [][]uint8, height, width int, sY, eY, sX, eX int, turn int, c distributorChannels) {

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
					handleEvent(c.events, CellFlipped{
						CompletedTurns: turn,
						Cell: util.Cell{
							X: x,
							Y: y,
						},
					})
				}

			} else { // this cell is dead
				if sum == 3 {
					sharedWorld[y][x] = 255
					handleEvent(c.events, CellFlipped{
						CompletedTurns: turn,
						Cell: util.Cell{
							X: x,
							Y: y,
						},
					})
				} else {
					sharedWorld[y][x] = 0
				}

			}
		}
	}
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
func distributor(p Params, c distributorChannels, keyPresses <-chan rune) {
	// TODO: Create a 2D slice to store the world.
	putData(c.ioCommand, ioInput)
	putData(c.ioFilename, strconv.Itoa(p.ImageWidth)+"x"+strconv.Itoa(p.ImageHeight))

	exit := make(chan bool)
	ticker := time.NewTicker(2 * time.Second)

	pause := &sync.WaitGroup{}

	world := make([][]uint8, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		world[i] = make([]uint8, p.ImageWidth)
		for j := 0; j < p.ImageWidth; j++ {
			world[i][j] = getData(c.ioInput)
			if world[i][j] == 255 {
				handleEvent(c.events, CellFlipped{
					CompletedTurns: 0,
					Cell:           util.Cell{X: j, Y: i},
				})
			}
		}
	}

	putData(c.ioCommand, ioCheckIdle)
	getData(c.ioIdle)

	// TODO: Start workers and piece together GoL

	immutableWorld := makeImmutableWorld(world)

	var turn = 0
	completedTurns := 0

	go func() {
		paused := false
		for {
			select {
			case key := <-keyPresses:
				switch key {
				case 'p':
					if paused {
						pause.Done()
						pause.Done()
						ticker.Reset(2 * time.Second)

						paused = false
						fmt.Println("Continuing")

						handleEvent(c.events, StateChange{CompletedTurns: completedTurns, NewState: Executing})
					} else {
						pause.Add(2)
						ticker.Stop()

						paused = true
						handleEvent(c.events, StateChange{CompletedTurns: completedTurns, NewState: Paused})
					}
				case 's', 'q':
					// output PGM
					putData(c.ioCommand, ioOutput)
					filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(completedTurns)
					putData(c.ioFilename, filename)

					for y := 0; y < p.ImageHeight; y++ {
						for x := 0; x < p.ImageWidth; x++ {
							putData(c.ioOutput, immutableWorld(x, y))
						}
					}

					handleEvent(c.events, ImageOutputComplete{CompletedTurns: completedTurns, Filename: filename})

					if key == 'q' {
						putData(c.ioCommand, ioCheckIdle)
						getData(c.ioIdle)

						handleEvent(c.events, StateChange{CompletedTurns: completedTurns, NewState: Quitting})

						os.Exit(0)
					}
				}
			}
		}
	}()

	sharedWorld := make([][]uint8, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		sharedWorld[i] = make([]uint8, p.ImageWidth)
	}

	wg := &sync.WaitGroup{}

	go func() {
		for {

			select {
			case <-exit:
				return
			case <-ticker.C:
				pause.Wait()

				count := getAlive(immutableWorld, p.ImageWidth)
				turns := completedTurns + 1

				handleEvent(c.events, AliveCellsCount{
					CompletedTurns: turns,
					CellsCount:     count,
				})
			}
		}
	}()

	for turn = 0; turn < p.Turns; turn++ {
		wg.Add(p.Threads)

		for w := 0; w < p.Threads-1; w++ {
			go worker(immutableWorld, w*(p.ImageHeight/p.Threads), (w+1)*(p.ImageHeight/p.Threads), 0, p.ImageWidth, sharedWorld, c, p, wg, turn)
		}
		go worker(immutableWorld, (p.Threads-1)*(p.ImageHeight/p.Threads), p.ImageHeight, 0, p.ImageWidth, sharedWorld, c, p, wg, turn)

		// block here until done
		wg.Wait()

		pause.Wait()

		immutableWorld = makeImmutableWorld(sharedWorld)
		completedTurns = turn
		handleEvent(c.events, TurnComplete{CompletedTurns: turn + 1})

	}
	fmt.Println("outside main loop")

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

	handleEvent(c.events, FinalTurnComplete{
		CompletedTurns: p.Turns,
		Alive:          alive,
	})

	ticker.Stop()
	exit <- true

	putData(c.ioCommand, ioOutput)
	filename := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.Turns)
	putData(c.ioFilename, filename)

	for y := 0; y < p.ImageHeight; y++ {
		for x := 0; x < p.ImageWidth; x++ {
			putData(c.ioOutput, sharedWorld[y][x])
		}
	}

	// Make sure that the Io has finished any output before exiting.
	putData(c.ioCommand, ioCheckIdle)
	getData(c.ioIdle)

	handleEvent(c.events, ImageOutputComplete{CompletedTurns: p.Turns, Filename: filename})

	handleEvent(c.events, StateChange{p.Turns, Quitting})
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	//close(c.events)
	fmt.Println("closin")
}
