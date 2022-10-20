package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
)

var res = 1000

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

			Ca := float64(mapValue(float64(-res/2), float64(res/2-1), -2.0, 2.0, float64(x)))
			Cb := float64(mapValue(float64(-res/2), float64(res/2-1), -2.0, 2.0, float64(y)))

			var Za, Zb float64

			var i int

			for Za*Za+Zb*Zb < 4 && i < 1000 {
				Za = (Za * Za) - (Zb * Zb) + Ca
				Zb = (2 * Za * Zb) + Cb
				i++
			}

			col := color.RGBA{
				R: 0,
				G: uint8Clamper(0, math.Sqrt(1000.0), math.Sqrt(float64(i))),
				B: 0,
				A: 255,
			}

			img.Set((res/2)+x, (res/2)+y, col)
		}
		//fmt.Println(y)
	}

	png.Encode(w, img)
}

func main() {
	http.HandleFunc("/", pageMain)
	http.HandleFunc("/madelbrot.png", pageImage)
	fmt.Println("Listening on http://localhost:3000/")
	http.ListenAndServe(":3000", nil)
}
