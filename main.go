// Package main - go-tui crazy demo
package main

import (
	"fmt"
	"math"
	"strings"
	"time"

	tui "github.com/grindlemire/go-tui"
)

func main() {
	app, err := tui.NewApp(tui.WithRootComponent(newCrazyApp()))
	if err != nil {
		panic(err)
	}
	defer app.Close()
	if err := app.Run(); err != nil {
		panic(err)
	}
}

// ===== HELPERS =====

var spinnerDots = []string{"⠋", "⠙", "⠚", "⠞", "⠖", "⠦", "⠴", "⠲", "⠳", "⠓"}
var spinnerLine = []string{"┤", "┘", "┴", "└", "├", "┌", "┬", "┐"}
var spinnerCircle = []string{"◜", "◠", "◝", "◞", "◡", "◟"}
var spinnerBraille = []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
var barBlocks = []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉"}
var symbolPool = []string{
	"─", "━", "│", "┃", "┌", "┐", "└", "┘", "├", "┬", "┴", "┼",
	"╭", "╮", "╰", "╯", "╱", "╲", "╳",
	"←", "↑", "→", "↓", "↔", "↕", "↖", "↗", "↘", "↙", "➜", "➤",
	"∑", "∫", "√", "∞", "≈", "≠", "±", "≤", "≥", "×", "÷", "∂", "∆", "∏",
	"α", "β", "γ", "δ", "ε", "ζ", "η", "θ", "λ", "μ", "π", "τ", "ω",
	"★", "☆", "◆", "◇", "●", "○", "◐", "◑", "◒", "◓", "☀", "☁", "☂", "☃",
	"♠", "♡", "♢", "♣", "♤", "♥", "♦", "♧", "♪", "♫", "♬", "♩",
	"☕", "⚓", "⚡", "⚙", "⚛", "⚠", "♻", "♾",
	"❶", "❷", "❸", "❹", "❺", "⏳", "⌛",
	"⣀", "⣤", "⣶", "⣿", "⡿", "⢿", "⣟", "⣯", "⣷",
	"█", "▀", "▄", "▌", "▐", "░", "▒", "▓",
	"🌀", "🌊", "🔥", "⭐", "✨", "💫", "⚡",
}

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	if s == 0 { v := uint8(l * 255); return v, v, v }
	h = math.Mod(h, 360)
	if h < 0 { h += 360 }
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60: r, g, b = c, x, 0
	case h < 120: r, g, b = x, c, 0
	case h < 180: r, g, b = 0, c, x
	case h < 240: r, g, b = 0, x, c
	case h < 300: r, g, b = x, 0, c
	default: r, g, b = c, 0, x
	}
	return uint8((r+m)*255), uint8((g+m)*255), uint8((b+m)*255)
}

func hslStyle(h, s, l float64) tui.Style {
	r, g, b := hslToRGB(h, s, l)
	return tui.NewStyle().Bold().Foreground(tui.RGBColor(r, g, b))
}

func easeInOutCubic(t float64) float64 {
	if t < 0.5 { return 4 * t * t * t }
	return 1 - math.Pow(-2*t+2, 3)/2
}

func renderBar(v float64, w int) string {
	if v < 0 { v = 0 }
	if v > 1 { v = 1 }
	steps := int(v * float64(w) * 8)
	fb := steps / 8
	rem := steps % 8
	var b strings.Builder
	if fb > w { fb = w }
	b.WriteString(strings.Repeat("█", fb))
	if fb < w {
		b.WriteString(barBlocks[rem])
		b.WriteString(strings.Repeat("░", w-fb-1))
	}
	return b.String()
}

func tickerText(phase float64) string {
	sp := int(phase * 20)
	text := "GO-TUI IS AWESOME! 🚀 THE FUTURE OF TERMINAL UI IS HERE! ✨ "
	if len(text) == 0 { return "" }
	return text[sp%len(text):] + text[:sp%len(text)]
}

func flex(dir tui.Direction, opts ...tui.Option) *tui.Element {
	all := make([]tui.Option, 0, len(opts)+2)
	all = append(all, tui.WithDisplay(tui.DisplayFlex), tui.WithDirection(dir))
	all = append(all, opts...)
	return tui.New(all...)
}

