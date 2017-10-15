package main

import (
	"bufio"
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/qeedquan/go-media/sdl"
	"github.com/qeedquan/go-media/sdl/sdlgfx"
	"github.com/qeedquan/go-media/sdl/sdlimage/sdlcolor"
	"github.com/qeedquan/go-media/sdl/sdlttf"
)

const (
	PLAY = iota + 1
	WIN
	LOSE
)

type history struct {
	matched bool
	char    rune
}

var (
	window   *sdl.Window
	renderer *sdl.Renderer
	fps      sdlgfx.FPSManager

	conf struct {
		width      int
		height     int
		fullscreen bool
		assets     string
		cheat      bool
	}

	face struct {
		smallb *Face
		small  *Face
		normal *Face
		large  *Face
	}

	font struct {
		large  *sdlttf.Font
		big    *sdlttf.Font
		normal *sdlttf.Font
		small  *sdlttf.Font
	}

	hangman *Hangman
	grid    *Grid
	words   []string
	game    struct {
		state int
		word  string
		score uint64
		undos uint64
		hist  []history
	}
)

func main() {
	runtime.LockOSThread()
	rand.Seed(time.Now().UnixNano())
	parseFlags()
	initSDL()
	load()
	reset()
	for {
		event()
		update()
		blit()
		fps.Delay()
	}
}

func parseFlags() {
	conf.width = 800
	conf.height = 600
	conf.assets = filepath.Join(sdl.GetBasePath(), "assets")
	flag.StringVar(&conf.assets, "assets", conf.assets, "directory to assets")
	flag.BoolVar(&conf.fullscreen, "fullscreen", false, "fullscreen mode")
	flag.BoolVar(&conf.cheat, "cheat", false, "cheat")
	flag.Parse()
}

func initSDL() {
	err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_TIMER)
	ck(err)

	err = sdlttf.Init()
	ck(err)

	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "best")

	wflag := sdl.WINDOW_RESIZABLE
	if conf.fullscreen {
		wflag |= sdl.WINDOW_FULLSCREEN_DESKTOP
	}

	window, renderer, err = sdl.CreateWindowAndRenderer(conf.width, conf.height, wflag)
	ck(err)

	window.SetTitle("Hangman")
	renderer.SetLogicalSize(conf.width, conf.height)

	fps.Init()
	fps.SetRate(60)
}

func load() {
	filename := filepath.Join(conf.assets, "UbuntuTitling-Bold-webfont.ttf")
	font.large = loadFont(filename, 100)
	font.big = loadFont(filename, 48)
	font.normal = loadFont(filename, 40)
	font.small = loadFont(filename, 30)

	face.small = loadFace(font.small, sdlcolor.Black)
	face.smallb = loadFace(font.small, sdlcolor.Blue)
	face.normal = loadFace(font.normal, sdlcolor.Black)
	face.large = loadFace(font.large, sdlcolor.Black)

	hangman = newHangman()
	grid = newGrid(font.small)
	words = loadWords("words.txt")
}

func loadWords(name string) []string {
	const (
		MIN = 3
		MAX = 10
	)
	filename := filepath.Join(conf.assets, name)
	f, err := os.Open(filename)
	ck(err)
	defer f.Close()

	var w []string
	s := bufio.NewScanner(f)
loop:
	for s.Scan() {
		line := s.Text()
		if len(line) <= MIN || len(line) >= MAX {
			continue
		}
		for _, ch := range line {
			ch = unicode.ToUpper(ch)
			if !('A' <= ch && ch <= 'Z') {
				continue loop
			}
		}
		w = append(w, strings.ToUpper(line))
	}
	return w
}

func reset() {
	if game.state != WIN {
		game.score = 0
		game.undos = 5
	}
	game.word = words[rand.Intn(len(words))]
	game.hist = game.hist[:0]
	grid.Reset()
	hangman.Show = 0
	game.state = PLAY
}

func keymap(k sdl.Keycode) rune {
	switch k {
	case sdl.K_a:
		return 'A'
	case sdl.K_b:
		return 'B'
	case sdl.K_c:
		return 'C'
	case sdl.K_d:
		return 'D'
	case sdl.K_e:
		return 'E'
	case sdl.K_f:
		return 'F'
	case sdl.K_g:
		return 'G'
	case sdl.K_h:
		return 'H'
	case sdl.K_i:
		return 'I'
	case sdl.K_j:
		return 'J'
	case sdl.K_k:
		return 'K'
	case sdl.K_l:
		return 'L'
	case sdl.K_m:
		return 'M'
	case sdl.K_n:
		return 'N'
	case sdl.K_o:
		return 'O'
	case sdl.K_p:
		return 'P'
	case sdl.K_q:
		return 'Q'
	case sdl.K_r:
		return 'R'
	case sdl.K_s:
		return 'S'
	case sdl.K_t:
		return 'T'
	case sdl.K_u:
		return 'U'
	case sdl.K_v:
		return 'V'
	case sdl.K_w:
		return 'W'
	case sdl.K_x:
		return 'X'
	case sdl.K_y:
		return 'Y'
	case sdl.K_z:
		return 'Z'
	}
	return 0
}

