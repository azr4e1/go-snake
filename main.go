package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenWidth     = 640
	screenHeight    = 480
	antiAlias       = true
	tileSize        = 10
	horizontalTiles = screenWidth / tileSize
	verticalTiles   = screenHeight / tileSize
	maxSpeed        = 50
)

var (
	dirUp    = Tile{0, -1}
	dirDown  = Tile{0, 1}
	dirLeft  = Tile{-1, 0}
	dirRight = Tile{1, 0}
)

type Tile struct {
	x, y int
}

func (t Tile) X() float32 {
	return float32(t.x * tileSize)
}

func (t Tile) Y() float32 {
	return float32(t.y * tileSize)
}

type Game struct {
	snake                []Tile // first element is head
	headColor, bodyColor color.RGBA
	deadColor            color.RGBA
	food                 []Tile
	foodColor            color.RGBA
	occupiedTiles        map[Tile]bool // manages occupied state of each tile
	direction            Tile          // (1,0) right (-1,0) left (0,-1) up (0,1) down
	updateTick           int           // keep  track of current tick
	speed                int
	wallTiles            []Tile
	wallColor            color.RGBA
	isPlaying            bool
	score                int64 // to keep track of score
	scoreFontFace        font.Face
	scoreTextColor       color.RGBA
	scoreText            string
	scorePosition        Tile
}

func (g *Game) Update() error {
	if !g.isPlaying {
		return nil
	}

	tps := ebiten.TPS() // get ticks per second
	g.updateTick++      //  track every tick
	// Detect space pressed
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyQ) && inpututil.IsKeyJustPressed(ebiten.KeyQ):
		os.Exit(0)
	}

	// determine snake direction
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp) && inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		if g.direction != dirDown {
			g.direction = dirUp
		}
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown) && inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		if g.direction != dirUp {
			g.direction = dirDown
		}
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		if g.direction != dirRight {
			g.direction = dirLeft
		}
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) && inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		if g.direction != dirLeft {
			g.direction = dirRight
		}
	case ebiten.IsKeyPressed(ebiten.KeyK) && inpututil.IsKeyJustPressed(ebiten.KeyK):
		if g.direction != dirDown {
			g.direction = dirUp
		}
	case ebiten.IsKeyPressed(ebiten.KeyJ) && inpututil.IsKeyJustPressed(ebiten.KeyJ):
		if g.direction != dirUp {
			g.direction = dirDown
		}
	case ebiten.IsKeyPressed(ebiten.KeyH) && inpututil.IsKeyJustPressed(ebiten.KeyH):
		if g.direction != dirRight {
			g.direction = dirLeft
		}
	case ebiten.IsKeyPressed(ebiten.KeyL) && inpututil.IsKeyJustPressed(ebiten.KeyL):
		if g.direction != dirLeft {
			g.direction = dirRight
		}
	case ebiten.IsKeyPressed(ebiten.KeyEqual) && inpututil.IsKeyJustPressed(ebiten.KeyEqual):
		g.increaseSpeed(1, maxSpeed)
	case ebiten.IsKeyPressed(ebiten.KeyMinus) && inpututil.IsKeyJustPressed(ebiten.KeyMinus):
		g.decreaseSpeed(1, 1)
	}

	// Move snake if the tick is half-way through TPS (creating 2-tile move per second)
	if g.updateTick%(tps/g.speed) == 0 {
		// Move the snake in the direction, by adding new head to the slice, and removing the tail
		head := g.snake[0]
		newHead := Tile{
			x: head.x + g.direction.x,
			y: head.y + g.direction.y,
		}

		// determine if there is food
		var hasFood bool
		if val, ok := g.occupiedTiles[newHead]; ok && val {
			// if tile is already occupied, maybe it's food
			for i, food := range g.food {
				if food == newHead {
					hasFood = true
					g.food = append(g.food[:i], g.food[i+1:]...)
					break
				}
			}
			if !hasFood {
				g.isPlaying = false
				// os.Exit(1)
			}
		}
		if g.isPlaying {
			g.occupiedTiles[newHead] = true

			if hasFood {
				g.snake = append([]Tile{newHead}, g.snake...)
				g.spawnFood()
				g.increaseSpeed(1, maxSpeed)
				g.updateScore(g.score + 1)
			} else {
				delete(g.occupiedTiles, g.snake[len(g.snake)-1])
				g.snake = append([]Tile{newHead}, g.snake[:len(g.snake)-1]...)
			}
		}
	}

	return nil
}