func textEl(s string, st tui.Style) *tui.Element {
	return tui.New(tui.WithText(s), tui.WithTextStyle(st))
}

// ===== THEME =====

var themeColors = map[string]struct{ fg, accent tui.Color }{
	"cyber":  {tui.Cyan, tui.Magenta},
	"ocean":  {tui.Blue, tui.Green},
	"forest": {tui.Green, tui.Yellow},
	"sunset": {tui.Yellow, tui.Red},
}
var themeList = []string{"cyber", "ocean", "forest", "sunset"}

// ===== APP STATE =====

type crazyApp struct {
	spinFrame     *tui.State[int]
	wavePhase     *tui.State[float64]
	scrollWave    *tui.State[float64]
	bouncePhase   *tui.State[float64]
	borderHue     *tui.State[float64]
	startTime     time.Time
	frame         int
	counter       *tui.State[int]
	theme         *tui.State[string]
	partyMode     *tui.State[bool]
	showModal     *tui.State[bool]
	visible       *tui.State[map[string]bool]

	ctrUp *tui.Ref; ctrDown *tui.Ref; ctrReset *tui.Ref
	modalOpen *tui.Ref; modalYes *tui.Ref; modalNo *tui.Ref
	matrixPhase [30]float64
	sectionRefs map[string]*tui.Ref
	themeRefs  map[string]*tui.Ref
}

func newCrazyApp() *crazyApp {
	def := map[string]bool{"matrix": true}
	return &crazyApp{
		spinFrame: tui.NewState(0), wavePhase: tui.NewState(0.0),
		scrollWave: tui.NewState(0.0), bouncePhase: tui.NewState(0.0),
		borderHue: tui.NewState(0.0), counter: tui.NewState(0),
		theme: tui.NewState("cyber"), partyMode: tui.NewState(false),
		showModal: tui.NewState(false), visible: tui.NewState(def),
		startTime: time.Now(),
		ctrUp: tui.NewRef(), ctrDown: tui.NewRef(), ctrReset: tui.NewRef(),
		modalOpen: tui.NewRef(), modalYes: tui.NewRef(), modalNo: tui.NewRef(),
		sectionRefs: map[string]*tui.Ref{
			"bounce": tui.NewRef(), "progress": tui.NewRef(),
			"metrics": tui.NewRef(), "fireworks": tui.NewRef(),
			"map": tui.NewRef(), "symbols": tui.NewRef(), "matrix": tui.NewRef(),
		},
		themeRefs: map[string]*tui.Ref{
			"cyber": tui.NewRef(), "ocean": tui.NewRef(),
			"forest": tui.NewRef(), "sunset": tui.NewRef(),
		},
	}
}

func (a *crazyApp) BindApp(app *tui.App) {
	a.spinFrame.BindApp(app); a.wavePhase.BindApp(app)
	a.scrollWave.BindApp(app); a.bouncePhase.BindApp(app)
	a.borderHue.BindApp(app); a.counter.BindApp(app)
	a.theme.BindApp(app); a.partyMode.BindApp(app)
	a.showModal.BindApp(app); a.visible.BindApp(app)
}

