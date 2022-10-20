package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
)

var res = 10000

func calcColor(col uint64) color.RGBA {

	rgb := color.RGBA{
		R: uint8(col >> 16),
		G: uint8((col >> 8) & 0xFF),
		B: uint8(col & 0xFF),
		A: 255,
	}
	return rgb
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

func uint8Clamper(mini float64, maxi float64, value float64) uint8 {
	return uint8(mapValue(mini, maxi, 0, 256, value))
}

func pageImage(w http.ResponseWriter, r *http.Request) {
	img := image.NewRGBA(image.Rect(0, 0, res, res))

	for y := -res / 2; y < res/2; y++ {
		for x := -res / 2; x < res/2; x++ {
			//col := calcColor(uint64(mapValue(0, float64(res*res), 0, math.Pow(2, 64), float64(x*y))))

			/*
				f(z) = z^2 + C
			*/

			col := color.RGBA{
				R: 0,
				G: uint8Clamper(float64(-res), float64(res), float64(x*y)),
				B: 0,
				A: 255,
			}

			img.Set((res/2)+x, (res/2)+y, col)
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
