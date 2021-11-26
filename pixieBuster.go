package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	text "github.com/hajimehoshi/ebiten/text"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	pixieBusterMovementLength = 3
	pixieBusterStartingSize   = 15
	pixelSize                 = 10
	maxPixelOnScreen          = 100
	displayPixels             = 10
	pixelSpawnCadence         = 10 * time.Second
)

// create sprite struct for the pixels
type sprite struct {
	active bool

	x, y  int
	size  int
	color color.Color
}

// create game struct
type Game struct {
	level  int
	ticks  int64
	sx, sy int

	lastPixelSpawnTime time.Time
	levelStartTime     time.Time
	levelDuration      time.Duration

	player         *sprite
	pixel          [maxPixelOnScreen]*sprite
	pixelAmount    int
	pixelCollected int
	pixelLeft      int
}

//logic for busting pixels
func intersects(p1, p2 sprite) bool {
	return ((p1.x >= p2.x-p2.size && p1.x <= p2.x+p2.size && p1.x >= p2.x) &&
		(p1.y >= p2.y-p2.size && p1.y <= p2.y+p2.size && p1.y >= p2.y)) ||
		((p2.x >= p1.x-p1.size && p2.x <= p1.x+p1.size && p2.x >= p1.x) &&
			(p2.y >= p1.y-p1.size && p2.y <= p1.y+p1.size && p2.y >= p1.y))
}


func keyPressed(key ebiten.Key) bool {
	return inpututil.KeyPressDuration(key) > 0
}

//grow pixelBuster when bust pixels
func (p *sprite) grow() {
	p.size++
}

// create new sprite for the game
func (g Game) newSprite(x, y int) *sprite {
	p := &sprite{}
	// create some random coordinates for the pixels
	if x == -1 {
		p.x = rand.Intn(g.sx-100)
	} else {
		p.x = x
	}
	if y == -1 {
		p.y = rand.Intn(g.sy-200)
	} else {
		p.y = y
	}
	p.color = colorful.FastHappyColor()
	p.active = true
	return p
}

// create new pixel
func (g Game) newPixel(x, y int) *sprite {
	//create new sprite
	s := g.newSprite(x, y)
	//set size
	s.size = pixelSize
	return s
}

// create pixieBuster
func (g Game) newPixelBuster(x, y int) *sprite {
	//create new sprite
	s := g.newSprite(x, y)
	//set size
	s.size = pixieBusterStartingSize
	//set color
	s.color = colornames.Red
	return s
}

//define how pixels render in the game screen
func (p *sprite) drawTo(i *ebiten.Image) {
	if !p.active {
		return
	}
	for x := 0; x < p.size; x++ {
		for y := 0; y < p.size; y++ {
			i.Set(p.x+x+x, p.y+y, p.color)
		}
	}
}

// init game level for the player
func (g *Game) init(level int) {
	gameLevel := level
	if level <= 0 {
		gameLevel = 1
	} else if level > 10 {
		gameLevel = 10
	}
	g.level=level
	//increase pixels when pass new level
	g.pixelAmount = gameLevel * displayPixels
	g.sx, g.sy = ebiten.ScreenSizeInFullscreen()
	// assign new pixelBuster struct to the g.player
	g.player = g.newPixelBuster(g.sx/2, g.sy/2)
	//define left pixels count
	g.pixelLeft = g.pixelAmount
	// assign new pixel struct to the g.pixel variable
	for i := 0; i < g.pixelAmount; i++ {
		g.pixel[i] = g.newPixel(-1, -1)
	}
	now := time.Now()
	g.lastPixelSpawnTime = now
	g.levelStartTime = now
}

//key controller logics
func (g *Game) KeyControllers(screen *ebiten.Image) *sprite {
	playerPixel := g.player
	// move player right
	if keyPressed(ebiten.KeyRight) && playerPixel.x+playerPixel.size <= g.sx {
		playerPixel.x = int(math.Min(float64(playerPixel.x+pixieBusterMovementLength), float64(g.sx-2*playerPixel.size)))
		// Game Exit when pixelBuster touch right corner of GameScreen
		if playerPixel.x+2*playerPixel.size ==g.sx{
			text.Draw(screen,"GameOver", basicfont.Face7x13, g.sx/2, g.sy/2, color.White)
			time.Sleep(2* time.Second)
			os.Exit(0)
		}
	}
	// move player left
	if keyPressed(ebiten.KeyLeft) && playerPixel.x >= 0 {
		playerPixel.x =  int(math.Max(float64(playerPixel.x-pixieBusterMovementLength), 0))
		// Game Exit when pixelBuster touch left corner of GameScreen
		if playerPixel.x <=0{
			text.Draw(screen,"GameOver", basicfont.Face7x13, g.sx/2, g.sy/2, color.White)
			time.Sleep(2* time.Second)
			os.Exit(0)
		}
	}
	//move player up
	if keyPressed(ebiten.KeyUp) && playerPixel.y >= 0 {
		playerPixel.y = int(math.Max(float64(playerPixel.y-pixieBusterMovementLength), 0))
		// Game Exit when pixelBuster touch top corner of GameScreen
		if playerPixel.y <=0{
			text.Draw(screen,"GameOver", basicfont.Face7x13, g.sx/2, g.sy/2, color.White)
			time.Sleep(2* time.Second)
			os.Exit(0)
		}
	}
	//move player down
	if keyPressed(ebiten.KeyDown) && playerPixel.y+playerPixel.size <= g.sy {
		playerPixel.y = int(math.Min(float64(playerPixel.y+pixieBusterMovementLength), float64(g.sy-playerPixel.size)))
		// Game Exit when pixelBuster touch bottom corner of GameScreen
		if playerPixel.y+playerPixel.size ==g.sy{
			text.Draw(screen,"GameOver", basicfont.Face7x13, g.sx/2, g.sy/2, color.White)
			time.Sleep(2* time.Second)
			os.Exit(0)
		}
	}
	return playerPixel
}

