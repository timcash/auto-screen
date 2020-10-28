package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
)

var imageStore = make(map[string]robotgo.Bitmap)
var actionStore = make(map[string]action)

// Config does foo
type Config struct {
	Name    string
	Actions []action
}

type action struct {
	Name string
	Path string
	Key  string
	Mods []string
}

func sendAction(path string) {
	a, ok := actionStore[path]
	if ok {
		if len(a.Mods) > 0 {
			robotgo.KeyTap(a.Key, a.Mods)
		} else {
			robotgo.KeyTap(a.Key)
		}
	}
}

func getAction(path string) (action, bool) {
	a, ok := actionStore[path]
	if ok {
		return a, true
	}
	return action{}, false
}

func addAction(path, name, key string, mods []string) {
	// fmt.Println("storing::", id)
	a := action{Path: path, Name: name, Key: key, Mods: mods}
	actionStore[path] = a
}

func printActions() {
	for _, a := range actionStore {
		fmt.Println(a)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func loadActions(path string) {
	fileHook, err := os.Open(path)
	defer fileHook.Close()
	if err != nil {
		fmt.Println(err.Error())
	}

	jsonParser := json.NewDecoder(fileHook)
	config := Config{}
	jsonParser.Decode(&config)
	for _, a := range config.Actions {
		addAction(a.Path, a.Name, a.Key, a.Mods)
	}
}

func printNBytesOfFile(f *os.File) {
	b1 := make([]byte, 150)
	n1, err := f.Read(b1)
	check(err)
	fmt.Printf("%d bytes: %s\n", n1, string(b1[:n1]))
}

func storeBitmap(id string, b robotgo.Bitmap) {
	// fmt.Println("storing::", id)
	imageStore[id] = b
}

func loadBitmaps(subPath string) {
	var files []string

	root := fmt.Sprintf("./bitmaps/%s", subPath)
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
			storeBitmap(file, bit)
		}
	}
}

func printStore() {
	for k := range imageStore {
		fmt.Println("key", k)
	}
}

func emptyStore() {
	for k := range imageStore {
		delete(imageStore, k)
	}
}

func deleteBitmap(path string) error {
	e := os.Remove(path)
	if e != nil {
		log.Fatal(e)
	}
	return e
}

func isKeyInStore(key string) bool {
	_, ok := imageStore[key]
	return ok
}

func makeBitmapPath(subPath string) string {
	return fmt.Sprintf("bitmaps\\%s\\%s.png", subPath, uuid())
}

func saveBitmapAtCoords(subPath string, x, y, w, h int) string {
	current := robotgo.CaptureScreen(x, y, w, h)
	path := makeBitmapPath(subPath)
	storeBitmap(path, robotgo.ToBitmap(current))
	robotgo.SaveBitmap(current, path)
	return path
}

func getActiveImage(subPath string, x, y, w, h int) (string, bool) {
	searchSpace := robotgo.CaptureScreen(x, y, w, h)
	for k, v := range imageStore {
		bitmapToFind := robotgo.ToCBitmap(v)
		fx, fy := robotgo.FindBitmap(bitmapToFind, searchSpace, 0.01)
		if fx > -1 && fy > -1 {
			return k, true
		}
	}
	// if we did not find the image
	path := makeBitmapPath(subPath)
	storeBitmap(path, robotgo.ToBitmap(searchSpace))
	saveBitmap(path, robotgo.ToBitmap(searchSpace))
	return "none", false
}

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

var wheeldown = false
var wheelup = false

func low() {
	EvChan := hook.Start()
	defer hook.End()

	for ev := range EvChan {
		// fmt.Println("hook: ", ev)
		// fmt.Println("hook: ", ev.Kind)
		if ev.Rawcode == 160 && ev.Kind == 5 {
			fmt.Println("PAUSE")
			wheeldown = false
			wheelup = false
		}
		if ev.Kind == 11 && ev.Rotation > 0 {
			fmt.Println("WHEELDOWN")
			wheeldown = true
			wheelup = false
		}
		if ev.Kind == 11 && ev.Rotation < 0 {
			fmt.Println("WHEELUP")
			wheeldown = false
			wheelup = true
		}
	}
}

func main() {
	actionPath := "config.json"
	subPath := "tests"
	loadBitmaps(subPath)
	loadActions(actionPath)
	printActions()
	go low()
	poll("test", &wheeldown, 5, 5, 48, 48)
	poll("test", &wheelup, 5, 55, 48, 98)
}

func poll(subPath string, pause *bool, x, y, w, h int) {
	currentActionName := "none"
	for {
		time.Sleep(time.Millisecond * 333)
		if *pause == false {
			resultKey, exists := getActiveImage(subPath, x, y, w, h)
			if exists {
				newAction, ok := getAction(resultKey)
				if ok {
					if currentActionName != newAction.Name {
						fmt.Println("ACTION", newAction.Name)
						currentActionName = newAction.Name
					}
					sendAction(resultKey)
				}
			} else {
				fmt.Println("no match")

			}
		}
	}
}