func (a *crazyApp) KeyMap() tui.KeyMap {
	return tui.KeyMap{
		tui.OnStop(tui.KeyEscape, func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.OnStop(tui.Rune('q'), func(ke tui.KeyEvent) { ke.App().Stop() }),
		tui.On(tui.Rune(' '), func(ke tui.KeyEvent) { a.partyMode.Set(!a.partyMode.Get()) }),
		tui.On(tui.Rune('m'), func(ke tui.KeyEvent) { a.showModal.Set(true) }),
		tui.On(tui.Rune('+'), func(ke tui.KeyEvent) { a.counter.Set(a.counter.Get() + 1) }),
		tui.On(tui.Rune('-'), func(ke tui.KeyEvent) { a.counter.Set(a.counter.Get() - 1) }),
		tui.On(tui.Rune('0'), func(ke tui.KeyEvent) { a.counter.Set(0) }),
	}
}

func (a *crazyApp) HandleMouse(me tui.MouseEvent) bool {
	// Section toggle clicks via ref map
	if me.Button == tui.MouseLeft && me.Action == tui.MousePress {
		for id, ref := range a.sectionRefs {
			if el := ref.El(); el != nil && el.ContainsPoint(me.X, me.Y) {
				a.toggle(id)
				return true
			}
		}
		for th, ref := range a.themeRefs {
			if el := ref.El(); el != nil && el.ContainsPoint(me.X, me.Y) {
				a.theme.Set(th)
				return true
			}
		}
	}
	return tui.HandleClicks(me,
		tui.Click(a.ctrUp, func() { a.counter.Set(a.counter.Get() + 1) }),
		tui.Click(a.ctrDown, func() { a.counter.Set(a.counter.Get() - 1) }),
		tui.Click(a.ctrReset, func() { a.counter.Set(0) }),
		tui.Click(a.modalOpen, func() { a.showModal.Set(true) }),
		tui.Click(a.modalYes, func() { a.counter.Set(0); a.showModal.Set(false) }),
		tui.Click(a.modalNo, func() { a.showModal.Set(false) }),
	)
}

func (a *crazyApp) Watchers() []tui.Watcher {
	return []tui.Watcher{tui.OnTimer(16*time.Millisecond, a.animate)}
}

func (a *crazyApp) animate() {
	a.frame++
	step := 0.06
	if a.partyMode.Get() { step = 0.15 }
	a.wavePhase.Update(func(v float64) float64 { return v + step })
	a.scrollWave.Update(func(v float64) float64 { return v + 0.02 })
	a.borderHue.Update(func(v float64) float64 { return v + 0.5 })
	if a.frame%5 == 0 { a.spinFrame.Update(func(v int) int { return v + 1 }) }
	a.bouncePhase.Update(func(v float64) float64 { return v + 0.04 })
}

func (a *crazyApp) fps() float64 {
	e := time.Since(a.startTime).Seconds()
	if e < 0.01 { return 0 }
	return float64(a.frame) / e
}

func (a *crazyApp) mc() tui.Color {
	if t, ok := themeColors[a.theme.Get()]; ok { return t.fg }
	return tui.Cyan
}
func (a *crazyApp) ac() tui.Color {
	if t, ok := themeColors[a.theme.Get()]; ok { return t.accent }
	return tui.Magenta
}
func (a *crazyApp) isOn(id string) bool { return a.visible.Get()[id] }
func (a *crazyApp) toggle(id string) {
	a.visible.Update(func(m map[string]bool) map[string]bool {
		m2 := make(map[string]bool, len(m))
		for k, v := range m { m2[k] = v }
		m2[id] = !m2[id]
		return m2
	})
}

// ===== SIDEBAR (left) =====

func (a *crazyApp) renderSidebar(sw, h int) *tui.Element {
	mc := a.mc()
	sb := flex(tui.Column, tui.WithScrollable(tui.ScrollVertical),
		tui.WithScrollbarHidden(true),
		tui.WithWidth(sw),
		tui.WithHeight(h-2),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1), tui.WithGap(1))

	sb.AddChild(textEl("📋 SIDEBAR", tui.NewStyle().Bold().Foreground(mc)))

	// Panel toggles
	sb.AddChild(textEl("📋 Sections", tui.NewStyle().Bold().Foreground(mc).Dim()))
	for _, p := range []struct{ id, icon, label string }{
		{"bounce", "📦", "Bouncing Boxes"},
		{"progress", "📊", "Progress Parade"},
		{"metrics", "⚡", "Live Metrics"},
		{"fireworks", "🎆", "Fireworks"},
		{"map", "🌊", "Wave Scroller"},
		{"symbols", "🌀", "Symbol Storm"},
		{"matrix", "💚", "Matrix Rain"},
	} {
		on := a.isOn(p.id)
		check := "⬜"
		if on { check = "✅" }
		col := tui.BrightBlack
		if on { col = mc }
		btn := flex(tui.Row, tui.WithBorder(tui.BorderRounded),
			tui.WithBorderStyle(tui.NewStyle().Foreground(col)),
			tui.WithPadding(0))
		btn.AddChild(textEl(fmt.Sprintf("%s %s", check, p.label), tui.NewStyle().Foreground(col)))
		if ref, ok := a.sectionRefs[p.id]; ok { ref.Set(btn) }
		idCopy := p.id
		btn.SetOnFocus(func(e *tui.Element) { a.toggle(idCopy) })
		sb.AddChild(btn)
	}

	sb.AddChild(textEl("", tui.NewStyle()))

	// Theme
	sb.AddChild(textEl("🎨 Theme", tui.NewStyle().Bold().Foreground(mc).Dim()))
	for _, th := range themeList {
		col := tui.BrightBlack
		if a.theme.Get() == th { col = mc }
		btn := flex(tui.Row, tui.WithBorder(tui.BorderRounded),
			tui.WithBorderStyle(tui.NewStyle().Foreground(col)),
			tui.WithPadding(0))
		btn.AddChild(textEl(fmt.Sprintf("  %s", th), tui.NewStyle().Foreground(col)))
		if ref, ok := a.themeRefs[th]; ok { ref.Set(btn) }
		thCopy := th
		btn.SetOnFocus(func(e *tui.Element) { a.theme.Set(thCopy) })
		btn.AddChild(textEl(fmt.Sprintf("  %s", th), tui.NewStyle().Foreground(col)))
		sb.AddChild(btn)
	}

	sb.AddChild(textEl("", tui.NewStyle()))

	// Counter
	sb.AddChild(textEl("🔢 Counter", tui.NewStyle().Bold().Foreground(mc).Dim()))
	sb.AddChild(miniBtn("➖ Decrease", tui.Red, a.ctrDown, sw))
	val := a.counter.Get()
	vCol := tui.Cyan
	if val > 0 { vCol = tui.Green }
	if val < 0 { vCol = tui.Red }
	sb.AddChild(textEl(fmt.Sprintf("   Value: %d", val), tui.NewStyle().Bold().Foreground(vCol)))
	sb.AddChild(miniBtn("➕ Increase", tui.Green, a.ctrUp, sw))
	sb.AddChild(miniBtn("🔄 Reset", tui.BrightBlack, a.ctrReset, sw))

	sb.AddChild(textEl("", tui.NewStyle()))

	// Danger
	sb.AddChild(textEl("⚠️ Danger", tui.NewStyle().Bold().Foreground(tui.Red).Dim()))
	sb.AddChild(miniBtn("⚠️ Reset Modal", tui.Red, a.modalOpen, sw))

	sb.AddChild(textEl("", tui.NewStyle()))
	sb.AddChild(textEl("⌨️ Keys", tui.NewStyle().Bold().Foreground(mc).Dim()))
	for _, k := range []string{" q/Esc quit", " Space party", " M modal", " +/- counter"} {
		sb.AddChild(textEl(k, tui.NewStyle().Dim()))
	}

	return sb
}

