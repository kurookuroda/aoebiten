package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type Game struct {
	input *Input

	fonts       map[int]font.Face
	currentSize int
	face        font.Face
	lineHeight  int

	paragraphs   []string
	pages        [][]string
	pageIndex    int
	revealed     int
	timer        int
	frameCount   int
	pageDone     bool
	skipCooldown int

	snd *SoundManager
}

func NewGame() (*Game, error) {
	g := &Game{
		input:       &Input{},
		fonts:       make(map[int]font.Face),
		currentSize: FontSizeDefault,
	}

	if err := g.loadFonts(); err != nil {
		return nil, err
	}

	g.paragraphs = g.loadTextData()
	g.face = g.fonts[g.currentSize]
	g.lineHeight = g.currentSize + 6
	g.rebuildPages()

	var err error
	g.snd, err = newSoundManager()
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Game) loadFonts() error {
	fontDataMap := map[int][]byte{
		10: font10Data,
		12: font12Data,
	}

	for size, data := range fontDataMap {
		if len(data) == 0 {
			path := FontConfig[size]
			d, err := os.ReadFile(path)
			if err != nil {
				log.Printf("warning: cannot read font %s: %v", path, err)
				g.fonts[size] = nil
				continue
			}
			data = d
		}
		ttf, err := opentype.Parse(data)
		if err != nil {
			return fmt.Errorf("failed to parse font size %d: %w", size, err)
		}
		face, err := opentype.NewFace(ttf, &opentype.FaceOptions{
			Size: float64(size),
			DPI:  72,
		})
		if err != nil {
			return fmt.Errorf("failed to create font face size %d: %w", size, err)
		}
		g.fonts[size] = face
	}
	return nil
}

