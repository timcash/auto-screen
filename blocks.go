package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-vgo/robotgo"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Config does foo
type Config struct {
	Name    string
	Actions []Action
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func printNBytesOfFile(f *os.File) {
	b1 := make([]byte, 150)
	n1, err := f.Read(b1)
	check(err)
	fmt.Printf("%d bytes: %s\n", n1, string(b1[:n1]))
}

func deleteBitmap(path string) error {
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
	return e
}

func makeBitmapPath(subPath string) string {
	return fmt.Sprintf("bitmaps\\%s\\%s.png", subPath, uuid())
}

// func saveBitmapAtCoords(subPath string, x, y, w, h int) string {
// 	current := robotgo.CaptureScreen(x, y, w, h)
// 	path := makeBitmapPath(subPath)
// 	storeBitmap(path, robotgo.ToBitmap(current))
// 	robotgo.SaveBitmap(current, path)
// 	return path
// }

func uuid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	id := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return id
}

func saveBitmap(path string, bitmap robotgo.Bitmap) {
	robotgo.SaveBitmap(robotgo.ToCBitmap(bitmap), path)
}

func getColor(inBit robotgo.Bitmap, x, y int) string {
	//bitmap := robotgo.CaptureScreen(10, 20, 30, 40)
	// use `defer robotgo.FreeBitmap(bit)` to free the bitmap
	bitmap := robotgo.ToCBitmap(inBit)
	color := robotgo.GetColors(bitmap, x, y)
	return color
	//return "222222"
}

func openBitmap(path string) robotgo.Bitmap {
	bitmap := robotgo.OpenBitmap(path)
	return robotgo.ToBitmap(bitmap)
}

// Game the game
type Game struct{}

var emptyAction = NewAction()
var counter = 10
var wheelDirection = 0.0
var wheelDown = false
var wheelUp = false
var shiftDown = false

// Update the update
func (g *Game) Update() error {
	counter--
	_, newWheelDirection := ebiten.Wheel()
	if wheelDirection != newWheelDirection {
		wheelDirection = newWheelDirection
	}
	if wheelDirection > 0 {
		wheelUp = true
		wheelDown = false
	}
	if wheelDirection < 0 {
		wheelDown = true
		wheelUp = false
	}
	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		shiftDown = true
	} else {
		shiftDown = false
	}

	if counter == 0 {

		counter = 10
		if wheelDown {
			fmt.Println("wheelDown")
			w1.poll()
		}
		if wheelUp {
			fmt.Println("wheelUp")
			//w2poll("tests", 5, 62, 48, 48)
		}
		if wheelDown && shiftDown {
			fmt.Println("shift DOWN")
		}
		if wheelUp && shiftDown {
			fmt.Println("shift UP")
		}
	}
	return nil
}

func (w *watch) loadActions() {
	fileHook, err := os.Open(w.actionPath)
	defer fileHook.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	jsonParser := json.NewDecoder(fileHook)
	config := Config{}
	jsonParser.Decode(&config)
	for _, a := range config.Actions {
		if existingAction, ok := w.actionMap[a.Path]; ok {
			existingAction.Exists = true
			existingAction.Path = a.Path
			existingAction.Name = a.Name
			existingAction.Key = a.Key
			existingAction.Mods = a.Mods

		} else {
			a.Exists = true
			w.actionMap[a.Path] = a
		}
	}
}

func (w *watch) loadBitmaps() {
	var files []string

	root := fmt.Sprintf("./bitmaps/%s", w.subPath)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		ext := filepath.Ext(file)
		if ext == ".png" {
			bit := openBitmap(file)
			if a, ok := w.actionMap[file]; ok {
				a.searchImg = bit
				a.setRenderImage(file)

			} else {
				b := NewAction()
				b.searchImg = bit
				b.setRenderImage(file)

				w.actionMap[file] = b
			}
		}
	}
}

func (a *Action) setRenderImage(file string) {
	var err error
	a.renderImg, _, err = ebitenutil.NewImageFromFile(file)

	if err != nil {
		log.Fatal(err)
	}
}

func (w *watch) setRenderOps() {
	w1.renderOps = &ebiten.DrawImageOptions{}
	w1.renderOps.GeoM.Translate(0, 0)
	w1.renderOps.GeoM.Scale(1, 1)
}

func (w *watch) setCoords(x, y, width, height int) {
	w.x = x
	w.y = y
	w.width = width
	w.height = height
}

func (w *watch) poll() {
	// find a match in image store
	matchedAction, searchSpace := w.checkBitmapForMatch()
	w.currentAction = matchedAction
	fmt.Println("w.currentAction")
	fmt.Println(w.currentAction)
	if matchedAction.Exists && len(matchedAction.Mods) > 0 {
		robotgo.KeyTap(matchedAction.Key, matchedAction.Mods)
		return
	}

	if matchedAction.Exists {
		robotgo.KeyTap(matchedAction.Key)
		return
	}
	// if we did not find the image save the image to the disk
	path := makeBitmapPath(w.subPath)
	saveBitmap(path, searchSpace)
}

func (w *watch) checkBitmapForMatch() (Action, robotgo.Bitmap) {
	fmt.Println("checkBitmapForMatch")
	searchSpace := robotgo.CaptureScreen(w.x, w.y, w.width, w.height)
	for _, actionToTest := range w.actionMap {
		bitmapToFind := robotgo.ToCBitmap(actionToTest.searchImg)
		fx, fy := robotgo.FindBitmap(bitmapToFind, searchSpace, 0.3)
		if fx > -1 && fy > -1 {
			return actionToTest, actionToTest.searchImg
		}
	}
	fmt.Println("no match found")
	return emptyAction, robotgo.ToBitmap(searchSpace)
}

// Draw the draw
func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(w1.currentAction.renderImg, w1.currentAction.Key)
	screen.DrawImage(w1.currentAction.renderImg, w1.renderOps)
}

// Layout the layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

type watch struct {
	name                string
	offset              int
	subPath             string
	actionPath          string
	renderOps           *ebiten.DrawImageOptions
	actionMap           map[string]Action
	currentAction       Action
	x, y, width, height int
}

// Action for actions
type Action struct {
	Exists    bool
	Name      string
	Path      string
	Key       string
	Mods      []string
	renderImg *ebiten.Image
	searchImg robotgo.Bitmap
}

// NewAction is new
func NewAction() Action {
	a := Action{Exists: false, Name: "none", Path: ".", Key: "i", Mods: []string{"shift", "ctrl", "alt"}}
	return a
}

// SendAction is new
func (a *Action) SendAction() {
	if a.Exists {
		if len(a.Mods) > 0 {
			robotgo.KeyTap(a.Key, a.Mods)
		} else {
			robotgo.KeyTap(a.Key)
		}
	}
}

var w1 watch = watch{name: "down"}
var w2 watch = watch{name: "up"}

func main() {
	w1.actionPath = "config.json"
	w1.subPath = "tests"
	w1.actionMap = make(map[string]Action)
	w1.currentAction = emptyAction
	w1.currentAction.setRenderImage("bitmaps\\tests\\7c91bf41-d1d5-f1a5-5751-e1b3084b84e1.png")
	w1.loadBitmaps()
	w1.loadActions()
	w1.setCoords(5, 5, 48, 48)
	w1.setRenderOps()

	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