func miniBtn(text string, col tui.Color, ref *tui.Ref, maxW int) *tui.Element {
	b := flex(tui.Row, tui.WithBorder(tui.BorderRounded),
		tui.WithBorderStyle(tui.NewStyle().Foreground(col)),
		tui.WithPadding(0))
	b.AddChild(textEl(" "+text, tui.NewStyle().Foreground(col)))
	b.AddChild(textEl(text, tui.NewStyle().Foreground(col).Dim()))
	if ref != nil { ref.Set(b) }
	return b
}

// ===== MAIN VIEW (right, scrollable) =====

func (a *crazyApp) Render(app *tui.App) *tui.Element {
	w, h := app.Size()
	sw := w * 35 / 100
	if sw < 30 { sw = 30 }

	outer := flex(tui.Row, tui.WithMinWidth(w), tui.WithMinHeight(h))
	outer.AddChild(a.renderSidebar(sw, h))

	// Main view: show only the selected panel
	main := flex(tui.Column, tui.WithFlexGrow(1),
		tui.WithPadding(1), tui.WithGap(1))

	main.AddChild(a.renderTitle())
	main.AddChild(a.renderTicker())

	for _, p := range []struct{ id string; fn func() *tui.Element }{
		{"bounce", a.renderBouncingBoxes},
		{"progress", a.renderProgress},
		{"metrics", a.renderMetrics},
		{"fireworks", a.renderFireworks},
		{"map", a.renderScrollableMap},
		{"symbols", a.renderSymbolStorm},
		{"matrix", a.renderMatrixRain},
	} {
		if a.isOn(p.id) {
			main.AddChild(p.fn())
		}
	}

	// Fallback empty state
	anyOn := false
	for _, id := range []string{"bounce", "progress", "metrics", "fireworks", "map", "symbols"} {
		if a.isOn(id) { anyOn = true; break }
	}
	if !anyOn {
		main.AddChild(textEl("✨ Select a section from the sidebar!", tui.NewStyle().Bold().Foreground(a.mc())))
	}

	main.AddChild(a.renderFooter())
	outer.AddChild(main)

	if modal := a.renderModalOverlay(); modal != nil { outer.AddChild(modal) }
	return outer
}