func (g *Game) loadTextData() []string {
	if len(textData) > 0 {
		s := string(textData)
		for len(s) > 0 && s[len(s)-1] == '\n' {
			s = s[:len(s)-1]
		}
		return splitLines(s)
	}
	paras, err := loadParagraphs(FilePath)
	if err != nil {
		return []string{"（ファイル読み込みエラー）"}
	}
	return paras
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func (g *Game) rebuildPages() {
	g.face = g.fonts[g.currentSize]
	g.lineHeight = g.currentSize + 6
	rowsPerPage := max(1, (BoxH-Padding*2-FooterH)/g.lineHeight)

	wrapped := wrapParagraphs(g.paragraphs, g.face, MaxTextW)
	g.pages = paginate(wrapped, rowsPerPage)
}

func (g *Game) reset() {
	g.pageIndex = 0
	g.revealed = 0
	g.timer = 0
	g.pageDone = false
	g.skipCooldown = 0
}

func (g *Game) toggleFontSize() {
	sizes := make([]int, 0, len(FontConfig))
	for k := range FontConfig {
		sizes = append(sizes, k)
	}
	for i := 0; i < len(sizes); i++ {
		for j := i + 1; j < len(sizes); j++ {
			if sizes[i] > sizes[j] {
				sizes[i], sizes[j] = sizes[j], sizes[i]
			}
		}
	}
	idx := 0
	for i, s := range sizes {
		if s == g.currentSize {
			idx = i
			break
		}
	}
	newSize := sizes[(idx+1)%len(sizes)]
	g.currentSize = newSize
	g.rebuildPages()

	if g.pageIndex >= len(g.pages) {
		g.pageIndex = max(0, len(g.pages)-1)
	}
	g.revealed = 0
	g.timer = 0
	g.pageDone = false
	g.skipCooldown = 5
}

func (g *Game) currentPageText() string {
	if g.pageIndex < 0 || g.pageIndex >= len(g.pages) {
		return ""
	}
	lines := g.pages[g.pageIndex]
	s := ""
	for i, l := range lines {
		if i > 0 {
			s += "\n"
		}
		s += l
	}
	return s
}

func (g *Game) getTypingSpeed() (interval, charsPerTick int) {
	if g.input.superSpeedCombo() {
		return 0, 3
	}
	if g.input.speedCombo() {
		return 0, 1
	}
	if g.input.spaceHeld() {
		return FastInterval, 1
	}
	return CharInterval, 1
}

func (g *Game) Update() error {
	g.frameCount++
	g.input.update()

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	if g.input.resetPressed() {
		g.reset()
		return nil
	}

	if g.input.fontTogglePressed() {
		g.toggleFontSize()
		return nil
	}

	if g.input.backPressed() {
		if g.pageIndex >= len(g.pages) {
			if len(g.pages) > 0 {
				g.pageIndex = len(g.pages) - 1
				g.revealed = len(g.currentPageText())
				g.pageDone = true
				g.skipCooldown = 5
			}
			return nil
		}
		if g.pageIndex > 0 {
			g.pageIndex--
			g.revealed = len(g.currentPageText())
			g.pageDone = true
			g.skipCooldown = 5
		}
		return nil
	}

	if g.skipCooldown > 0 {
		g.skipCooldown--
		return nil
	}

	if g.pageIndex >= len(g.pages) {
		return nil
	}

	txt := g.currentPageText()

	if !g.pageDone {
		if g.input.skipPressed() {
			g.revealed = len(txt)
			g.pageDone = true
			g.skipCooldown = 8
			return nil
		}

		interval, charsPerTick := g.getTypingSpeed()
		g.timer++
		if g.timer > interval {
			g.timer = 0
			advanced := false

			for c := 0; c < charsPerTick; c++ {
				for g.revealed < len(txt) && txt[g.revealed] == '\n' {
					g.revealed++
				}
				if g.revealed < len(txt) {
					g.revealed++
					advanced = true
				}
			}

			if advanced {
				g.snd.playTalk(interval, charsPerTick)
			}

			if g.revealed >= len(txt) {
				g.revealed = len(txt)
				g.pageDone = true
				g.skipCooldown = 5
			}
		}
	} else {
		if g.input.nextPressed() {
			g.pageIndex++
			g.revealed = 0
			g.timer = 0
			g.pageDone = false
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Pyxel cls(0) - 黒背景
	screen.Fill(PyxelBlack)

	// Pyxel rect(BOX_X+1, BOX_Y+1, BOX_W-2, BOX_H-2, 1) - 内側暗青
	vector.DrawFilledRect(screen, float32(BoxX+1), float32(BoxY+1), float32(BoxW-2), float32(BoxH-2), PyxelDarkBlue, false)

	// Pyxel rectb(BOX_X, BOX_Y, BOX_W, BOX_H, 7) - 枠線白
	vector.StrokeRect(screen, float32(BoxX), float32(BoxY), float32(BoxW), float32(BoxH), 1, PyxelWhite, false)

	if g.pageIndex >= len(g.pages) {
		y := BoxY + Padding
		if g.face != nil {
			m := g.face.Metrics()
			y += m.Ascent.Ceil()
		}
		text.Draw(screen, "-- 読了 --", g.face, BoxX+Padding, y, PyxelWhite)
		g.drawUI(screen)
		return
	}

	txt := g.currentPageText()
	if g.revealed > len(txt) {
		g.revealed = len(txt)
	}
	shown := txt[:g.revealed]
	lines := splitLines(shown)

	baseY := BoxY + Padding
	if g.face != nil {
		m := g.face.Metrics()
		baseY += m.Ascent.Ceil()
	}

	for i, line := range lines {
		y := baseY + i*g.lineHeight
		text.Draw(screen, line, g.face, BoxX+Padding, y, PyxelWhite)
	}

	// Pyxel ▼ 点滅 (frame_count % 30 < 15)
	if g.pageDone && (g.frameCount%30) < 15 {
		y := BoxY + BoxH - 12
		if g.face != nil {
			m := g.face.Metrics()
			y = BoxY + BoxH - Padding - g.currentSize + m.Ascent.Ceil()
		}
		text.Draw(screen, "▼", g.face, BoxX+BoxW-14, y, PyxelWhite)
	}

	g.drawUI(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	if len(g.pages) == 0 {
		return
	}

	total := len(g.pages)
	current := g.pageIndex + 1
	if current > total {
		current = total
	}
	pageLabel := fmt.Sprintf("%d/%d", current, total)
	fontLabel := fmt.Sprintf("%dpx", g.currentSize)

	footerY := BoxY + BoxH - Padding - g.currentSize
	if g.face != nil {
		m := g.face.Metrics()
		footerY = BoxY + BoxH - Padding - g.currentSize + m.Ascent.Ceil()
	}

	labelW := len(pageLabel) * g.currentSize
	if g.face != nil {
		labelW = font.MeasureString(g.face, pageLabel).Ceil()
	}
	x := BoxX + BoxW - Padding - labelW

	// Pyxel color 5 = dark_gray
	text.Draw(screen, pageLabel, g.face, x, footerY, PyxelDarkGray)
	text.Draw(screen, fontLabel, g.face, BoxX+Padding, footerY, PyxelDarkGray)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth*2, ScreenHeight*2)
	ebiten.SetWindowTitle("Aozora Reader")

	game, err := NewGame()
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
