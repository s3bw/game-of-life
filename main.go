package main

import (
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/fatih/set"
	"golang.org/x/image/colornames"
)

const (
	size float64 = 16
	maxX float64 = 1024
	maxY float64 = 768
)

type delta struct {
	dx float64
	dy float64
}

var transformations = []*delta{
	&delta{dx: 0, dy: 1},
	&delta{dx: 0, dy: -1},
	&delta{dx: 1, dy: 0},
	&delta{dx: -1, dy: 0},
	&delta{dx: 1, dy: 1},
	&delta{dx: -1, dy: -1},
	&delta{dx: -1, dy: 1},
	&delta{dx: 1, dy: -1},
}

// Tile is a position on the map
type Tile struct {
	x     float64
	y     float64
	n     int
	alive bool
}

func createTile(x float64, y float64) *Tile {
	return &Tile{
		x:     x,
		y:     y,
		n:     0,
		alive: false,
	}
}

func (t *Tile) around() <-chan *Tile {
	ch := make(chan *Tile, 8)
	go func() {
		defer close(ch)
		for _, d := range transformations {
			nx, ny := t.x-(d.dx*size), t.y-(d.dy*size)
			if insideWindow(nx, ny) {
				ch <- createTile(nx, ny)
			}
		}
	}()
	return ch
}

func drawTile(imd *imdraw.IMDraw, x float64, y float64) {
	imd.Color = pixel.RGB(0.5, 0.5, 1)
	imd.Push(pixel.V(x, y), pixel.V(x+size, y+size))
	imd.Rectangle(0)
}

func round(i float64) float64 {
	return math.Floor(i/size) * size
}

func insideWindow(x float64, y float64) bool {
	return (((0.0 <= x) || (x < maxX)) || ((0.0 <= y) || (y < maxY)))
}

type key struct {
	x float64
	y float64
}

func step(cells set.Interface) {
	// Calculate crowding.
	// I'm tempted to make a data structure
	// for this tile set.
	tileMap := make(map[key]*Tile)
	for cells.Size() > 0 {
		e := cells.Pop()
		tile := e.(*Tile)

		k := key{x: tile.x, y: tile.y}
		if t, b := tileMap[k]; b {
			t.alive = true
		} else {
			tile.alive = true
			tileMap[k] = tile
		}

		for t := range tile.around() {
			k := key{x: t.x, y: t.y}
			if _tile, b := tileMap[k]; b {
				_tile.n++
			} else {
				t.n++
				tileMap[k] = t
			}

		}
	}

	// Calculate survivors
	for _, v := range tileMap {
		if (2 <= v.n) && (v.n < 4) && (v.alive) {
			v.n = 0
			cells.Add(v)
		} else if (v.n == 3) && (!v.alive) {
			v.n = 0
			cells.Add(v)
		}
	}

}

func run() {
	// 64x48 tiles of 16x16 pixels
	cfg := pixelgl.WindowConfig{
		Title:  "Gome-of-Life",
		Bounds: pixel.R(0, 0, maxX, maxY),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(nil)

	cells := set.New(set.NonThreadSafe)

	for !win.Closed() {
		win.UpdateInputWait(-1)

		if win.JustPressed(pixelgl.MouseButtonLeft) {
			pos := win.MousePosition()
			x := round(pos.X)
			y := round(pos.Y)
			tile := createTile(x, y)
			tile.alive = true

			// Avoid adding tiles on top
			// of each other.
			// This is currently broken, I'll
			// need to write a Tile specific set.
			if !cells.Has(tile) {
				cells.Add(tile)
			}
		}
		// Single game step
		if win.JustPressed(pixelgl.KeyEnter) {
			imd.Clear()
			step(cells)
		}

		// Draw calls
		for _, e := range cells.List() {
			tile := e.(*Tile)
			drawTile(imd, tile.x, tile.y)
		}
		win.Clear(colornames.Aliceblue)
		imd.Draw(win)
		win.Update()
	}
}

func main() {
	pixelgl.Run(run)
}
