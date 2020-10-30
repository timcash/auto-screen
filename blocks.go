package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/hajimehoshi/ebiten/examples/resources/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	hook "github.com/robotn/gohook"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// ===========================
// 			HELPERS
// ===========================

func timestampMillisecond() int64 {
	return time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

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

// ===========================
// 			GAME
// ===========================

// Game the game
type Game struct {
}

var emptyAction = NewEmptyAction()
var lastPool int64 = 0
var wheelDirection = 0.0
var wheelDown = false
var wheelUp = false
var ctlDown = false
var shiftDown = false
var altDown = false
var paused = false
var shiftWheelUp = false
var shiftWheelDown = false

// Update the update
func (g *Game) Update() error {
	t := timestampMillisecond()
	if t-lastPool > 222 && paused == false {
		lastPool = t
		if wheelDown {
			w1.poll()
		}
		if wheelUp {
			w2.poll()
			//w2poll("tests", 5, 62, 48, 48)
		}
		if shiftWheelDown {
			// fmt.Println("shift DOWN")
			w3.poll()
		}
		if shiftWheelUp {
			// fmt.Println("shift UP")
			w4.poll()
		}
	}
	return nil
}

// Draw the draw
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw info
	msg := fmt.Sprintf("TPS: %0.2f Paused: %t Shift: %v", ebiten.CurrentTPS(), paused, kt1.timeDownInMs("shift"))
	text.Draw(screen, msg, baseFont, 5, 20, color.White)

	if w1.currentAction.hasImage {
		screen.DrawImage(w1.currentAction.renderImg, w1.renderOps)
		x := w1.renderX + w1.width + 5
		text.Draw(screen, w1.currentAction.Key, baseFont, x, w1.renderY+14, color.White)
	}
	if w2.currentAction.hasImage {
		x := w2.renderX + w2.width + 5
		screen.DrawImage(w2.currentAction.renderImg, w2.renderOps)
		text.Draw(screen, w2.currentAction.Key, baseFont, x, w2.renderY+14, color.White)
	}
	if w3.currentAction.hasImage {
		x := w3.renderX + w3.width + 5
		screen.DrawImage(w3.currentAction.renderImg, w3.renderOps)
		text.Draw(screen, w3.currentAction.Key, baseFont, x, w3.renderY+14, color.White)
	}
	if w4.currentAction.hasImage {
		x := w4.renderX + w4.width + 5
		screen.DrawImage(w4.currentAction.renderImg, w4.renderOps)
		text.Draw(screen, w4.currentAction.Key, baseFont, x, w4.renderY+14, color.White)
	}
}

// Layout the layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

// ===========================
// 			WATCH
// ===========================

type watch struct {
	name                string
	offset              int
	subPath             string
	actionPath          string
	renderOps           *ebiten.DrawImageOptions
	actionMap           map[string]Action
	currentAction       Action
	x, y, width, height int
	renderX, renderY    int
}

func (w *watch) setRenderOps(x, y int) {
	w.setRenderArea(x, y)
	w.renderOps = &ebiten.DrawImageOptions{}
	w.renderOps.GeoM.Translate(float64(w.renderX), float64(w.renderY))
	w.renderOps.GeoM.Scale(1, 1)
}

func (w *watch) addImageOnlyAction(path string) {
	a := NewEmptyAction()
	a.Path = path
	a.setRenderImage(a.Path)
	a.setSearchImage(a.Path)
	a.hasImage = true
	w.actionMap[a.Path] = a
}

func (w *watch) addConfigAction(a Action) {
	a.setRenderImage(a.Path)
	a.setSearchImage(a.Path)
	a.hasImage = true
	a.hasConfig = true
	w.actionMap[a.Path] = a
}

func (w *watch) setCoords(x, y, width, height int) {
	w.x = x
	w.y = y
	w.width = width
	w.height = height
}

func (w *watch) setRenderArea(x, y int) {
	w.renderX = x
	w.renderY = y
}

func (w *watch) poll() {
	// find a match in image store
	matchedAction, searchSpace, exists := w.checkBitmapForMatch()
	w.currentAction = matchedAction
	if exists && matchedAction.hasConfig {
		fmt.Println("sending keys", matchedAction)
		if len(matchedAction.Mods) > 0 {
			robotgo.KeyTap(matchedAction.Key, matchedAction.Mods)
			return
		}
		robotgo.KeyTap(matchedAction.Key)
		return
	} else if exists {
		// exists but not in config yet do nothing
		return
	}
	// if we did not find the image save the image to the disk
	fmt.Println("new image")
	path := makeBitmapPath(w.subPath)
	saveBitmap(path, searchSpace)
	// save it to our actionMap so we match on the next round
	w.addImageOnlyAction(path)
}