func (a *crazyApp) renderTitle() *tui.Element {
	t := "🚀 GO-TUI CRAZY DEMO 🎉"
	if a.partyMode.Get() { t = "🎊 PARTY MODE ACTIVATED! 🎊" }
	r := flex(tui.Row, tui.WithJustify(tui.JustifyCenter))
	r.AddChild(textEl(t, tui.NewStyle().Bold().Foreground(a.mc())))
	return r
}

func (a *crazyApp) renderTicker() *tui.Element {
	w := flex(tui.Column, tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1))
	w.AddChild(textEl("📡 SCROLLING TICKER", tui.NewStyle().Bold().Foreground(a.mc())))
	w.AddChild(textEl(tickerText(a.scrollWave.Get()), hslStyle(a.wavePhase.Get()*36, 1.0, 0.6)))
	return w
}

func (a *crazyApp) renderBouncingBoxes() *tui.Element {
	r := flex(tui.Column, tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1), tui.WithGap(1))
	r.AddChild(textEl("📦 BOUNCING BOXES", tui.NewStyle().Bold().Foreground(a.mc())))
	row := flex(tui.Row, tui.WithGap(2), tui.WithJustify(tui.JustifyCenter))
	labels := []string{"🚀", "🎨", "⚡", "🌈"}
	colors := []tui.Color{tui.Cyan, tui.Magenta, tui.Yellow, tui.Green}
	for i := 0; i < 4; i++ {
		phase := a.bouncePhase.Get() + float64(i)*1.5
		hue := math.Mod(float64(i)*60+a.borderHue.Get(), 360)
		rr, gg, bb := hslToRGB(hue, 1.0, 0.6)
		bs := tui.NewStyle().Foreground(tui.RGBColor(rr, gg, bb))

		// Use margin-top to create visible bounce (3-10px oscillation)
		bounce := int(math.Abs(math.Sin(phase)) * 7)
		box := flex(tui.Column, tui.WithAlign(tui.AlignCenter),
			tui.WithMinWidth(10), tui.WithHeight(7),
			tui.WithBorder(tui.BorderRounded), tui.WithBorderStyle(bs),
			tui.WithMarginTRBL(bounce, 0, 0, 0),
			tui.WithPadding(1))
		box.AddChild(textEl(labels[i], bs.Bold()))
		box.AddChild(textEl(fmt.Sprintf("BOX %d", i+1), tui.NewStyle().Bold().Foreground(colors[i])))
		row.AddChild(box)
	}
	r.AddChild(row)
	return r
}

func (a *crazyApp) renderProgress() *tui.Element {
	elapsed := time.Since(a.startTime).Seconds()
	ct := math.Mod(elapsed, 4.0)
	lin := ct / 3.0
	if lin > 1 { lin = 1 }
	w := flex(tui.Column, tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1), tui.WithGap(1))
	w.AddChild(textEl("📊 PROGRESS PARADE", tui.NewStyle().Bold().Foreground(a.mc())))
	w.AddChild(textEl("3s loop: Linear vs Cubic Ease", tui.NewStyle().Dim()))
	w.AddChild(barRow("📈 Linear:", renderBar(lin, 30), tui.Cyan))
	w.AddChild(barRow("🎯 Eased:", renderBar(easeInOutCubic(lin), 30), tui.Green))
	return w
}

