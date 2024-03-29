package main

import (
	"context"
	"flag"
	"image"
	"image/draw"
	"image/gif"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	errC   chan error
	exitWG sync.WaitGroup
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

	ctx, cancelF := context.WithCancel(context.Background())
	if flagDisp != "" {
		exitWG.Add(1)
		go runDisp(ctx, flagDisp)
	}
	if flagLEDs {
		exitWG.Add(1)
		go runLEDs(ctx)
	}

	exitC := make(chan os.Signal, 1)
	signal.Notify(exitC, syscall.SIGINT, syscall.SIGTERM)

	<-exitC
	cancelF()
	exitWG.Wait()
}

// runLEDs runs loop that changes the LEDs colors.
func runLEDs(ctx context.Context) {
	defer exitWG.Done()
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

	sky := NewSky(ctx, starCnt)
	sky.Start()

	tk := time.NewTicker(time.Second / 24) //24 fps
	defer tk.Stop()

	updateF := func() {
		img := sky.Image()
		if err := d.Draw(d.Bounds(), img, image.Point{}); err != nil {
			errC <- errors.Wrap(err, "failed to draw leds")
			return
		}
	}

	for {
		select {
		case <-ctx.Done():
			// 별을 모두 끄기 위해
			time.Sleep(time.Second)
			updateF()
			return
		case <-tk.C:
			updateF()
		}
	}
}

func runDisp(ctx context.Context, gifFN string) {
	defer exitWG.Done()
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
		select {
		case <-ctx.Done():
			return
		default:
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
