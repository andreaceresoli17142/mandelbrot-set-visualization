package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	res     = 1000
	zoom    = 0.05
	xoffset = 0.4
	yoffset = 0.37
	iter    = 1000
)

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
	file, err := os.ReadFile("index.html")
	if err != nil {
		fmt.Printf("Error reading the files: %v", err)
	}
	w.Write(file)
}

func uint8Clamper(mini float64, maxi float64, value float64) uint8 {
	return uint8(mapValue(mini, maxi, 0, 256, value))
}

func calculate_pixel(img *image.RGBA, res int, x int, y int) {

	//fmt.Println("starting to calculate: ", x, y)

	Ca := float64(mapValue(float64(-res/2), float64(res/2-1), -zoom+xoffset, zoom+xoffset, float64(x)))
	Cb := float64(mapValue(float64(-res/2), float64(res/2-1), -zoom+yoffset, zoom+yoffset, float64(y)))

	var Za, Zb float64

	var i int

	for Za*Za+Zb*Zb < 4 && i < iter {
		Za, Zb = (Za*Za)-(Zb*Zb)+Ca, (2*Za*Zb)+Cb
		i++
	}

	col := color.RGBA{
		R: 0,
		G: uint8Clamper(0, float64(iter), float64(i)),
		B: 0,
		A: 255,
	}

	img.Set((res/2)+x, (res/2)+y, col)
	//fmt.Println("finished to calculate: ", x, y)
}

func loadingImage(w http.ResponseWriter, r *http.Request) {
	ret, _ := os.ReadFile("loading.png")
	w.Write(ret)
}

func pageImage(w http.ResponseWriter, r *http.Request) {

	rest := r.URL.Query().Get("res")
	rzoom := r.URL.Query().Get("zoom")
	rxoffs := r.URL.Query().Get("xoffs")
	ryoffs := r.URL.Query().Get("yoffs")

	if rest == "" || rzoom == "" || rxoffs == "" || ryoffs == "" {
		return
	}

	res, err := strconv.Atoi(rest)
	if err != nil {
		return
	}

	zoom, err = strconv.ParseFloat(rzoom, 64)
	if err != nil {
		return
	}

	xoffset, err = strconv.ParseFloat(rxoffs, 64)
	if err != nil {
		return
	}

	yoffset, err = strconv.ParseFloat(ryoffs, 64)
	if err != nil {
		return
	}

	img := image.NewRGBA(image.Rect(0, 0, res, res))

	fmt.Println("starting to generate image")

	fmt.Println("ofx: ", -zoom+xoffset, zoom+xoffset)
	fmt.Println("ofx: ", -zoom+yoffset, zoom+yoffset)

	start := time.Now()

	//iter = int(mapValue(1, 0, 0, 1000, zoom))

	for y := -res / 2; y < res/2; y++ {
		for x := -res / 2; x < res/2; x++ {
			//col := calcColor(uint64(mapValue(0, float64(res*res), 0, math.Pow(2, 64), float64(x*y))))

			/*
				f(z) = z^2 + C
			*/
			calculate_pixel(img, res, x, y)
		}

		// print progress bar
		if y%100 == 0 {
			perc := -(-res/2 - y) / (res / 100)

			percBar := ""

			for i := 0; i < 15; i++ {
				if int((float64(perc)/100.0)*15.0) > i {
					percBar += "="
				} else {
					percBar += " "
				}
			}

			fmt.Printf("\r [%v]  %v%v ", percBar, perc, "%")
		}

	}

	fmt.Print("\r [===============]  100%")

	png.Encode(w, img)

	fmt.Printf("\nset generation took %v\n", time.Since(start))
}

func main() {
	http.HandleFunc("/", pageMain)
	http.HandleFunc("/mandelbrot.png", pageImage)
	http.HandleFunc("/loading.png", loadingImage)
	fmt.Println("Listening on http://localhost:3000/")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Printf("An error has occured serving the webpages: %v", err)
	}
}