func event() {
	for {
		ev := sdl.PollEvent()
		if ev == nil {
			break
		}

		switch ev := ev.(type) {
		case sdl.QuitEvent:
			os.Exit(0)

		case sdl.KeyDownEvent:
			switch ev.Sym {
			case sdl.K_ESCAPE:
				os.Exit(0)
			case sdl.K_SPACE:
				reset()
			case sdl.K_BACKSPACE:
				undo()
			default:
				keyDown(&ev)
			}

		case sdl.KeyUpEvent:
			guess(keymap(ev.Sym))

		case sdl.MouseMotionEvent:
			moveMouse(&ev)

		case sdl.MouseButtonUpEvent:
			pressMouse(&ev)
		}
	}
}

func keyDown(ev *sdl.KeyDownEvent) {
	c := keymap(ev.Sym)
	if c == 0 {
		return
	}

	grid.Chars[c].State |= 0x2
}

func moveMouse(ev *sdl.MouseMotionEvent) {
	if game.state != PLAY {
		return
	}
	_, ch := grid.CharAt(ev.X, ev.Y)
	if grid.Hic != nil {
		grid.Hic.State &^= 0x1
	}
	grid.Hic = ch
	if ch != nil {
		ch.State |= 0x1
	}
}

func pressMouse(ev *sdl.MouseButtonUpEvent) {
	if game.state != PLAY {
		return
	}
	switch ev.Button {
	case 1, 2:
		c, _ := grid.CharAt(ev.X, ev.Y)
		if c == 0 {
			return
		}
		guess(c)
	case 3:
		undo()
	}
}

func undo() {
	if game.state != PLAY {
		return
	}
	if len(game.hist) == 0 || game.undos == 0 {
		return
	}

	n := len(game.hist) - 1
	h := &game.hist[n]
	if h.matched && !conf.cheat {
		game.score--
	} else {
		hangman.Show >>= 1
	}
	grid.Used[h.char] = false
	grid.Chars[h.char].State &^= 0x2

	game.hist = game.hist[:n]
	if !conf.cheat {
		game.undos--
	}
}

func guess(c rune) {
	const MAXSCORE = 9999999999
	if game.state != PLAY {
		return
	}
	if !('A' <= c && c <= 'Z') || grid.Used[c] {
		return
	}

	h := history{char: c}
	grid.Used[c] = true
	if strings.IndexRune(game.word, c) >= 0 {
		if game.score < MAXSCORE {
			game.score += 1
		}
		h.matched = true
	} else if !conf.cheat && hangman.Show < 0x3f {
		hangman.Show = (hangman.Show << 1) | 1
	}
	game.hist = append(game.hist, h)
}

func blitStatus() {
	f := face.small
	x := int(0.56 * float64(conf.width))
	y := int(0.02 * float64(conf.height))
	blitText(f, x, y, "SCORE   %010d", game.score)

	y += f.Size
	blitText(f, x, y, "UNDOS  %01d", game.undos)

	switch game.state {
	case LOSE:
		blitWinLose(sdlcolor.Red, "YOU LOST!")

	case WIN:
		blitWinLose(sdlcolor.Blue, "YOU WON!")
	}
}

func blitWinLose(c sdl.Color, s string) {
	f := face.large
	x := 150
	y := 200
	renderer.SetDrawColor(c)
	renderer.FillRect(&sdl.Rect{int32(x) - 80, int32(y) - 5, 600, 200})
	blitText(f, x, y, s)

	x -= 30
	y += f.Size
	f = face.normal
	blitText(f, x, y, "PRESS SPACE TO PLAY AGAIN")
}

func blitWord() {
	const pad = 8
	x := int(0.01 * float64(conf.width))
	y := int(0.90 * float64(conf.height))
	f := face.small
	u := face.smallb
	s := f.Size
	for _, ch := range game.word {
		if grid.Used[ch] || game.state == LOSE {
			var c *Char
			if !grid.Used[ch] {
				c = &u.Chars[ch]
			} else {
				c = &f.Chars[ch]
			}
			t := c.Normal
			_, _, w, h, err := t.Query()
			ck(err)
			renderer.Copy(t, nil, &sdl.Rect{
				int32(x + (s-w)/2), int32(y), int32(w), int32(h),
			})
		}

		renderer.SetDrawColor(sdlcolor.Black)
		renderer.FillRect(&sdl.Rect{
			int32(x), int32(y) + int32(f.Size), int32(s), 4,
		})
		x += s + pad
	}
}

func update() {
	switch {
	case hangman.Show == 0x3f:
		game.state = LOSE
	case isWon():
		game.state = WIN
	}
}

func isWon() bool {
	for _, ch := range game.word {
		if !grid.Used[ch] {
			return false
		}
	}
	return true
}

func blit() {
	renderer.SetDrawColor(sdlcolor.White)
	renderer.Clear()

	hangman.Draw()
	grid.Draw()
	blitWord()
	blitStatus()

	renderer.Present()
}

func ck(err error) {
	if err != nil {
		sdl.LogCritical(sdl.LOG_CATEGORY_APPLICATION, "%v", err)
		sdl.ShowSimpleMessageBox(sdl.MESSAGEBOX_ERROR, "Error", err.Error(), window)
		os.Exit(1)
	}
}

func max(x ...int) int {
	m := x[0]
	for _, x := range x[1:] {
		if x > m {
			m = x
		}
	}
	return m
}
