package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-vgo/robotgo"
)

var imageStore = make(map[string]robotgo.Bitmap)

func main() {
	startPolling1()
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
		if fx == -1 || fy == -1 {
			return "none", false
		}
		return k, true
	}
	// if we did not find the image
	// path := makeBitmapPath(subPath)
	// storeBitmap(path, robotgo.ToBitmap(current))

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

func saveBitmap() string {
	bitmap := robotgo.CaptureScreen(10, 20, 30, 40)
	// use `defer robotgo.FreeBitmap(bit)` to free the bitmap
	defer robotgo.FreeBitmap(bitmap)

	fmt.Println("...", bitmap)

	fx, fy := robotgo.FindBitmap(bitmap)
	fmt.Println("FindBitmap------ ", fx, fy)

	robotgo.SaveBitmap(bitmap, "test.png")
	return "tes1t.png"
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

func clearMemForBitmap() {

}

func doSomething(s string) {
	fmt.Println("doing something", s)
}

func startPolling1() {
	for {
		time.Sleep(2 * time.Second)
		go doSomething("from polling 1")
	}
}
