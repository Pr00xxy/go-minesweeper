package main

import (
	"log"
	"math/rand"
	"strconv"

	"github.com/famz/SetLocale"
	gc "github.com/rthornton128/goncurses"
)

type GameSettings struct {
	height int
	width  int
	mines  int
}

type Cell struct {
	isBomb    bool
	isFlagged bool
	isOpen    bool
	adjacency int
}

type Board struct {
	window   *gc.Window
	settings GameSettings
	grid     [][]Cell
}

var straightDirections = []struct{ dr, dc int }{
	{-1, 0},
	{0, -1} /*   current   */, {0, 1},
	{1, 0},
}

var diagonalDirections = []struct{ dr, dc int }{
	{-1, -1}, {-1, 1},
	{1, -1}, {1, 1},
}

func generateGrid(settings *GameSettings) [][]Cell {

	grid := make([][]Cell, settings.height)
	for i := range grid {
		grid[i] = make([]Cell, settings.width)
	}

	placed := 0
	for placed < settings.mines { // Change <= to <
		// generate a random coordinate
		r, c := rand.Intn(settings.height), rand.Intn(settings.width)
		if grid[r][c].isBomb {
			continue
		}

		grid[r][c].isBomb = true
		placed++
	}

	// Calculate adjacency counts
	for row := range grid {
		for col := range grid[row] {
			if grid[row][col].isBomb {
				for _, dir := range append(straightDirections, diagonalDirections...) {
					adjRow, adjCol := row+dir.dr, col+dir.dc
					if adjRow >= 0 && adjRow < settings.height && adjCol >= 0 && adjCol < settings.width {
						grid[adjRow][adjCol].adjacency++
					}
				}
			}
		}
	}

	return grid
}

func NewBoard(settings *GameSettings, window *gc.Window, gameSettings GameSettings) Board {
	grid := generateGrid(settings)
	return Board{
		grid:     grid,
		window:   window,
		settings: gameSettings,
	}
}

func (b *Board) render() {
	for _, row := range b.grid {
		rowString := ""
		for _, cell := range row {
			if cell.isFlagged {
				rowString = rowString + "F"
				continue
			}
			if !cell.isOpen {
				rowString = rowString + "."
				continue
			}
			if cell.isOpen && cell.isBomb == false {
				rowString = rowString + strconv.Itoa(cell.adjacency)
				continue
			}
			if cell.isBomb {
				rowString = rowString + "X"
			}
		}
		b.window.Print(rowString)
	}
}

func (b *Board) flagCell(x int, y int) {
	b.grid[x][y].isFlagged = true
}

func (b *Board) openCell(x int, y int) {
	currentCell := b.grid[x][y]

	if currentCell.isFlagged {
		return
	}

	currentCell.isOpen = true

	if currentCell.adjacency > 0 || currentCell.isBomb {
		return
	}

	for _, dir := range straightDirections {
		adjRow, adjCol := x+dir.dr, y+dir.dc
		if adjRow >= 0 && adjRow < b.settings.height && adjCol >= 0 && adjCol < b.settings.width {
			adjacentCell := b.grid[adjRow][adjCol]
			if !adjacentCell.isOpen && !adjacentCell.isBomb {
				b.openCell(adjRow, adjCol)
			}
		}
	}
}

func main() {
	height, width := 20, 30
	difficulty := 5
	SetLocale.SetLocale(SetLocale.LC_ALL, "")
	settings := GameSettings{
		height: height,
		width:  width,
		mines:  (height * width) / (15 - difficulty),
	}

	stdscr, err := gc.Init()
	stdscr.Clear()
	gc.Cursor(1)
	gc.Echo(true)

	if err != nil {
		log.Fatal("init:", err)
	}

	defer gc.End()
	stdscr.Clear()
	stdscr.Keypad(true)

	rows, cols := stdscr.MaxYX()
	maxY, maxX := (rows-height)/2, (cols-width)/2
	boardWindow, err := gc.NewWindow(settings.height, settings.width, maxY, maxX)
	boardWindow.Keypad(true)

	if err != nil {
		log.Fatal("failed to create window:", err)
	}

	in := make(chan gc.Char)
	ready := make(chan bool)
	go func(w *gc.Window, ch chan<- gc.Char) {
		for {
			<-ready
			ch <- gc.Char(w.GetChar())
		}
	}(boardWindow, in)

	board := NewBoard(&settings, boardWindow, settings)

	x, y := 0, 0
	board.render()
	boardWindow.Move(y, x)

	gc.Update()

	for {
		var c gc.Char
		select {
		case c = <-in:
			switch gc.Key(c) {
			case gc.KEY_UP:
				if y > 0 {
					y--
				}
			case gc.KEY_DOWN:
				if y < settings.height-1 {
					y++
				}
			case gc.KEY_LEFT:
				if x > 0 {
					x--
				}
			case gc.KEY_RIGHT:
				if x < settings.width-1 {
					x++
				}
			case gc.KEY_RETURN:
				board.openCell(y, x)

			case 32:
				board.flagCell(y, x)
			}

			boardWindow.Erase()
			board.render()

			boardWindow.Move(y, x)
			boardWindow.NoutRefresh()
			gc.Update()

		case ready <- true:
		}
		if c == gc.Char('q') {
			break
		}
	}
}