func barRow(label, bar string, col tui.Color) *tui.Element {
	r := flex(tui.Row, tui.WithGap(1))
	// Fixed-width label area so progress bars start at same column
	r.AddChild(tui.New(tui.WithText(label), tui.WithTextStyle(tui.NewStyle().Dim()), tui.WithWidth(12)))
	r.AddChild(textEl(bar, tui.NewStyle().Foreground(col)))
	return r
}

func (a *crazyApp) renderMetrics() *tui.Element {
	f := a.spinFrame.Get()
	w := flex(tui.Column, tui.WithFlexGrow(1),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1), tui.WithGap(1))
	w.AddChild(textEl("⚡ LIVE METRICS", tui.NewStyle().Bold().Foreground(a.mc())))
	inner := flex(tui.Row, tui.WithGap(2))
	c1 := flex(tui.Column, tui.WithGap(1))
	for _, s := range []struct{ ch string; c tui.Color; l string }{
		{spinnerDots[f%len(spinnerDots)], tui.Cyan, "Dots"},
		{spinnerLine[f%len(spinnerLine)], tui.Green, "Line"},
		{spinnerCircle[f%len(spinnerCircle)], tui.Yellow, "Circle"},
		{spinnerBraille[f%len(spinnerBraille)], tui.Magenta, "Braille"},
	} { c1.AddChild(spinRow(s.ch, s.c, s.l)) }
	inner.AddChild(c1)
	c2 := flex(tui.Column, tui.WithGap(1))
	c2.AddChild(textEl(fmt.Sprintf("🖼️ Frames: %d", a.frame), tui.NewStyle().Foreground(tui.Blue)))
	c2.AddChild(textEl(fmt.Sprintf("⏱️ Time: %.1fs", time.Since(a.startTime).Seconds()), tui.NewStyle().Foreground(tui.Green)))
	c2.AddChild(textEl(fmt.Sprintf("⚡ FPS: %.0f", a.fps()), tui.NewStyle().Foreground(tui.Magenta)))
	if a.partyMode.Get() {
		pm := []string{"🎉", "🎊", "✨", "🌟", "💥", "🔥", "💫"}
		c2.AddChild(textEl(pm[a.frame%len(pm)]+" PARTY! "+pm[(a.frame+3)%len(pm)], hslStyle(float64(a.frame)*36, 1.0, 0.6)))
	}
	inner.AddChild(c2)
	w.AddChild(inner)
	return w
}

func spinRow(s string, c tui.Color, l string) *tui.Element {
	r := flex(tui.Row, tui.WithGap(1))
	r.AddChild(textEl(s, tui.NewStyle().Foreground(c)))
	r.AddChild(textEl(l, tui.NewStyle().Dim()))
	return r
}

func (a *crazyApp) renderFireworks() *tui.Element {
	w := flex(tui.Column, tui.WithFlexGrow(1),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.BrightBlack)),
		tui.WithPadding(1))
	w.AddChild(textEl("🎆 FIREWORKS", tui.NewStyle().Bold().Foreground(tui.Yellow)))
	grid := flex(tui.Column, tui.WithGap(0), tui.WithMinHeight(8))
	t := time.Since(a.startTime).Seconds()
	cs := []tui.Color{tui.Red, tui.Yellow, tui.Green, tui.Cyan, tui.Blue, tui.Magenta}
	for row := 0; row < 6; row++ {
		var sb strings.Builder
		for col := 0; col < 30; col++ {
			val := math.Sin(float64(col)*0.5+t*3+float64(row)*0.3) * math.Cos(float64(row)*0.5+t*2)
			switch {
			case val > 0.6: sb.WriteRune('●')
			case val > 0.2: sb.WriteRune('·')
			case val < -0.6: sb.WriteRune('◆')
			case val < -0.2: sb.WriteRune('·')
			default: sb.WriteByte(' ')
			}
		}
		grid.AddChild(textEl(sb.String(), tui.NewStyle().Foreground(cs[row])))
	}
	w.AddChild(grid)
	return w
}

