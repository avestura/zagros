//go:build windows

package ui

import (
	"strings"
)

// DefaultExplanation returns a plain-English description for a permission
// name when the extension host did not supply one.
func DefaultExplanation(name string) string {
	n := strings.ToLower(name)
	switch {
	case strings.HasPrefix(n, "net"), strings.HasPrefix(n, "http"):
		return "Connect to the internet and exchange data with the listed servers."
	case strings.Contains(n, "write"):
		return "Create, change or delete files in the listed location."
	case strings.HasPrefix(n, "fs"), strings.HasPrefix(n, "file"):
		return "Read files in the listed location."
	case strings.HasPrefix(n, "clip"):
		return "Read or replace the contents of your clipboard."
	case strings.HasPrefix(n, "exec"), strings.HasPrefix(n, "run"), strings.HasPrefix(n, "proc"):
		return "Start programs on your computer."
	case strings.HasPrefix(n, "notif"):
		return "Show notifications from the system tray."
	case strings.HasPrefix(n, "env"):
		return "Read environment variables."
	case strings.HasPrefix(n, "system"):
		return "See running programs, open window titles, services and folder locations."
	case strings.HasPrefix(n, "reg"):
		return "Read the Windows registry."
	case strings.HasPrefix(n, "secret"), strings.HasPrefix(n, "pass"), strings.HasPrefix(n, "cred"):
		return "Store and read secrets for this extension."
	case n == "open":
		return "Open files, folders and links with their default apps."
	case strings.HasPrefix(n, "store"), strings.HasPrefix(n, "data"):
		return "Keep its own settings and data on disk."
	}
	return "Use the \"" + name + "\" capability."
}

type promptState struct {
	req     PermissionRequest
	allow   []int8 // -1 undecided, 0 deny, 1 allow
	scrollY int
	done    chan<- PermissionResult
	sent    bool
	// layout cache
	rowH    []int
	content Rect
}

// PromptPermissions shows the consent dialog and blocks until the user
// decides. It is safe to call from any goroutine and implements the ext
// permission prompter. Nothing is stored here.
func (a *App) PromptPermissions(req PermissionRequest) PermissionResult {
	if a.hwnd == 0 {
		return PermissionResult{Cancelled: true}
	}
	ch := make(chan PermissionResult, 1)
	a.Post(func() { a.openPrompt(req, ch) })
	select {
	case r := <-ch:
		return r
	case <-a.quitCh:
		return PermissionResult{Cancelled: true}
	}
}

func (a *App) openPrompt(req PermissionRequest, ch chan<- PermissionResult) {
	st := &promptState{req: req, done: ch, allow: make([]int8, len(req.Asks))}
	for i := range st.allow {
		st.allow[i] = -1
	}
	p := a.newPanel("Permission request", 560, 200+78*len(req.Asks))
	p.draw = st.draw
	p.onBtn = func(id int) {
		switch {
		case id == 1: // allow all
			for i := range st.allow {
				st.allow[i] = 1
			}
		case id == 2: // done
			st.finish(false)
			p.close()
		case id == 3:
			st.finish(true)
			p.close()
		case id >= 1000:
			i := (id - 1000) / 2
			if i < len(st.allow) {
				st.allow[i] = int8(1 - (id-1000)%2)
			}
		}
	}
	p.onWheel = func(d int) { st.scrollY += d * p.s(40) }
	p.onKey = func(vk int) bool {
		if vk == vkReturn {
			st.finish(false)
			p.close()
			return true
		}
		return false
	}
	p.onClose = func() { st.finish(true) }
	p.show()
}

func (st *promptState) finish(cancel bool) {
	if st.sent {
		return
	}
	st.sent = true
	res := PermissionResult{Cancelled: cancel}
	if !cancel {
		for i, a := range st.req.Asks {
			if st.allow[i] == 1 {
				res.Granted = append(res.Granted, a.Name)
			}
		}
	}
	st.done <- res
}

