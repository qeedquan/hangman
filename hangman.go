package main

import (
	"math"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlgfx"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
)

type Hangman struct {
	blocks []Block
	head   Head
	joints []Joint
	Show   uint
}

type Block struct {
	sdl.Rect
	sdl.Color
}

type Head struct {
	X, Y, R1, R2 int
}

type Joint []sdl.Point

func newHangman() *Hangman {
	blocks := []Block{
		{Color: DarkGreen},
		{Color: PaleRed},
		{Color: PaleRed},
		{Color: sdlcolor.Black},
	}

	const m1, m2 = 0.45, 0.03
	w := float64(conf.width)
	h := float64(conf.height)

	blocks[0].Rect = sdl.Rect{
		int32(0.55 * w),
		int32(0.9 * h),
		int32(m1 * w),
		int32(m2 * h),
	}
	blocks[1].Rect = sdl.Rect{
		blocks[0].X + 5,
		blocks[0].Y - blocks[0].W*9/8,
		blocks[0].H,
		blocks[0].W * 9 / 8,
	}
	blocks[2].Rect = sdl.Rect{
		blocks[0].X + 5,
		blocks[0].Y - blocks[1].H,
		blocks[0].W * 4 / 6,
		blocks[0].H,
	}
	blocks[3].Rect = sdl.Rect{
		blocks[2].X + blocks[2].W - 50 + int32(m2*h),
		blocks[1].Y + blocks[2].H,
		10,
		30,
	}

	head := Head{
		int(blocks[3].X + blocks[3].W/2),
		int(blocks[3].Y + blocks[3].H + 40),
		30,
		40,
	}

	jw, jh := int32(200), int32(20)
	jp1 := []sdl.Point{
		{0, 0},
		{jw, 0},
		{jw, jh},
		{0, jh},
	}
	jq1 := sdl.Point{jw / 2, jh / 2}

	jw, jh = 100, 20
	jp2 := []sdl.Point{
		{0, 0},
		{jw, 0},
		{jw, jh},
		{0, jh},
	}
	jq2 := sdl.Point{jw / 2, jh / 2}

	joints := []Joint{
		orient(jp1, jq1, math.Pi/2, head.X-100, head.Y+120),
		orient(jp2, jq2, math.Pi/4, head.X-90, head.Y+60),
		orient(jp2, jq2, -math.Pi/4, head.X-10, head.Y+60),
		orient(jp2, jq2, -5*math.Pi/4, head.X-90, head.Y+250),
		orient(jp2, jq2, -7*math.Pi/4, head.X-10, head.Y+250),
	}

	return &Hangman{
		blocks: blocks,
		head:   head,
		joints: joints,
	}
}

func (h *Hangman) Draw() {
	for _, b := range h.blocks {
		renderer.SetDrawColor(b.Color)
		renderer.FillRect(&b.Rect)
	}
	if h.Show&0x1 != 0 {
		p := &h.head
		sdlgfx.FilledCircle(renderer, p.X, p.Y, p.R2, sdlcolor.Black)
		sdlgfx.FilledCircle(renderer, p.X, p.Y, p.R1, sdlcolor.White)
	}

	v := uint(0x2)
	for _, j := range h.joints {
		if h.Show&v != 0 {
			sdlgfx.FilledPolygon(renderer, j, sdlcolor.Black)
		}
		v <<= 1
	}
}
