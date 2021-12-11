package main

import (
	"image"
	"image/draw"
	"image/gif"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/nfnt/resize"
	"github.com/sztanpet/sh1106"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/devices/apa102"
	"periph.io/x/periph/host"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	go startDisp()
	go startLEDs()

	c := make(chan struct{})
	<-c
}

// startLEDs returns an *apa102.Dev, or fails back to *screen.Dev if no SPI port
// is found.
func startLEDs() {
	s, err := spireg.Open("")
	if err != nil {
		panic(err)
	}
	// Change the option values to see their effects.
	var dispOpts = &apa102.Opts{
		NumPixels:        starCnt, // 150 LEDs is a common strip length.
		Intensity:        128,     // Full blinding power.
		Temperature:      5000,    // More pleasing white balance than NeutralTemp.
		DisableGlobalPWM: false,   // Use full 13 bits range.
	}
	d, err := apa102.New(s, dispOpts)
	if err != nil {
		log.Fatal(err)
	}

	stars := NewStars(starCnt)

	tk := time.NewTicker(time.Second / 30)
	for t := range tk.C {
		img := stars.Refresh(t)
		if err := d.Draw(d.Bounds(), img, image.Point{}); err != nil {
			log.Fatal(err)
		}
	}
}

func startDisp() {
	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open("")
	if err != nil {
		log.Fatal(err)
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	dev, err := sh1106.NewI2C(bus, &sh1106.DefaultOpts)
	if err != nil {
		log.Fatal(err)
	}

	// Decodes an animated GIF as specified on the command line:
	if len(os.Args) != 2 {
		log.Fatal("please provide the path to an animated GIF")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Converts every frame to image.Gray and resize them:
	imgs := make([]*image.Gray, len(g.Image))
	for i := range g.Image {
		imgs[i] = convertAndResizeAndCenter(dev.Bounds().Dx(), dev.Bounds().Dy(), g.Image[i])
	}

	// Display the frames in a loop:
	var i int
	for {
		index := i % len(imgs)
		c := time.After(time.Duration(10*g.Delay[index]) * time.Millisecond)
		img := imgs[index]
		dev.Draw(img.Bounds(), img, image.Point{})
		<-c
		i++
	}
}

// convertAndResizeAndCenter takes an image, resizes and centers it on a
// image.Gray of size w*h.
func convertAndResizeAndCenter(w, h int, src image.Image) *image.Gray {
	src = resize.Thumbnail(uint(w), uint(h), src, resize.Bicubic)
	img := image.NewGray(image.Rect(0, 0, w, h))
	r := src.Bounds()
	r = r.Add(image.Point{(w - r.Max.X) / 2, (h - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}