func (st *promptState) draw(p *panel, dc uintptr) {
	c := p.th.Colors
	pad := p.s(14)
	t := p.titleH()
	y := t + pad
	p.text(dc, p.fBold, pad, y, "Allow "+st.req.Extension+" to use the following?", uint32(c.Text))
	y += p.hMain + p.s(4)
	who := "Publisher: " + orDash(st.req.Publisher) + "    Source: " + orDash(st.req.Source)
	p.textFit(dc, p.fSmall, pad, y, p.w-2*pad, who, uint32(c.TextDim))
	y += p.hSmall + pad

	footerH := p.s(56)
	st.content = Rect{pad, y, p.w - pad, p.h - footerH}
	// compute row heights
	pSelectObject.Call(dc, p.fSmall)
	textW := st.content.W() - p.s(170)
	st.rowH = st.rowH[:0]
	total := 0
	lines := make([][]string, len(st.req.Asks))
	for i, ask := range st.req.Asks {
		ex := ask.Explanation
		if ex == "" {
			ex = DefaultExplanation(ask.Name)
		}
		lines[i] = wrapText(dc, ex, textW, 3)
		n := len(lines[i])
		if ask.Scope != "" {
			n++
		}
		h := p.s(12) + p.hMain + n*p.hSmall
		st.rowH = append(st.rowH, h)
		total += h
	}
	if maxS := total - st.content.H(); st.scrollY > maxS {
		st.scrollY = maxS
	}
	if st.scrollY < 0 {
		st.scrollY = 0
	}
	restore := p.clip(dc, st.content)
	ry := st.content.T - st.scrollY
	for i, ask := range st.req.Asks {
		h := st.rowH[i]
		row := Rect{st.content.L, ry, st.content.R, ry + h}
		p.fill(dc, Rect{row.L, row.B - 1, row.R, row.B}, uint32(c.Border))
		tx := row.L
		ty := row.T + p.s(6)
		label := ask.Label
		if label == "" {
			label = ask.Name
		}
		p.text(dc, p.fBold, tx, ty, label, uint32(c.Text))
		ty += p.hMain
		if ask.Scope != "" {
			p.textFit(dc, p.fSmall, tx, ty, textW, "Scope: "+ask.Scope, uint32(c.Accent))
			ty += p.hSmall
		}
		for _, l := range lines[i] {
			p.text(dc, p.fSmall, tx, ty, l, uint32(c.TextDim))
			ty += p.hSmall
		}
		bw, bh := p.s(70), p.s(28)
		by := row.T + (h-bh)/2
		deny := Rect{row.R - bw, by, row.R, by + bh}
		allow := Rect{deny.L - p.s(8) - bw, by, deny.L - p.s(8), by + bh}
		for _, b := range []struct {
			r     Rect
			id    int
			label string
			on    bool
		}{{allow, 1000 + 2*i, "Allow", st.allow[i] == 1}, {deny, 1001 + 2*i, "Deny", st.allow[i] == 0}} {
			if vr, ok := intersect(b.r, st.content); ok {
				p.onTint = 0x2E8B57 // allow: green
				if b.label == "Deny" {
					p.onTint = 0xB3392F
				}
				p.button(dc, b.r, b.id, b.label, false, false, b.on)
				p.onTint = 0
				p.btns[len(p.btns)-1].r = vr
			}
		}
		ry += h
	}
	restore()

	// footer
	fy := p.h - footerH + p.s(12)
	bh := p.s(32)
	bw := p.s(96)
	right := p.w - pad
	p.button(dc, Rect{right - bw, fy, right, fy + bh}, 3, "Cancel", false, false, false)
	right -= bw + p.s(8)
	p.button(dc, Rect{right - bw, fy, right, fy + bh}, 2, "Done", false, true, false)
	right -= bw + p.s(8)
	p.button(dc, Rect{right - bw, fy, right, fy + bh}, 1, "Allow all", false, false, false)
	p.textFit(dc, p.fSmall, pad, fy+(bh-p.hSmall)/2, right-bw-pad, "Permissions left undecided are denied.", uint32(c.TextDim))
}

func orDash(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