func (w *watch) checkBitmapForMatch() (Action, robotgo.Bitmap, bool) {
	searchSpace := robotgo.CaptureScreen(w.x, w.y, w.width, w.height)
	for _, actionToTest := range w.actionMap {
		bitmapToFind := robotgo.ToCBitmap(actionToTest.searchImg)
		fx, fy := robotgo.FindBitmap(bitmapToFind, searchSpace, 0.3)
		if fx > -1 && fy > -1 {
			return actionToTest, actionToTest.searchImg, true
		}
	}
	return emptyAction, robotgo.ToBitmap(searchSpace), false
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
			w.addImageOnlyAction(file)
		}
	}
}

func (w *watch) printActions() {
	for k, v := range w.actionMap {
		fmt.Println(k, v)
	}
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
		w.addConfigAction(a)
	}
}

// ===========================
// 			ACTION
// ===========================

// Action for actions
type Action struct {
	hasConfig bool
	Name      string
	Path      string
	Key       string
	Mods      []string
	hasImage  bool
	renderImg *ebiten.Image
	searchImg robotgo.Bitmap
}

// NewEmptyAction is new
func NewEmptyAction() Action {
	a := Action{hasConfig: false, Name: "none", hasImage: false, Path: ".", Key: "null", Mods: []string{"shift", "ctrl", "alt"}}
	return a
}

