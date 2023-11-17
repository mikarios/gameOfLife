package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

var yellow = color.RGBA{255, 230, 120, 255}

type Game struct {
	generation uint64
	grid       [][]bool
	buffer     [][]bool
	width      int
	height     int
}

func (g *Game) Update(screen *ebiten.Image) error {
	g.generation++

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x <= g.width-4 && y <= g.height-4 && x >= 3 && y >= 3 {
			rangeX := []int{-3, -2, -1, 0, 1, 2, 3}
			rangeY := []int{-3, -2 - 1, 0, 1, 2, 3}

			switch x {
			case 0:
				rangeX = []int{0, 1}
			case g.width - 1:
				rangeX = []int{-1, 0}
			case g.width:
				rangeX = []int{-1}
			}

			switch y {
			case 0:
				rangeY = []int{0, 1}
			case g.height - 1:
				rangeY = []int{-1, 0}
			case g.height:
				rangeY = []int{-1}
			}

			for _, xx := range rangeX {
				for _, yy := range rangeY {
					g.grid[x+xx][y+yy] = rand.Float32() < 0.8
				}
			}
		}
	}

	wg := &sync.WaitGroup{}
	wg.Add(g.height)
	for y := 0; y < g.height; y++ {
		go g.updateX(wg, y)
	}
	wg.Wait()

	for x := range g.grid {
		copy(g.grid[x], g.buffer[x])
	}
	g.draw(screen)

	_ = ebitenutil.DebugPrint(screen, fmt.Sprintf("Generation: %v", g.generation))
	return nil
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	if outsideWidth != g.width || outsideHeight != g.height {
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

		g.width, g.height = outsideWidth, outsideHeight
	}

	return g.width, g.height
}

func (g *Game) draw(window *ebiten.Image) {
	_ = window.Fill(color.Black)

	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			if g.grid[x][y] {
				window.Set(x, y, yellow)
			}
		}
	}

}

func (g *Game) updateX(ywg *sync.WaitGroup, y int) {
	defer ywg.Done()
	for x := 0; x < g.width; x++ {
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
	case g.width - 1:
		rangeX = []int{-1, 0}
	case g.width:
		rangeX = []int{-1}
	}

	switch y {
	case 0:
		rangeY = []int{0, 1}
	case g.height - 1:
		rangeY = []int{-1, 0}
	case g.height:
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
	g := &Game{width: 640, height: 480}
	g.grid = make([][]bool, g.width)
	g.buffer = make([][]bool, g.width)
	ebiten.SetWindowSize(g.width, g.height)
	ebiten.SetWindowTitle("Game of life")
	ebiten.SetWindowResizable(true)

	for x := 0; x < g.width; x++ {
		g.grid[x] = make([]bool, g.height)
		g.buffer[x] = make([]bool, g.height)
		for y := 0; y < g.height; y++ {
			g.grid[x][y] = rand.Float32() < 0.5
		}
	}

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
