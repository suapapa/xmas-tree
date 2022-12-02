package main

import (
	"context"
	"image"
	"image/color"
	"math/rand"
	"sync"
	"time"
)

const (
	starCnt = 30
)

type Star struct {
	ctx context.Context
	c   *HSV
	d   time.Duration
	sync.RWMutex
}

func NewStar(ctx context.Context) *Star {
	return &Star{
		ctx: ctx,
		c:   &HSV{},
	}
}

func (s *Star) Start() {
	go func() {
		defer func() {
			// 종료시 모두 끔
			s.Lock()
			s.c.H = 0
			s.c.S = 0
			s.c.V = 0
			s.Unlock()
		}()

		var dur time.Duration
		// 처음 시작시 한번에 다 켜지는 것을 방지하기 위해
		time.Sleep(time.Duration(rand.Intn(5000)) * time.Millisecond)
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.Lock()
				if flagV {
					s.c.V -= 0.01 // 밝기를 점점 줄임
				} else {
					s.c.V = 0
				}
				if s.c.V <= 0 || dur <= 0 {
					s.c.H = rand.Float64()
					s.c.S = 0.5 + 0.5*rand.Float64() // 채도가 너무 낮은 색은 지양
					s.c.V = 1.0
					// 3초 ~ 10초
					s.d = 3*time.Second + time.Duration(time.Duration(rand.Intn(7_000))*time.Millisecond)
					dur = s.d / 100
				}
				s.Unlock()
				if flagV {
					time.Sleep(dur)
				} else {
					time.Sleep(s.d)
				}
			}
		}
	}()
}

func (s *Star) GetNRGBA() color.NRGBA {
	s.Lock()
	defer s.Unlock()

	r, g, b, _ := s.c.RGBA()
	c := color.NRGBA{
		R: uint8(r >> 24),
		G: uint8(g >> 24),
		B: uint8(b >> 24),
		A: 0xff,
	}

	return c
}

type Sky struct {
	ctx   context.Context
	cnt   int
	stars []*Star
	img   *image.NRGBA
}

func NewSky(ctx context.Context, cnt int) *Sky {
	sky := Sky{
		ctx:   ctx,
		cnt:   cnt,
		stars: make([]*Star, cnt),
		img:   image.NewNRGBA(image.Rect(0, 0, cnt, 1)),
	}

	return &sky
}

func (s *Sky) Start() {
	for i := range s.stars {
		s.stars[i] = NewStar(s.ctx)
		s.stars[i].Start()
	}
}

func (s *Sky) Image() image.Image {
	for i, star := range s.stars {
		s.img.SetNRGBA(i, 0, star.GetNRGBA())
	}

	return s.img
}
