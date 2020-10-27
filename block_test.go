package main

import (
	"fmt"
	"reflect"
	"testing"
)

var testImagePath string
var subPath = "tests"

func TestSerial(t *testing.T) {
	testImagePath := saveBitmapAtCoords(subPath, 10, 20, 30, 40)
	loadBitmaps(subPath)
	if isKeyInStore(testImagePath) != true {
		t.Fatalf("%s did not exists", testImagePath)
	}
	openBitmap(testImagePath)
	resultKey, exists := getActiveImage(subPath, 0, 0, 64, 64)
	fmt.Println("result match", resultKey)
	if exists == false {
		t.Fatalf("no match for active image")
	}
	deleteError := deleteBitmap(testImagePath)
	if deleteError != nil {
		t.Fatalf("could not delete %s", testImagePath)
	}
}

// func TestSaveBitmapAtCoords(t *testing.T) {
// 	resultID, testImagePath := saveBitmapAtCoords(subPath, 10, 20, 30, 40)
// 	fmt.Println(resultID, testImagePath)
// }

// func TestLoadBitmaps(t *testing.T) {
// 	loadBitmaps(subPath)
// 	if isKeyInStore(testImagePath) != true {
// 		t.Fatalf("%s did not exists", testImagePath)
// 	}
// }

// func TestOpenBitmap(t *testing.T) {
// 	openBitmap(testImagePath)
// }

func TestUuid(t *testing.T) {
	id := uuid()
	if reflect.TypeOf(id) != reflect.TypeOf("s") {
		t.Fatalf("wanted type %s got %s", reflect.TypeOf("s"), reflect.TypeOf(id))
	}
}

// inBit := openBitmap("test.png")
// color := getColor(inBit, 1, 1)
// fmt.Println(color)

// for i := 0; i < 40; i++ {
// 	start := time.Now()
// 	// Code to measure

// 	// Formatted string, such as "2h3m0.5s" or "4.503μs"
// 	color := getColor(i, 0)
// 	duration := time.Since(start)
// 	// if path != "test.png" {
// 	// 	t.Errorf("wrong path %s", path)
// 	// }
// 	fmt.Println(color)
// 	fmt.Println(duration)
// }
