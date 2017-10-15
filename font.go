package main

import (
	"fmt"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

type Face struct {
	Chars []Char
	Size  int
}

type Char struct {
	Normal    *sdl.Texture
	Underline *sdl.Texture
	State     uint
}

func loadFace(font *sdlttf.Font, color sdl.Color) *Face {
	chars := make([]Char, 128)
	size := 0
	for r := rune(32); r < 128; r++ {
		font.SetStyle(sdlttf.STYLE_NORMAL)
		normal := loadChar(font, r, color)

		font.SetStyle(sdlttf.STYLE_UNDERLINE)
		underline := loadChar(font, r, color)

		_, _, nw, nh, err := normal.Query()
		ck(err)

		_, _, uw, uh, err := underline.Query()
		ck(err)

		size = max(size, nw, nh, uw, uh)
		chars[r] = Char{
			Normal:    normal,
			Underline: underline,
		}
	}

	return &Face{
		Chars: chars,
		Size:  size,
	}
}

func loadChar(font *sdlttf.Font, r rune, color sdl.Color) *sdl.Texture {
	surface, err := font.RenderGlyphBlended(r, color)
	ck(err)
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	ck(err)
	return texture
}

func loadFont(name string, size int) *sdlttf.Font {
	font, err := sdlttf.OpenFont(name, size)
	ck(err)
	return font
}

func blitText(f *Face, x, y int, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	sx := x
	for _, ch := range text {
		if ch >= 128 {
			ch = '?'
		}
		c := f.Chars[ch]
		t := c.Normal
		_, _, w, h, err := t.Query()
		ck(err)

		if ch == '\n' {
			x = sx
			y += h
			continue
		}

		renderer.Copy(t, nil, &sdl.Rect{
			int32(x), int32(y), int32(w), int32(h),
		})
		x += w
	}
}