// SendAction is new
func (a *Action) SendAction() {
	if a.hasConfig {
		if len(a.Mods) > 0 {
			robotgo.KeyTap(a.Key, a.Mods)
		} else {
			robotgo.KeyTap(a.Key)
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

func (a *Action) setSearchImage(file string) {
	bit := openBitmap(file)
	a.searchImg = bit
}

// ===========================
// 			MAIN
// ===========================

var w1 watch = watch{name: "down"}
var w2 watch = watch{name: "up"}
var w3 watch = watch{name: "shift-down"}
var w4 watch = watch{name: "shift-up"}

var nameToCode = hook.Keycode
var codeToName = make(map[uint16]string)
var kt1 = newKeyTracker()

var baseFont font.Face

func main() {

	tt, err := opentype.Parse(fonts.MPlus1pRegular_ttf)
	if err != nil {
		log.Fatal(err)
	}

	const dpi = 72
	baseFont, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    14,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	if err != nil {
		log.Fatal(err)
	}

	fillCodeToName()
	w1.actionPath = "config.json"
	w1.subPath = "tests"
	w1.actionMap = make(map[string]Action)
	w1.currentAction = emptyAction
	// w1.currentAction.setRenderImage("bitmaps\\tests\\7c91bf41-d1d5-f1a5-5751-e1b3084b84e1.png")
	w1.loadBitmaps()
	w1.loadActions()
	w1.printActions()
	w1.setCoords(5, 5+(58*0), 48, 48)
	w1.setRenderOps(5.0, 30.0+(50.0*0))

	w2.actionPath = "config.json"
	w2.subPath = "tests"
	w2.actionMap = make(map[string]Action)
	w2.currentAction = emptyAction
	// w2.currentAction.setRenderImage("bitmaps\\tests\\7c91bf41-d1d5-f1a5-5751-e1b3084b84e1.png")
	w2.loadBitmaps()
	w2.loadActions()
	w2.setCoords(5, 5+(58*1), 48, 48)
	w2.setRenderOps(5.0, 30.0+(50.0*1))

	w3.actionPath = "config.json"
	w3.subPath = "tests"
	w3.actionMap = make(map[string]Action)
	w3.currentAction = emptyAction
	// w3.currentAction.setRenderImage("bitmaps\\tests\\7c91bf41-d1d5-f1a5-5751-e1b3084b84e1.png")
	w3.loadBitmaps()
	w3.loadActions()
	w3.setCoords(5, 5+(58*2)-1, 48, 48)
	w3.setRenderOps(5.0, 30.0+(50.0*2))

	w4.actionPath = "config.json"
	w4.subPath = "tests"
	w4.actionMap = make(map[string]Action)
	w4.currentAction = emptyAction
	// w4.currentAction.setRenderImage("bitmaps\\tests\\7c91bf41-d1d5-f1a5-5751-e1b3084b84e1.png")
	w4.loadBitmaps()
	w4.loadActions()
	w4.setCoords(5, 5+(58*3)-1, 48, 48)
	w4.setRenderOps(5.0, 30.0+(50.0*3))

	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("Tracker")

	go hookAllInput()
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}

// ===========================
// 			HOOKS
// ===========================
func fillCodeToName() {
	for k, v := range nameToCode {
		codeToName[v] = k
	}
}

func hookAllInput() {
	EvChan := hook.Start()
	defer hook.End()

	for ev := range EvChan {
		if ev.Kind == hook.MouseMove || ev.Kind == hook.MouseDrag {
			// do nothing
		} else if ev.Kind == hook.KeyHold || ev.Kind == hook.KeyUp || ev.Kind == hook.KeyDown {
			if ev.Kind == hook.KeyHold || ev.Kind == hook.KeyDown {
				kt1.setDownByCode(ev.Keycode)
			}
			if ev.Kind == hook.KeyUp {
				kt1.setUpByCode(ev.Keycode)
			}
		}
		if ev.Kind == hook.MouseWheel && ev.Rotation > 0 {
			paused = false
			wheelDown = true
			wheelUp = false
			shiftWheelUp = false
			shiftWheelDown = false
			if kt1.isDown("shift") {
				wheelUp = false
				wheelDown = false
				shiftWheelUp = false
				shiftWheelDown = true
			}
		}
		if ev.Kind == hook.MouseWheel && ev.Rotation < 0 {
			paused = false
			wheelUp = true
			wheelDown = false
			shiftWheelUp = false
			shiftWheelDown = false
			if kt1.timeDownInMs("shift") > 1 {
				wheelUp = false
				wheelDown = false
				shiftWheelUp = true
				shiftWheelDown = false
			}
		}
		if kt1.timeDownInMs("shift") > 200 {
			paused = true
			wheelUp = false
			wheelDown = false
		}
	}
}

func print(s string) {
	fmt.Println(s)
}

// ===========================
// 	      KEY TRACKER
// ===========================
type keyTracker struct {
	db map[uint16]*keyStats
}

func newKeyTracker() keyTracker {
	database := make(map[uint16]*keyStats)
	return keyTracker{db: database}
}

func (kt *keyTracker) timeDownInMs(keyname string) int64 {
	code := nameToCode[keyname]
	if stat, ok := kt.db[code]; ok {
		return stat.timeDownInMs()
	}
	return 0
}

func (kt *keyTracker) isDown(keyname string) bool {
	code := nameToCode[keyname]
	if stat, ok := kt.db[code]; ok {
		return stat.down
	}
	return false
}

func (kt *keyTracker) isUp(keyname string) bool {
	code := nameToCode[keyname]
	if stat, ok := kt.db[code]; ok {
		return stat.isUp()
	}
	return false
}

func (kt *keyTracker) setDownByCode(code uint16) int64 {
	if stat, inDb := kt.db[code]; inDb {
		t := stat.setDown()
		return t
	}

	n := kt.newKeyStats(codeToName[code], code)
	t := n.setDown()
	fmt.Println("DB ADD", n)
	return t
}

func (kt *keyTracker) setUpByCode(code uint16) int64 {
	if stat, inDb := kt.db[code]; inDb {
		t := stat.setUp()
		return t
	}

	n := kt.newKeyStats(codeToName[code], code)
	t := n.setUp()
	return t
}

func (kt *keyTracker) newKeyStats(name string, keycode uint16) *keyStats {
	t := timestampMillisecond()
	kt.db[keycode] = &keyStats{name: name, down: false, keycode: keycode, holdStartTime: t, holdEndTime: t + 1}
	return kt.db[keycode]
}

// ===========================
// 			KEY STATS
// ===========================

type keyStats struct {
	name          string
	keycode       uint16
	down          bool
	holdStartTime int64
	holdEndTime   int64
}

func (ks *keyStats) setDown() int64 {
	if ks.down == false {
		ks.down = true
		ks.holdStartTime = timestampMillisecond()
		return ks.holdStartTime
	}
	return ks.holdStartTime
}

func (ks *keyStats) setUp() int64 {
	t := timestampMillisecond()
	ks.down = false
	ks.holdEndTime = t
	return ks.holdEndTime
}

func (ks *keyStats) isUp() bool {
	// if startTime > endTime then it is down
	return ks.holdStartTime <= ks.holdEndTime
}

func (ks *keyStats) timeDownInMs() int64 {
	if ks.isUp() {
		return 0
	}
	return timestampMillisecond() + 1 - ks.holdStartTime
}

// func (kt *keyTracker) setDown(keyname string) int64 {
// 	if code, hasCode := nameToCode[keyname]; hasCode {
// 		if stat, inDb := kt.db[code]; inDb {
// 			return stat.setDown()
// 		}

// 		n := kt.newKeyStats(keyname, code)
// 		return n.setDown()
// 	}
// 	// we don't track this keycode
// 	return 0
// }

// func (kt *keyTracker) setUp(keyname string) int64 {
// 	if code, hasCode := nameToCode[keyname]; hasCode {
// 		if stat, inDb := kt.db[code]; inDb {
// 			return stat.setUp()
// 		}

// 		n := kt.newKeyStats(keyname, code)
// 		return n.setUp()
// 	}
// 	// we don't track this keycode
// 	return 0
// }
