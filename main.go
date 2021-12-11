package main

import (
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

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
	stars := NewStars(starCnt)

	tk := time.NewTicker(time.Second / 30)
	for t := range tk.C {
		img := stars.Refresh(t)
		if err := d.Draw(d.Bounds(), img, image.Point{}); err != nil {
			log.Fatal(err)
		}
	}
}

// getLEDs returns an *apa102.Dev, or fails back to *screen.Dev if no SPI port
// is found.
func getLEDs() display.Drawer {
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
	return d
}

// NRGBA convert color.Color to color.NRGBA
func NRGBA(c color.Color) color.NRGBA {
	r, g, b, _ := c.RGBA()
	fr, fg, fb := float64(r), float64(g), float64(b)
	return color.NRGBA{
		R: uint8(math.Round(fr * 0xff)),
		G: uint8(math.Round(fg * 0xff)),
		B: uint8(math.Round(fb * 0xff)),
		A: 0xff,
	}
}
