package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/eiannone/keyboard"
)

type Point struct {
	X, Y int
}

const (
	SNAKE_BODY   = '■'
	SNAKE_HEAD   = '●'
	FOOD         = '♥'
	SPECIAL_FOOD = '⚝'
	WALL         = '█'
	EMPTY        = ' '
)

type Game struct {
	Snake         []Point
	Direction     Point
	Food          Point
	Width, Height int
	Score         int
	Speed         time.Duration
	IsOver        bool
	IsSpecialFood bool
}

var screenBuffer strings.Builder

// Di chuyen con tro
func moveCursor(x, y int) {
	screenBuffer.WriteString(fmt.Sprintf("\033[%d;%dH", y+1, x+1))
}

// Xoa man hinh
func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

// New game
func NewGame(width, height int) *Game {
	startX, startY := width/2, height/2
	snake := []Point{{startX, startY}, {startX - 1, startY}, {startX - 2, startY}}
	game := &Game{
		Width:     width,
		Height:    height,
		Snake:     snake,
		Direction: Point{1, 0},
		Speed:     250 * time.Millisecond,
	}
	game.placeFood()
	return game
}

// Food
func (g *Game) placeFood() {
	for {
		g.Food = Point{rand.Intn(g.Width-2) + 1, rand.Intn(g.Height-2) + 1}
		onSnake := slices.Contains(g.Snake, g.Food)
		if !onSnake {
			break
		}
	}
}

// initialDraw
func (g *Game) initialDraw() {
	clearScreen()

	//Ve tuong
	for i := 0; i < g.Width; i++ {
		moveCursor(i, 0)
		screenBuffer.WriteRune(WALL)
		moveCursor(i, g.Height-1)
		screenBuffer.WriteRune(WALL)
	}

	for i := 0; i < g.Height; i++ {
		moveCursor(0, i)
		screenBuffer.WriteRune(WALL)
		moveCursor(g.Width-1, i)
		screenBuffer.WriteRune(WALL)
	}

	//Ve ran va moi lan dau
	for i, p := range g.Snake {
		moveCursor(p.X, p.Y)
		if i == 0 {
			screenBuffer.WriteRune(SNAKE_HEAD)
		} else {
			screenBuffer.WriteRune(SNAKE_BODY)
		}
	}
	moveCursor(g.Food.X, g.Food.Y)
	screenBuffer.WriteRune(FOOD)
	//In diem so
	moveCursor(0, g.Height)
	screenBuffer.WriteString(fmt.Sprintf("Score: %d", g.Score))

	//In ra man hinh tat ca
	fmt.Print(screenBuffer.String())
	screenBuffer.Reset()
}

// Update screen
func (g *Game) updateScreen(tail Point, ateFood bool) {
	//Dau cu thanh than
	oldHead := g.Snake[1]
	moveCursor(oldHead.X, oldHead.Y)
	screenBuffer.WriteRune(SNAKE_BODY)

	//Ve dau moi
	newHead := g.Snake[0]
	moveCursor(newHead.X, newHead.Y)
	screenBuffer.WriteRune(SNAKE_HEAD)

	// Neu khong an xoa duoi cu
	if !ateFood {
		moveCursor(tail.X, tail.Y)
		screenBuffer.WriteRune(EMPTY)
	} else {
		//Neu an thi ve moi moi
		moveCursor(g.Food.X, g.Food.Y)
		if ranNum := rand.Intn(3); ranNum == 2 {
			g.IsSpecialFood = true
			screenBuffer.WriteRune(SPECIAL_FOOD)
		} else {
			g.IsSpecialFood = false
			screenBuffer.WriteRune(FOOD)
		}

	}
	moveCursor(0, g.Height)
	screenBuffer.WriteString(fmt.Sprintf("Score: %d", g.Score))

	//In ra tat ca
	fmt.Print(screenBuffer.String())
	screenBuffer.Reset()
}

// Logic game
func (g *Game) Update() (Point, bool) {
	head := g.Snake[0]
	tail := g.Snake[len(g.Snake)-1]
	newHead := Point{head.X + g.Direction.X, head.Y + g.Direction.Y}

	//Kiem tra va cham tuong
	if newHead.X <= 0 || newHead.X >= g.Width-1 || newHead.Y <= 0 || newHead.Y >= g.Height-1 {
		g.IsOver = true
		return tail, false
	}

	//Kiem tra va cham than
	for i := 1; i < len(g.Snake); i++ {
		if newHead == g.Snake[i] {
			g.IsOver = true
			return tail, false
		}
	}

	ateFood := false
	if newHead == g.Food {
		ateFood = true
		if g.IsSpecialFood {
			g.Score += 5
		} else {
			g.Score++
		}

		if g.Speed > 50*time.Millisecond {
			g.Speed -= 5 * time.Millisecond
		}
		//Tang them chieu dai
		g.Snake = append([]Point{newHead}, g.Snake...)
		g.placeFood()
	} else {
		copy(g.Snake[1:], g.Snake[:len(g.Snake)-1])
		g.Snake[0] = newHead
	}
	return tail, ateFood
}

func main() {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer keyboard.Close()

	game := NewGame(40, 20)

	//Ve man hinh
	game.initialDraw()

	keyEvents := make(chan keyboard.Key)

	go func() {
		for {
			r, key, err := keyboard.GetKey()
			if err != nil {
				close(keyEvents)
				return
			}
			if key != 0 {
				keyEvents <- key
			} else if r != 0 {
				keyEvents <- keyboard.Key(r)
			}

			keyEvents <- key
		}
	}()

	for !game.IsOver {
		select {
		case key := <-keyEvents:
			if key == keyboard.KeyEsc || key == 'q' || key == 'Q' {
				game.IsOver = true
			}
			currentDir := game.Direction
			switch key {
			case 'w', 'W', keyboard.KeyArrowUp:
				if currentDir.Y == 0 {
					game.Direction = Point{0, -1}
				}
			case 's', 'S', keyboard.KeyArrowDown:
				if currentDir.Y == 0 {
					game.Direction = Point{0, 1}
				}
			case 'a', 'A', keyboard.KeyArrowLeft:
				if currentDir.X == 0 {
					game.Direction = Point{-1, 0}
				}
			case 'd', 'D', keyboard.KeyArrowRight:
				if currentDir.X == 0 {
					game.Direction = Point{1, 0}
				}
			}
		case <-time.After(game.Speed):
			tail, ateFood := game.Update()
			if game.IsOver {
				break
			}
			game.updateScreen(tail, ateFood)
		}
		if game.IsOver {
			break
		}
	}

	//In ra game over
	msg := "GAME OVER!\n"
	moveCursor(game.Width/2-(len(msg)/2), game.Height/2)
	screenBuffer.WriteString(msg)
	moveCursor(0, game.Height+1)
	fmt.Print(screenBuffer.String())
}
