package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

const (
	starCnt = 30
)

type Star struct {
	c        color.Color
	lifeTime time.Time
}

type Stars struct {
	cnt   int
	stars []Star
	img   *image.NRGBA
}

func NewStars(cnt int) *Stars {
	return &Stars{
		cnt:   cnt,
		stars: make([]Star, cnt),
		img:   image.NewNRGBA(image.Rect(0, 0, cnt, 1)),
	}
}

func (s *Stars) Refresh(t time.Time) image.Image {
	for i, b := range s.stars {
		if b.lifeTime.Before(t) {
			s.stars[i].c = &HSV{
				H: rand.Float64(),
				S: rand.Float64(),
				V: 1.0,
			}
			s.stars[i].lifeTime = time.Now().Add((time.Second + time.Duration(rand.Intn(2))*time.Second))
		}
	}

	for x := 0; x < s.cnt; x++ {
		s.img.SetNRGBA(x, 0, NRGBA(s.stars[x].c))
	}

	return s.img
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