func (a *crazyApp) renderScrollableMap() *tui.Element {
	w := flex(tui.Column, tui.WithFlexGrow(1),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1), tui.WithMinHeight(10))
	w.AddChild(textEl("🌊 WAVE SCROLLER", tui.NewStyle().Bold().Foreground(a.mc())))

	// Ping-pong scroll position
	const totalRows = 60
	t := a.scrollWave.Get()
	pos := int(t * 12)
	cycle := totalRows * 2
	pos = pos % cycle
	if pos > totalRows {
		pos = cycle - pos
	}

	sc := flex(tui.Column, tui.WithScrollable(tui.ScrollVertical),
		tui.WithScrollOffset(0, pos), tui.WithMinHeight(10))
	in := flex(tui.Column, tui.WithMinWidth(120), tui.WithMinHeight(totalRows+5))

	phase := a.wavePhase.Get()
	for row := 0; row < totalRows; row++ {
		var sb strings.Builder
		for col := 0; col < 40; col++ {
			// Sine wave with phase shift per row
			val := math.Sin(float64(col)*0.4+float64(row)*0.15+phase*2)
			// Second overlapping wave for visual interest
			val2 := math.Sin(float64(col)*0.2-float64(row)*0.1+phase*3) * 0.5
			combined := val + val2

			ch := " "
			if combined > 0.8 {
				ch = "█"
			} else if combined > 0.5 {
				ch = "▓"
			} else if combined > 0.2 {
				ch = "▒"
			} else if combined > -0.1 {
				ch = "░"
			} else if combined > -0.4 {
				ch = "·"
			}
			sb.WriteString(ch)
		}
		// Color gradient from top to bottom
		hue := math.Mod(float64(row)*3+phase*20, 360)
		rr, gg, bb := hslToRGB(hue, 0.8, 0.5)
		in.AddChild(textEl(sb.String(), tui.NewStyle().Foreground(tui.RGBColor(rr, gg, bb))))
	}
	sc.AddChild(in)
	w.AddChild(sc)
	return w
}

// ===== SYMBOL STORM =====

func randomSymbol(phase float64, idx int) (string, tui.Color) {
	i := int(math.Sin(phase+float64(idx)*0.7)*50 + 50) % len(symbolPool)
	if i < 0 { i += len(symbolPool) }
	h := math.Mod(float64(idx)*37+phase*60, 360)
	r, g, b := hslToRGB(h, 1.0, 0.5+0.3*math.Sin(phase+float64(idx)*0.3))
	return symbolPool[i], tui.RGBColor(r, g, b)
}

func (a *crazyApp) renderSymbolStorm() *tui.Element {
	w := flex(tui.Column, tui.WithMinHeight(10),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(a.ac())),
		tui.WithPadding(1))
	w.AddChild(textEl("🌀 SYMBOL STORM", tui.NewStyle().Bold().Foreground(a.mc())))
	phase := a.wavePhase.Get()
	sc := flex(tui.Column, tui.WithScrollable(tui.ScrollVertical),
		tui.WithScrollOffset(0, int(phase*10)%60),
		tui.WithMinHeight(7))
	in := flex(tui.Column, tui.WithMinWidth(200), tui.WithMinHeight(300))
	for row := 0; row < 20; row++ {
		var sb strings.Builder
		for col := 0; col < 30; col++ {
			sym, _ := randomSymbol(phase, row*30+col)
			sb.WriteString(sym)
		}
		in.AddChild(textEl(sb.String(), tui.NewStyle().Foreground(a.mc()).Dim()))
	}
	sc.AddChild(in)
	w.AddChild(sc)
	return w
}