// update game screen and other functions
func (g *Game) Update(screen *ebiten.Image) (err error) {
	// exit game or next level
	defer func() {
		if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			os.Exit(0)
		} else if g.pixelLeft == 0 && inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			// increase level
			g.level++
			//initialize new level
			g.init(g.level)
		}
		err = g.Draw(screen)
	}()
	if g.pixelLeft == 0 {
		return
	}
	now := time.Now()
	// set duration of level
	g.levelDuration = now.Sub(g.levelStartTime)
	// set key controllers of player
	player := g.KeyControllers(screen)
	// increase size and change color when collecting pixel
	for i := 0; i < g.pixelAmount; i++ {
		fd := g.pixel[i]
		if fd.active && intersects(*fd, *player) {
			if g.pixelCollected%1 == 0 {
				player.grow()
				player.color = fd.color

			}
			fd.active = false
			g.pixelLeft--
			g.pixelCollected++
		}
	}

	//add pixel left count when random pixel render in game window
	if g.ticks%50 == 0 {
		if now.Add(-pixelSpawnCadence).After(g.lastPixelSpawnTime) {
			for i := 0; i < g.pixelAmount; i++ {
				if !g.pixel[i].active {
					g.pixel[i] = g.newPixel(-1, -1)
					g.pixelLeft++
					g.lastPixelSpawnTime = now
					break
				}
			}
		}
	}
	return
}

// background screen draw for the game details
func (g *Game) Draw(screen *ebiten.Image) error {
	if err := screen.Fill(color.Black); err != nil {
		return err
	}
	// render pixel buster in game screen
	g.player.drawTo(screen)
	// render pixels in game screen
	for i := 0; i < g.pixelAmount; i++ {
		g.pixel[i].drawTo(screen)
	}
	g.KeyControllers(screen)
	// show title in game screen
	text.Draw(screen, "Pixel Buster", basicfont.Face7x13, g.sx/2, 15, color.White)
	// show count of left pixels
	text.Draw(screen, fmt.Sprintf("Pixel Left : %d", g.pixelLeft), basicfont.Face7x13, 10, 15, colorful.FastHappyColor())
	// show count of collected pixels
	text.Draw(screen, fmt.Sprintf("Collected Pixel : %d", g.pixelCollected), basicfont.Face7x13, 10, 30, colorful.FastHappyColor())
	//set time
	hours := int(g.levelDuration.Hours())
	minutes := int(g.levelDuration.Minutes())
	secs := int(g.levelDuration.Seconds())
	//show level number in game window
	text.Draw(screen, fmt.Sprintf("Level : %d", g.level), basicfont.Face7x13, g.sx-100, 15, colorful.FastHappyColor())
	text.Draw(screen, fmt.Sprintf("%02d:%02d:%02d", hours, minutes-(hours*60), secs-(minutes*60)), basicfont.Face7x13, g.sx-100, 30, colorful.FastHappyColor())
	//when complete level
	if g.pixelLeft == 0 {
		text.Draw(screen, "Complete!", basicfont.Face7x13, g.sx/2, g.sy/2, colorful.FastHappyColor())
		text.Draw(screen, "Press Enter to the next level...", basicfont.Face7x13, g.sx/2, g.sy-800/2, colorful.FastHappyColor())
	}
	return nil
}

// game layout size
func (g *Game) Layout(_, _ int) (int, int) {
	return g.sx, g.sy
}

func main() {
	//create new game object
	game := &Game{}
	//set game title
	ebiten.SetWindowTitle("Pixel Buster")
	//set to full screen mode
	ebiten.SetFullscreen(true)
	//cursor hide
	ebiten.SetCursorMode(ebiten.CursorModeHidden)
	//set starting level
	game.init(1)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
