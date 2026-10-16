package ui

import (
	"strings"
	"unicode"
)

// Editor is a single-line text editor model with caret and selection. It is
// pure (no Win32) so it can be unit tested.
type Editor struct {
	text   []rune
	caret  int
	anchor int // selection anchor; equals caret when nothing is selected
}

func (e *Editor) Text() string { return string(e.text) }
func (e *Editor) Len() int     { return len(e.text) }
func (e *Editor) Caret() int   { return e.caret }

func (e *Editor) SetText(s string) {
	e.text = []rune(s)
	e.caret = len(e.text)
	e.anchor = e.caret
}

func (e *Editor) Clear() { e.text, e.caret, e.anchor = nil, 0, 0 }

// Selection returns the ordered selection range.
func (e *Editor) Selection() (lo, hi int) {
	if e.anchor < e.caret {
		return e.anchor, e.caret
	}
	return e.caret, e.anchor
}

func (e *Editor) HasSelection() bool { return e.anchor != e.caret }

func (e *Editor) SelectedText() string {
	lo, hi := e.Selection()
	return string(e.text[lo:hi])
}

func (e *Editor) SelectAll() { e.anchor, e.caret = 0, len(e.text) }

func (e *Editor) deleteSelection() bool {
	if !e.HasSelection() {
		return false
	}
	lo, hi := e.Selection()
	e.text = append(e.text[:lo], e.text[hi:]...)
	e.caret, e.anchor = lo, lo
	return true
}

// Insert replaces the selection with s. Control characters and newlines are
// dropped (newlines become spaces for pasted text).
func (e *Editor) Insert(s string) {
	s = strings.Map(func(r rune) rune {
		switch {
		case r == '\n' || r == '\r' || r == '\t':
			return ' '
		case unicode.IsControl(r):
			return -1
		}
		return r
	}, s)
	e.deleteSelection()
	rs := []rune(s)
	out := make([]rune, 0, len(e.text)+len(rs))
	out = append(out, e.text[:e.caret]...)
	out = append(out, rs...)
	out = append(out, e.text[e.caret:]...)
	e.text = out
	e.caret += len(rs)
	e.anchor = e.caret
}

func (e *Editor) Backspace() {
	if e.deleteSelection() || e.caret == 0 {
		return
	}
	e.text = append(e.text[:e.caret-1], e.text[e.caret:]...)
	e.caret--
	e.anchor = e.caret
}

func (e *Editor) Delete() {
	if e.deleteSelection() || e.caret >= len(e.text) {
		return
	}
	e.text = append(e.text[:e.caret], e.text[e.caret+1:]...)
	e.anchor = e.caret
}

func isWordRune(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' }

// wordLeft returns the start of the word before i (skipping spaces first).
func (e *Editor) wordLeft(i int) int {
	for i > 0 && !isWordRune(e.text[i-1]) {
		i--
	}
	for i > 0 && isWordRune(e.text[i-1]) {
		i--
	}
	return i
}

func (e *Editor) wordRight(i int) int {
	for i < len(e.text) && isWordRune(e.text[i]) {
		i++
	}
	for i < len(e.text) && !isWordRune(e.text[i]) {
		i++
	}
	return i
}

// CtrlBackspace deletes the previous word (or the selection).
func (e *Editor) CtrlBackspace() {
	if e.deleteSelection() {
		return
	}
	start := e.wordLeft(e.caret)
	e.text = append(e.text[:start], e.text[e.caret:]...)
	e.caret, e.anchor = start, start
}

func (e *Editor) CtrlDelete() {
	if e.deleteSelection() {
		return
	}
	end := e.wordRight(e.caret)
	e.text = append(e.text[:e.caret], e.text[end:]...)
	e.anchor = e.caret
}

// Move moves the caret. dir is -1 or +1; word moves by word; extend keeps the
// selection anchor. Without extend, an existing selection collapses to its edge.
func (e *Editor) Move(dir int, word, extend bool) {
	if !extend && e.HasSelection() && !word {
		lo, hi := e.Selection()
		if dir < 0 {
			e.caret = lo
		} else {
			e.caret = hi
		}
		e.anchor = e.caret
		return
	}
	switch {
	case word && dir < 0:
		e.caret = e.wordLeft(e.caret)
	case word:
		e.caret = e.wordRight(e.caret)
	default:
		e.caret += dir
	}
	e.clamp(extend)
}

func (e *Editor) Home(extend bool) { e.caret = 0; e.clamp(extend) }
func (e *Editor) End(extend bool)  { e.caret = len(e.text); e.clamp(extend) }

// SetCaret places the caret at i, optionally extending the selection.
func (e *Editor) SetCaret(i int, extend bool) { e.caret = i; e.clamp(extend) }

func (e *Editor) clamp(extend bool) {
	if e.caret < 0 {
		e.caret = 0
	}
	if e.caret > len(e.text) {
		e.caret = len(e.text)
	}
	if !extend {
		e.anchor = e.caret
	}
}