func (a *crazyApp) renderMatrixRain() *tui.Element {
	w := flex(tui.Column, tui.WithFlexGrow(1),
		tui.WithBorder(tui.BorderSingle),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Green)),
		tui.WithPadding(1), tui.WithMinHeight(12))
	w.AddChild(textEl("💚 MATRIX RAIN", tui.NewStyle().Bold().Foreground(tui.Green)))

	in := flex(tui.Column, tui.WithMinWidth(300), tui.WithMinHeight(400))

	chars := []string{
		"ﾀ","ﾁ","ﾂ","ﾃ","ﾄ","ﾅ","ﾆ","ﾇ","ﾈ","ﾉ",
		"ﾊ","ﾋ","ﾌ","ﾍ","ﾎ","ﾏ","ﾐ","ﾑ","ﾒ","ﾓ",
		"ﾔ","ﾕ","ﾖ","ﾗ","ﾘ","ﾙ","ﾚ","ﾛ","ﾜ","ｦ",
		"1","0",":",";",".","-","+","*",
	}

	const cols = 30
	const rows = 30

	for row := 0; row < rows; row++ {
		rowEl := flex(tui.Row, tui.WithGap(0))
		for col := 0; col < cols; col++ {
			// Advance this columns phase (persists across frames in matrixPhase array)
			speed := 1 + ((col*7+3)*11)%8
			a.matrixPhase[col] += float64(speed) * 0.002

			startOff := ((col*13+7)*37) % 50
			dropLen := 8 + ((col*7+3)*11)%10
			gapLen := 6 + ((col*11+5)*7)%20
			cycleLen := dropLen + gapLen
			cyclePos := (int(a.matrixPhase[col]) + startOff) % cycleLen
			dropPos := cyclePos - gapLen
			dist := dropPos - row

			chIdx := (row*13 + col*7 + a.frame*2) % len(chars)
			ch := chars[chIdx]

			if dist >= 0 && dist < dropLen {
				switch {
				case dist == 0:
					ch = chars[(col*7+a.frame*3)%len(chars)]
					rowEl.AddChild(textEl(ch, tui.NewStyle().Bold().Foreground(tui.RGBColor(200,255,200))))
				case dist < 3:
					rowEl.AddChild(textEl(ch, tui.NewStyle().Foreground(tui.Green)))
				case dist < 5:
					rowEl.AddChild(textEl(ch, tui.NewStyle().Foreground(tui.RGBColor(0, 160, 0))))
				case dist < 7:
					rowEl.AddChild(textEl(ch, tui.NewStyle().Foreground(tui.RGBColor(0, 90, 0))))
				default:
					rowEl.AddChild(textEl(ch, tui.NewStyle().Foreground(tui.RGBColor(0, 40, 0))))
				}
			} else {
				rowEl.AddChild(textEl(" ", tui.NewStyle()))
			}
		}
		in.AddChild(rowEl)
	}

	sc := flex(tui.Column, tui.WithScrollable(tui.ScrollVertical),
		tui.WithScrollbarHidden(true),
		tui.WithMinHeight(9))
	sc.AddChild(in)
	w.AddChild(sc)
	return w
}

func (a *crazyApp) renderFooter() *tui.Element {
	r := flex(tui.Row, tui.WithGap(2), tui.WithJustify(tui.JustifyCenter))
	r.AddChild(textEl("⌨️ q/Esc quit", tui.NewStyle().Dim()))
	r.AddChild(textEl("🎉 Space party", tui.NewStyle().Dim()))
	if a.partyMode.Get() {
		r.AddChild(textEl("🎊 PARTY MODE", hslStyle(float64(a.frame)*36, 1.0, 0.6)))
	}
	return r
}

func (a *crazyApp) renderModalOverlay() *tui.Element {
	if !a.showModal.Get() { return nil }
	modal := flex(tui.Column, tui.WithBorder(tui.BorderDouble),
		tui.WithBorderStyle(tui.NewStyle().Foreground(tui.Red)),
		tui.WithPadding(2), tui.WithGap(1),
		tui.WithAlign(tui.AlignCenter), tui.WithMinWidth(30))
	modal.AddChild(textEl("⚠️ CONFIRM RESET", tui.NewStyle().Bold().Foreground(tui.Red)))
	modal.AddChild(textEl("Reset counter to 0?", tui.NewStyle().Dim()))
	modal.AddChild(textEl(fmt.Sprintf("Current: %d", a.counter.Get()), tui.NewStyle().Foreground(tui.Yellow)))
	btnRow := flex(tui.Row, tui.WithGap(2), tui.WithJustify(tui.JustifyCenter))
	btnRow.AddChild(miniBtn("✅ Yes", tui.Green, a.modalYes, 20))
	btnRow.AddChild(miniBtn("❌ No", tui.Red, a.modalNo, 20))
	modal.AddChild(btnRow)
	overlay := tui.New(tui.WithOverlay(true), tui.WithDirection(tui.Column),
		tui.WithJustify(tui.JustifyCenter), tui.WithAlign(tui.AlignCenter))
	overlay.AddChild(modal)
	return overlay
}
