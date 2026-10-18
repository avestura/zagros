//go:build windows

package ui

import (
	"math"
	"time"
)

const (
	smoothTau    = 70 * time.Millisecond // easing time constant: the glide covers ~95% in 3 tau
	smoothTickMS = 15
)

// smooth is a pixel scroll position that glides toward a target, so a wheel
// notch moves the list over a few frames instead of jumping whole rows.
type smooth struct {
	pos, target float64
	last        time.Time
}

// clamp keeps pos and target inside the scrollable range of a list of content
// pixels shown through a view of view pixels.
func (s *smooth) clamp(content, view int) {
	hi := float64(max(content-view, 0))
	s.pos = math.Min(math.Max(s.pos, 0), hi)
	s.target = math.Min(math.Max(s.target, 0), hi)
}

func (s *smooth) by(px float64) {
	if !s.moving() {
		s.last = time.Now()
	}
	s.target += px
}

// jump moves to px at once (thumb dragging follows the mouse exactly).
func (s *smooth) jump(px float64) { s.pos, s.target = px, px }

func (s *smooth) moving() bool { return math.Abs(s.target-s.pos) >= 0.5 }

// step advances the glide to now and reports whether it is still moving.
func (s *smooth) step(now time.Time) bool {
	dt := now.Sub(s.last)
	s.last = now
	if !s.moving() {
		s.pos = s.target
		return false
	}
	dt = min(max(dt, time.Millisecond), 50*time.Millisecond)
	s.pos += (s.target - s.pos) * (1 - math.Exp(-float64(dt)/float64(smoothTau)))
	if !s.moving() {
		s.pos = s.target
		return false
	}
	return true
}
