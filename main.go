package main

import (
	"flag"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/nfnt/resize"
	"github.com/pkg/errors"
	"github.com/sztanpet/sh1106"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/spi/spireg"
	"periph.io/x/devices/v3/apa102"
	"periph.io/x/host/v3"
)

var (
	flagDisp string
	flagLEDs bool
	flagV    bool

	errC chan error
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.StringVar(&flagDisp, "disp", "", "run gif on display")
	flag.BoolVar(&flagLEDs, "leds", false, "run leds")
	flag.BoolVar(&flagV, "v", false, "decrease v smoothly")
	flag.Parse()

	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

	errC = make(chan error, 2)
	go func() {
		for err := range errC {
			log.Printf("error: %v", err)
		}
	}()

	if flagDisp != "" {
		go runDisp(flagDisp)
	}
	if flagLEDs {
		go runLEDs()
	}

	c := make(chan struct{})
	<-c
}

// runLEDs returns an *apa102.Dev, or fails back to *screen.Dev if no SPI port
// is found.
func runLEDs() {
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

	sky := NewSky(starCnt)

	tk := time.NewTicker(time.Second / 60)
	defer tk.Stop()
	for t := range tk.C {
		img := sky.Refresh(t)
		if err := d.Draw(d.Bounds(), img, image.Point{}); err != nil {
			errC <- errors.Wrap(err, "failed to draw leds")
			return
		}
	}
}

func runDisp(gifFN string) {
	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open("")
	if err != nil {
		errC <- errors.Wrap(err, "failed to open i2c bus")
		return
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	dev, err := sh1106.NewI2C(bus, &sh1106.DefaultOpts)
	if err != nil {
		errC <- errors.Wrap(err, "failed to open display")
		return
	}

	// Decodes an animated GIF as specified on the command line:
	f, err := os.Open(gifFN)
	if err != nil {
		errC <- errors.Wrap(err, "failed to open gif")
		return
	}
	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		errC <- errors.Wrap(err, "failed to decode gif")
		return
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
		if err := dev.Draw(img.Bounds(), img, image.Point{}); err != nil {
			errC <- errors.Wrap(err, "failed to draw image")
			return
		}
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
