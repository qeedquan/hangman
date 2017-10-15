package main

import (
	"math"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

var (
	DarkBlue   = sdl.Color{0x00, 0x00, 0x55, 0xFF}
	DarkGreen  = sdl.Color{0x44, 0x88, 0x44, 0xFF}
	ZeldaGreen = sdl.Color{65, 117, 8, 255}
	PaleRed    = sdl.Color{0xD9, 0x4C, 0x38, 0xFF}
)

func rotate(pts []sdl.Point, origin sdl.Point, rad float64) []sdl.Point {
	z := make([]sdl.Point, len(pts))
	s, c := math.Sincos(rad)
	for i, p := range pts {
		ox := float64(origin.X)
		oy := float64(origin.Y)
		z[i] = sdl.Point{
			int32(ox + (float64(p.X)-ox)*c - (float64(p.Y)-oy)*s),
			int32(oy + (float64(p.X)-ox)*s + (float64(p.Y)-oy)*c),
		}
	}
	return z
}

func translate(pts []sdl.Point, x, y int) []sdl.Point {
	z := make([]sdl.Point, len(pts))
	for i, p := range pts {
		z[i] = sdl.Point{p.X + int32(x), p.Y + int32(y)}
	}
	return z
}

func orient(pts []sdl.Point, origin sdl.Point, rad float64, x, y int) []sdl.Point {
	pts = rotate(pts, origin, rad)
	pts = translate(pts, x, y)
	return pts
}

type Grid struct {
	*Face
	Hic     *Char
	Used    map[rune]bool
	rect    sdl.Rect
	bg      sdl.Color
	hilight sdl.Color
	pad     int32
}

func newGrid(font *sdlttf.Font) *Grid {
	face := loadFace(font, sdlcolor.White)
	size := int32(face.Size) * 9 / 8
	return &Grid{
		Face:    face,
		bg:      DarkBlue,
		hilight: ZeldaGreen,
		pad:     5,
		rect: sdl.Rect{
			int32(0.02 * float64(conf.width)),
			int32(0.02 * float64(conf.height)),
			size,
			size,
		},
	}
}

func (g *Grid) Reset() {
	g.Used = make(map[rune]bool)
	g.Hic = nil
	for i := range g.Chars {
		g.Chars[i].State = 0x0
	}
}

func (g *Grid) CharAt(x, y int32) (rune, *Char) {
	r := g.rect
	n := 0
	s := int32(5)
	p := sdl.Point{x, y}
	for i := 'A'; i <= 'Z'; i++ {
		if p.In(r) {
			return i, &g.Chars[i]
		}
		r.X += r.W + g.pad + s
		if n++; n >= 8 {
			n = 0
			r.Y += r.H + g.pad + s
			r.X = g.rect.X
		}
	}
	return 0, nil
}

func (g *Grid) Draw() {
	r := g.rect
	n := 0
	s := int32(5)
	for i := 'A'; i <= 'Z'; i++ {
		renderer.SetDrawColor(g.bg)
		renderer.FillRect(&r)

		c := &g.Chars[i]
		if !g.Used[i] {
			t := c.Normal
			if c.State&0x3 != 0 {
				const enlarge = 5
				rp := sdl.Rect{
					r.X - enlarge/2,
					r.Y - enlarge/2,
					r.W + enlarge,
					r.H + enlarge,
				}
				renderer.SetDrawColor(g.hilight)
				renderer.FillRect(&rp)
				t = c.Underline
			}
			_, _, w, h, err := t.Query()
			ck(err)

			renderer.Copy(t, nil, &sdl.Rect{
				r.X + (r.W-int32(w))/2,
				r.Y + (r.H-int32(h))/2,
				int32(w),
				int32(h),
			})
		}

		r.X += r.W + g.pad + s
		if n++; n >= 8 {
			n = 0
			r.Y += r.H + g.pad + s
			r.X = g.rect.X
		}
	}
}
