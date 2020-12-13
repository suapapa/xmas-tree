package main

import (
	"math"
)

// HSV represents HSV color space
type HSV struct {
	H, S, V float64
}

// RGBA implements color.Color interface
func (c *HSV) RGBA() (r, g, b, a uint32) {
	fr, fg, fb := c.rgb()
	r = uint32(math.Round(fr * 0xffffffff))
	g = uint32(math.Round(fg * 0xffffffff))
	b = uint32(math.Round(fb * 0xffffffff))
	a = 0xffffffff
	return
}

func (c *HSV) rgb() (r, g, b float64) {
	if c.S == 0 { //HSV from 0 to 1
		r = c.V * 255
		g = c.V * 255
		b = c.V * 255
	} else {
		h := c.H * 6
		if h == 6 {
			h = 0
		} //H must be < 1
		i := math.Floor(h) //Or ... var_i = floor( var_h )
		v1 := c.V * (1 - c.S)
		v2 := c.V * (1 - c.S*(h-i))
		v3 := c.V * (1 - c.S*(1-(h-i)))

		if i == 0 {
			r = c.V
			g = v3
			b = v1
		} else if i == 1 {
			r = v2
			g = c.V
			b = v1
		} else if i == 2 {
			r = v1
			g = c.V
			b = v3
		} else if i == 3 {
			r = v1
			g = v2
			b = c.V
		} else if i == 4 {
			r = v3
			g = v1
			b = c.V
		} else {
			r = c.V
			g = v1
			b = v2
		}

		// r = r * 255 //RGB results from 0 to 255
		// g = g * 255
		// b = b * 255
	}
	return r, g, b
}
