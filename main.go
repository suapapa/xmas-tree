package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"periph.io/x/extra/devices/screen"
	"periph.io/x/periph/conn/display"
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
	d := getLEDs()
	img := image.NewNRGBA(d.Bounds())
	log.Println(img.Rect)

	var sX = 0
	tk := time.NewTicker(time.Second / 1)
	for range tk.C {
		for x := 0; x < img.Rect.Max.X; x++ {
			hsv := HSV{
				H: rand.Float64(),
				S: rand.Float64(),
				V: 1.0,
			}
			img.SetNRGBA(x, 0, hsv.NRGBA())
		}
		if err := d.Draw(d.Bounds(), img, image.Point{}); err != nil {
			log.Fatal(err)
		}
		sX++
		if sX >= 30 {
			sX = 0
		}
	}
}

// getLEDs returns an *apa102.Dev, or fails back to *screen.Dev if no SPI port
// is found.
func getLEDs() display.Drawer {
	s, err := spireg.Open("")
	if err != nil {
		fmt.Printf("Failed to find a SPI port, printing at the console:\n")
		return screen.New(30)
	}
	// Change the option values to see their effects.
	var dispOpts = &apa102.Opts{
		NumPixels:        30,    // 150 LEDs is a common strip length.
		Intensity:        128,   // Full blinding power.
		Temperature:      5000,  // More pleasing white balance than NeutralTemp.
		DisableGlobalPWM: false, // Use full 13 bits range.
	}
	d, err := apa102.New(s, dispOpts)
	if err != nil {
		log.Fatal(err)
	}
	return d
}