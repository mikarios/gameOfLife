package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

// const width = 1000
// const height = 1000

var (
	width  = 640
	height = 480
	white  = color.White
	black  = color.Black
	blue   = color.RGBA{69, 145, 196, 255}
	yellow = color.RGBA{255, 230, 120, 255}
)

type Game struct {
	generation uint64
	grid       [][]bool
	buffer     [][]bool
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.generation++

	if rand.Float32() < 0.01 {
		start := time.Now()
		defer func(t *time.Time) {
			fmt.Printf("updated in %v\n", time.Since(*t))
		}(&start)
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x <= width-4 && y <= height-4 && x >= 3 && y >= 3 {
			rangeX := []int{-3, -2, -1, 0, 1, 2, 3}
			rangeY := []int{-3, -2 - 1, 0, 1, 2, 3}

			switch x {
			case 0:
				rangeX = []int{0, 1}
			case width - 1:
				rangeX = []int{-1, 0}
			case width:
				rangeX = []int{-1}
			}

			switch y {
			case 0:
				rangeY = []int{0, 1}
			case height - 1:
				rangeY = []int{-1, 0}
			case height:
				rangeY = []int{-1}
			}

			for _, xx := range rangeX {
				for _, yy := range rangeY {
					if xx == 0 && yy == 0 {
						g.grid[x][y] = true
						continue
					}

					g.grid[x+xx][y+yy] = rand.Float32() < 0.8
				}
			}

		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(height)
	for y := 0; y < height; y++ {
		go g.updateX(wg, y)
	}
	wg.Wait()

	for x := range g.grid {
		copy(g.grid[x], g.buffer[x])
	}
	g.draw(screen)

	ebitenutil.DebugPrint(screen, fmt.Sprintf("Generation: %v", g.generation))
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if outsideWidth != width || outsideHeight != height {
		log.Println(outsideWidth, outsideHeight)
		switch {
		case len(g.grid) > outsideWidth:
			g.grid = g.grid[:outsideWidth]
			g.buffer = g.buffer[:outsideWidth]
		case len(g.grid) < outsideWidth:
			for i := len(g.grid); i < outsideWidth; i++ {
				g.grid = append(g.grid, make([]bool, outsideHeight))
				g.buffer = append(g.buffer, make([]bool, outsideHeight))
			}
		}

		switch {
		case len(g.grid[0]) > outsideHeight:
			for i := range g.grid {
				g.grid[i] = g.grid[i][:outsideHeight]
				g.buffer[i] = g.buffer[i][:outsideHeight]
			}
		case len(g.grid[0]) < outsideHeight:
			for i := range g.grid {
				if len(g.grid[i]) == outsideHeight {
					break
				}

				g.grid[i] = append(g.grid[i], make([]bool, outsideHeight-len(g.grid[i]))...)
				g.buffer[i] = append(g.buffer[i], make([]bool, outsideHeight-len(g.buffer[i]))...)
			}
		}

		width, height = outsideWidth, outsideHeight
	}

	return width, height
}

func (g *Game) draw(window *ebiten.Image) {
	window.Fill(black)

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if g.grid[x][y] {
				window.Set(x, y, yellow)
			}
		}
	}

}

func (g *Game) updateX(ywg *sync.WaitGroup, y int) {
	defer ywg.Done()
	for x := 0; x < width; x++ {
		g.updatePoint(x, y)
	}
}

func (g *Game) updatePoint(x, y int) {
	neighbours := 0
	rangeX := []int{-1, 0, 1}
	rangeY := []int{-1, 0, 1}

	switch x {
	case 0:
		rangeX = []int{0, 1}
	case width - 1:
		rangeX = []int{-1, 0}
	case width:
		rangeX = []int{-1}
	}

	switch y {
	case 0:
		rangeY = []int{0, 1}
	case height - 1:
		rangeY = []int{-1, 0}
	case height:
		rangeY = []int{-1}
	}

	for _, xx := range rangeX {
		for _, yy := range rangeY {
			if xx == 0 && yy == 0 {
				continue
			}
			if g.grid[x+xx][y+yy] {
				neighbours++
			}
		}
	}

	c := g.grid[x][y]
	g.buffer[x][y] = c

	switch {
	case !c && neighbours == 3:
		g.buffer[x][y] = true
	case c:
		if neighbours < 2 || neighbours > 3 {
			g.buffer[x][y] = false
		}
	}
}

func main() {
	g := &Game{}
	g.grid = make([][]bool, width)
	g.buffer = make([][]bool, width)
	ebiten.SetWindowSize(width, height)
	ebiten.SetWindowTitle("Your game's title")
	ebiten.SetWindowResizable(true)

	for x := 0; x < width; x++ {
		g.grid[x] = make([]bool, height)
		g.buffer[x] = make([]bool, height)
		for y := 0; y < height; y++ {
			g.grid[x][y] = rand.Float32() < 0.5
		}
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
