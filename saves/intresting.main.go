package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
)

var resolution = 1000

const N32B = 8 * 8 * 8 * 8
const RGBA32 = 8 * 8 * 8

func calcColor(col uint32) color.RGBA {

	mapped := uint(mapValue(0, N32B, 0, RGBA32, float64(col)))

	ret := color.RGBA{
		R: uint8((mapped >> 8) & (1<<0 - 1)),
		G: uint8((mapped >> 8) & (1<<8 - 1)),
		B: uint8((mapped >> 8) & (1<<16 - 1)),
		A: 255,
	}
	return ret
}

func mapValue(imin float64, imax float64, omin float64, omax float64, value float64) float64 {
	//fmt.Printf("(%v - %v) / (%v - %v)\n", value, imin, imax, imin)
	x := (value - imin) / (imax - imin)

	return x*(omax-omin) + omin
}

func pageMain(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
			<!doctype html>
			<title>madelbrot set</title>
			<h1>madelbrot set</h1>
			<img src="/madelbrot.png">
			`))
}

func pageImage(w http.ResponseWriter, r *http.Request) {
	img := image.NewRGBA(image.Rect(0, 0, resolution, resolution))

	for y := 0; y < resolution; y++ {
		for x := 0; x < resolution; x++ {
			//fmt.Println(x*y, uint32(mapValue(0, float64(y*x), 0, math.Pow(2, 32), float64(x*y))))
			img.Set(x, y, calcColor(uint32(mapValue(0, float64(resolution*resolution), 0, math.Pow(2, 31), float64(x*y)))))
		}
	}

	png.Encode(w, img)
}

func main() {
	http.HandleFunc("/", pageMain)
	http.HandleFunc("/madelbrot.png", pageImage)
	fmt.Println("Listening on http://localhost:3000/")
	http.ListenAndServe(":3000", nil)
}
