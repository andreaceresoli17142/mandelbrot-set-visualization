package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type pixelRequest struct {
	wg   *sync.WaitGroup
	Img  *image.RGBA
	Res  int
	X    int
	Y    int
	Poly int
}

var (
	res       = 1000
	zoom      = 0.05
	xoffset   = 0.4
	yoffset   = 0.37
	iter      = 1000
	threadsN  = 300
	compCount = 0
	queues    = make([]chan pixelRequest, threadsN)
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

func CtoN(Z complex128, n int) complex128 {

	origZ := Z

	for i := 1; i < n; i++ {
		Z *= origZ
	}
	return Z
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

/*
	calculate a single pixel
*/
func calculate_pixel(wg *sync.WaitGroup, img *image.RGBA, res int, x int, y int, multiBrot int) {

	real_n := float64(mapValue(float64(-res/2), float64(res/2-1), -zoom+xoffset, zoom+xoffset, float64(x)))
	imaginary_n := float64(mapValue(float64(-res/2), float64(res/2-1), -zoom+yoffset, zoom+yoffset, float64(y)))

	var C = complex(real_n, imaginary_n)

	var Z complex128
	var Rho float64
	var i int

	for Rho < 4.0 && i < iter {
		Z = CtoN(Z, multiBrot) + C
		Rho = real(Z)
		i++
	}

	col := color.RGBA{
		R: 0,
		G: uint8Clamper(0, float64(iter), float64(i)),
		B: 0,
		A: 255,
	}

	img.Set((res/2)+x, (res/2)+y, col)
	(*wg).Done()
}

/*
	serve the loading image
*/
func loadingImage(w http.ResponseWriter, r *http.Request) {
	ret, _ := os.ReadFile("loading.png")
	w.Write(ret)
}

/*
	serves the image
*/
func pageImage(w http.ResponseWriter, r *http.Request) {

	polyi := r.URL.Query().Get("poly")
	rest := r.URL.Query().Get("res")
	rzoom := r.URL.Query().Get("zoom")
	rxoffs := r.URL.Query().Get("xoffs")
	ryoffs := r.URL.Query().Get("yoffs")

	var wg sync.WaitGroup

	if polyi == "" || rest == "" || rzoom == "" || rxoffs == "" || ryoffs == "" {
		return
	}

	poly, err := strconv.Atoi(polyi)
	if err != nil {
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
	var c int
	for y := -res / 2; y < res/2; y++ {
		for x := -res / 2; x < res/2; x++ {
			wg.Add(1)
			go sendChan(queues[c%threadsN], pixelRequest{
				wg:   &wg,
				Img:  img,
				Res:  res,
				X:    x,
				Y:    y,
				Poly: poly,
			})
			c++
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

	wg.Wait()

	png.Encode(w, img)

	fmt.Printf("\nset generation took %v\n", time.Since(start))
}

/*
	processor code
	waits until a requst is sent through the channel then processes the pixel
*/
func listenChan(queue chan pixelRequest) {
	for {
		req := <-queue
		calculate_pixel(req.wg, req.Img, req.Res, req.X, req.Y, req.Poly)
	}
}

// send pixel to process to a processor queue
func sendChan(channel chan pixelRequest, data pixelRequest) {
	channel <- data
}

func main() {

	// start pixel processing threads
	for i := 0; i < threadsN; i++ {
		queues[i] = make(chan pixelRequest)
		go listenChan(queues[i])
	}

	// start webserver
	http.HandleFunc("/", pageMain)
	http.HandleFunc("/mandelbrot.png", pageImage)
	http.HandleFunc("/loading.png", loadingImage)
	fmt.Println("Listening on http://localhost:3000/")
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		fmt.Printf("An error has occured serving the webpages: %v", err)
	}
}
