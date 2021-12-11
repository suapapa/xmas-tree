package main

import (
	"image"
	"image/color"
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
