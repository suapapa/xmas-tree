package main

import (
	"image"
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	starCnt = 30
)

type Star struct {
	i int
	c *HSV
	d time.Duration
	sync.RWMutex
}

func (s *Star) Run() {
	dieF := func(dur time.Duration) {
		for {
			s.Lock()
			if flagV {
				s.c.V -= 0.01
			} else {
				s.c.V = 0
			}
			if s.c.V <= 0 || dur <= 0 {
				s.c = &HSV{
					H: rand.Float64(),
					S: 1.0,
					V: 1.0,
				}
				s.d = 10*time.Second + time.Duration(time.Duration(rand.Intn(10_000))*time.Millisecond)
				dur = s.d / 100
				if s.i == 0 {
					log.Printf("s.d: %v, dur: %v", s.d, dur)
				}
			}
			s.Unlock()
			if flagV {
				time.Sleep(dur)
			} else {
				time.Sleep(s.d)
			}
		}
	}

	s.c = &HSV{}
	// 10초 ~ 20초 사이
	s.d = 10*time.Second + time.Duration(time.Duration(rand.Intn(10_000))*time.Millisecond)
	go dieF(0)
}

func (s *Star) GetNRGBA() color.NRGBA {
	s.Lock()
	defer s.Unlock()

	r, g, b, _ := s.c.RGBA()
	if s.i == 0 {
		log.Printf("r: %08x, g: %08x, b: %08x", r, g, b)
	}
	c := color.NRGBA{
		R: uint8(r >> 24),
		G: uint8(g >> 24),
		B: uint8(b >> 24),
		A: 0xff,
	}

	if s.i == 0 {
		log.Printf("c: %v", c)
	}

	return c
}

type Sky struct {
	cnt   int
	stars []*Star
	img   *image.NRGBA
}

func NewSky(cnt int) *Sky {
	sky := Sky{
		cnt:   cnt,
		stars: make([]*Star, cnt),
		img:   image.NewNRGBA(image.Rect(0, 0, cnt, 1)),
	}

	for i := range sky.stars {
		sky.stars[i] = &Star{
			i: i,
		}
		go sky.stars[i].Run()
	}

	return &sky
}

func (s *Sky) Refresh(t time.Time) image.Image {
	for i, star := range s.stars {
		s.img.SetNRGBA(i, 0, star.GetNRGBA())
	}

	return s.img
}