func (g *Game) updateScore(score int64) {
	g.score = score
	g.scoreText = fmt.Sprintf("%d", g.score)

	// determine  text size
	bounds, _ := font.BoundString(g.scoreFontFace, g.scoreText)

	g.scorePosition.y = tileSize * 1.5
	g.scorePosition.x = (screenWidth - bounds.Max.X.Round()) / 2
}

func (g *Game) increaseSpeed(n, maxSpeed int) {
	if n < 0 {
		return
	}
	g.speed = min(g.speed+n, maxSpeed)
}

func (g *Game) decreaseSpeed(n, minSpeed int) {
	if n < 0 {
		return
	}
	g.speed = max(g.speed-n, minSpeed)
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Draw walls
	for _, wallTile := range g.wallTiles {
		vector.DrawFilledRect(screen, wallTile.X(), wallTile.Y(), tileSize, tileSize, g.wallColor, antiAlias)
	}

	headColor, bodyColor := g.headColor, g.bodyColor
	if !g.isPlaying {
		headColor, bodyColor = g.deadColor, g.deadColor
	}

	// Draw snake head (position 0)
	snakeHead := g.snake[0]
	vector.DrawFilledRect(screen, snakeHead.X(), snakeHead.Y(), float32(tileSize), float32(tileSize), headColor, antiAlias)

	// Draw the rest of the snake body (position 1 and onwards)
	for i := 1; i < len(g.snake); i++ {
		snakeBody := g.snake[i]
		vector.DrawFilledRect(screen, snakeBody.X(), snakeBody.Y(), float32(tileSize), float32(tileSize), bodyColor, antiAlias)
	}

	// Draw food
	for _, food := range g.food {
		vector.DrawFilledRect(screen, food.X(), food.Y(), tileSize, tileSize, g.foodColor, antiAlias)
	}

	// Draw High score
	text.Draw(screen, g.scoreText, g.scoreFontFace, g.scorePosition.x, g.scorePosition.y, g.scoreTextColor)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) spawnFood() {
	newFood := Tile{rand.Intn(horizontalTiles), rand.Intn(verticalTiles-2) + 2}
	for g.occupiedTiles[newFood] {
		newFood = Tile{rand.Intn(horizontalTiles), rand.Intn(verticalTiles-2) + 2}
	}
	g.food = append(g.food, newFood)
	g.occupiedTiles[newFood] = true
}

func initGame() *Game {
	var wallTiles = []Tile{}
	occupiedTiles := make(map[Tile]bool)
	for i := 0; i < horizontalTiles; i++ {
		wallTileTop, wallTileBottom := Tile{i, 2}, Tile{i, verticalTiles - 1}
		wallTiles = append(wallTiles, wallTileTop, wallTileBottom)
		occupiedTiles[wallTileBottom] = true
		occupiedTiles[wallTileTop] = true
	}
	for i := 2; i < verticalTiles; i++ {
		wallTileLeft, wallTileRight := Tile{0, i}, Tile{horizontalTiles - 1, i}
		wallTiles = append(wallTiles, wallTileLeft, wallTileRight)
		occupiedTiles[wallTileRight] = true
		occupiedTiles[wallTileLeft] = true
	}
	snake := []Tile{{3, 3}, {2, 3}, {1, 3}}
	for _, t := range snake {
		occupiedTiles[t] = true
	}

	g := &Game{
		snake:          snake,                      // head, body, body
		bodyColor:      color.RGBA{0, 135, 0, 255}, // Dark green
		headColor:      color.RGBA{0, 255, 0, 255}, // Bright green
		foodColor:      color.RGBA{200, 0, 0, 255}, // Red
		occupiedTiles:  occupiedTiles,
		direction:      Tile{1, 0},
		speed:          5,
		wallTiles:      wallTiles,
		wallColor:      color.RGBA{105, 105, 105, 255},
		isPlaying:      true,
		deadColor:      color.RGBA{150, 25, 75, 255},
		scoreTextColor: color.RGBA{255, 255, 255, 255}, // White
	}

	g.spawnFood()

	// parse font to use
	pixelFont, err := opentype.Parse(fonts.PressStart2P_ttf)
	if err != nil {
		log.Panicln(err)
	}

	// create facce based on the font and the line height we want by pixels
	g.scoreFontFace, err = opentype.NewFace(pixelFont, &opentype.FaceOptions{
		Size:    tileSize,
		DPI:     80,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		log.Panicln(err)
	}
	g.scoreFontFace = text.FaceWithLineHeight(g.scoreFontFace, float64(tileSize))
	g.updateScore(0)

	return g
}

func main() {
	ebiten.SetWindowTitle("Snake Game")
	ebiten.SetWindowSize(screenWidth, screenHeight)

	game := initGame()
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
