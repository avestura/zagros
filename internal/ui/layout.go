package ui

import "github.com/avestura/zagros/internal/theme"

// Rect is an integer rectangle (right/bottom exclusive).
type Rect struct{ L, T, R, B int }

func (r Rect) W() int { return r.R - r.L }
func (r Rect) H() int { return r.B - r.T }
func (r Rect) Contains(x, y int) bool {
	return x >= r.L && x < r.R && y >= r.T && y < r.B
}

// Metrics are the theme metrics scaled to physical pixels.
type Metrics struct {
	Width, InputH, RowH, Pad, Border, StatusH int
	MaxRows                                   int
}

// NormalizeTheme fills in defaults for zero-valued theme fields.
func NormalizeTheme(t theme.Theme) theme.Theme {
	def := func(p *int, v int) {
		if *p <= 0 {
			*p = v
		}
	}
	if t.Font == "" {
		t.Font = "Segoe UI"
	}
	if t.MonoFont == "" {
		t.MonoFont = "Consolas"
	}
	def(&t.FontSize, 11)
	def(&t.Width, 640)
	def(&t.InputHeight, 48)
	def(&t.RowHeight, 44)
	def(&t.Padding, 12)
	if t.BorderWidth < 0 {
		t.BorderWidth = 0
	}
	if t.Radius < 0 {
		t.Radius = 0
	}
	return t
}

// Scale multiplies a 96-DPI value by dpi/96, rounding to nearest.
func Scale(v, dpi int) int {
	if dpi <= 0 {
		dpi = 96
	}
	return (v*dpi + 48) / 96
}

// ComputeMetrics scales the theme metrics. statusH is the status line height
// in physical pixels (derived from the font).
func ComputeMetrics(t theme.Theme, dpi, maxRows, statusH int) Metrics {
	if maxRows <= 0 {
		maxRows = 8
	}
	m := Metrics{
		Width:   Scale(t.Width, dpi),
		InputH:  Scale(t.InputHeight, dpi),
		RowH:    Scale(t.RowHeight, dpi),
		Pad:     Scale(t.Padding, dpi),
		Border:  Scale(t.BorderWidth, dpi),
		StatusH: statusH,
		MaxRows: maxRows,
	}
	if t.BorderWidth > 0 && m.Border < 1 {
		m.Border = 1
	}
	return m
}

// VisibleRows is how many rows are shown for n results.
func (m Metrics) VisibleRows(n int) int {
	if n > m.MaxRows {
		return m.MaxRows
	}
	if n < 0 {
		return 0
	}
	return n
}

// Height is the total window height for n results.
func (m Metrics) Height(n int) int {
	v := m.VisibleRows(n)
	h := 2*m.Border + m.InputH
	if v > 0 {
		h += v*m.RowH + m.StatusH
	}
	return h
}

func (m Metrics) InputRect() Rect {
	return Rect{m.Border, m.Border, m.Width - m.Border, m.Border + m.InputH}
}

// RowRect is the rectangle of visible row i (0-based from the top of the list).
func (m Metrics) RowRect(i int) Rect {
	top := m.Border + m.InputH + i*m.RowH
	return Rect{m.Border, top, m.Width - m.Border, top + m.RowH}
}

func (m Metrics) StatusRect(n int) Rect {
	v := m.VisibleRows(n)
	top := m.Border + m.InputH + v*m.RowH
	return Rect{m.Border, top, m.Width - m.Border, top + m.StatusH}
}

// HitRow returns the visible-row index under (x,y), or -1.
func (m Metrics) HitRow(x, y, n int) int {
	v := m.VisibleRows(n)
	if x < m.Border || x >= m.Width-m.Border {
		return -1
	}
	y -= m.Border + m.InputH
	if y < 0 || m.RowH <= 0 {
		return -1
	}
	i := y / m.RowH
	if i >= v {
		return -1
	}
	return i
}

// Scroll tracks selection and the scroll window over a list.
type Scroll struct {
	N, Vis, Top, Sel int
}

func (s *Scroll) Reset(n, vis int) {
	s.N, s.Vis, s.Top, s.Sel = n, vis, 0, 0
}

// SetCount updates the list size, clamping selection and window.
func (s *Scroll) SetCount(n, vis int) {
	s.N, s.Vis = n, vis
	if s.Sel >= n {
		s.Sel = n - 1
	}
	if s.Sel < 0 {
		s.Sel = 0
	}
	s.ensure()
}

func (s *Scroll) ensure() {
	if s.Sel < s.Top {
		s.Top = s.Sel
	}
	if s.Vis > 0 && s.Sel >= s.Top+s.Vis {
		s.Top = s.Sel - s.Vis + 1
	}
	if max := s.N - s.Vis; s.Top > max {
		s.Top = max
	}
	if s.Top < 0 {
		s.Top = 0
	}
}

// Select moves the selection to i (clamped).
func (s *Scroll) Select(i int) {
	if i >= s.N {
		i = s.N - 1
	}
	if i < 0 {
		i = 0
	}
	s.Sel = i
	s.ensure()
}

// Move moves the selection by delta. With wrap, stepping past an end wraps.
func (s *Scroll) Move(delta int, wrap bool) {
	if s.N == 0 {
		return
	}
	i := s.Sel + delta
	if wrap && (delta == 1 || delta == -1) {
		i = (i + s.N) % s.N
	}
	s.Select(i)
}

func (s *Scroll) PageDown() { s.Move(max(s.Vis-1, 1), false) }
func (s *Scroll) PageUp()   { s.Move(-max(s.Vis-1, 1), false) }
func (s *Scroll) Home()     { s.Select(0) }
func (s *Scroll) End()      { s.Select(s.N - 1) }

// Wheel scrolls the window by delta rows without moving the selection beyond it.
func (s *Scroll) Wheel(delta int) {
	s.Top += delta
	if max := s.N - s.Vis; s.Top > max {
		s.Top = max
	}
	if s.Top < 0 {
		s.Top = 0
	}
}

// Abs converts a visible-row index to an absolute item index, or -1.
func (s *Scroll) Abs(vis int) int {
	i := s.Top + vis
	if vis < 0 || i >= s.N || vis >= s.Vis {
		return -1
	}
	return i
}

// Rel converts an absolute index to a visible index, or -1 if scrolled out.
func (s *Scroll) Rel(abs int) int {
	r := abs - s.Top
	if r < 0 || r >= s.Vis {
		return -1
	}
	return r
}

// CaretIndexAt maps an x offset within text to a rune index, given the
// cumulative pixel width of each prefix (widths[i] = width of first i runes,
// len = runes+1).
func CaretIndexAt(widths []int, x int) int {
	for i := 1; i < len(widths); i++ {
		if x < (widths[i-1]+widths[i])/2 {
			return i - 1
		}
	}
	return len(widths) - 1
}

// HScroll returns the horizontal scroll offset so the caret stays visible in
// a view of viewW pixels, given the previous offset.
func HScroll(caretX, textW, viewW, off int) int {
	if textW <= viewW {
		return 0
	}
	if caretX-off > viewW {
		off = caretX - viewW
	}
	if caretX-off < 0 {
		off = caretX
	}
	if off > textW-viewW+2 {
		off = textW - viewW + 2
	}
	if off < 0 {
		off = 0
	}
	return off
}
