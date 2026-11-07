package gui

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"time"

	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"golang.org/x/exp/shiny/materialdesign/icons"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

func rgb(v uint32) color.NRGBA {
	return color.NRGBA{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 0xff}
}

var (
	colAccent  = rgb(0xdd0428) // official Flashpoint red, used sparingly
	colLight   = rgb(0xe85971) // official Flashpoint light red: disabled primary buttons
	colInk     = rgb(0x1f1f1f)
	colMuted   = rgb(0x666666)
	colLine    = rgb(0xdedede)
	colPanel   = rgb(0xf4f4f4)
	colSel     = rgb(0xfce6e9) // the red at 10 % on white
	colGreen   = rgb(0x2e7d32)
	colAmber   = rgb(0x9a5b00)
	colBlue    = rgb(0x1565c0)
	colBanner  = rgb(0xfff4ce)
	colScrim   = color.NRGBA{A: 0x70}
	colWhite   = rgb(0xffffff)
	colButton2 = rgb(0xe8e8e8)
)

//go:embed assets/logo.png
var logoPNG []byte

func mustIcon(data []byte) *widget.Icon {
	ic, err := widget.NewIcon(data)
	if err != nil {
		panic(err)
	}
	return ic
}

var (
	icChecked   = mustIcon(icons.ToggleCheckBox)
	icUnchecked = mustIcon(icons.ToggleCheckBoxOutlineBlank)
	icPartial   = mustIcon(icons.ToggleIndeterminateCheckBox)
	icRight     = mustIcon(icons.NavigationChevronRight)
	icDown      = mustIcon(icons.NavigationExpandMore)
)

func newTheme() *material.Theme {
	th := material.NewTheme()
	// Only the embedded Go fonts: scanning the system fonts is slow on old PCs.
	th.Shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()))
	th.TextSize = 14
	th.Palette.Fg = colInk
	th.Palette.ContrastBg = colAccent
	return th
}

func loadLogo() paint.ImageOp {
	img, err := png.Decode(bytes.NewReader(logoPNG))
	if err != nil {
		panic(err) // embedded
	}
	return paint.NewImageOp(img)
}

// lbl is a one-line label (truncated with an ellipsis).
func lbl(th *material.Theme, sz unit.Sp, s string, col color.NRGBA) material.LabelStyle {
	l := material.Label(th, sz, s)
	l.Color = col
	l.MaxLines = 1
	return l
}

func bold(l material.LabelStyle) material.LabelStyle {
	l.Font.Weight = font.Bold
	return l
}

// para is a wrapping paragraph.
func para(th *material.Theme, s string, col color.NRGBA) material.LabelStyle {
	l := material.Label(th, th.TextSize, s)
	l.Color = col
	return l
}

func fill(gtx C, col color.NRGBA) D {
	defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, col)
	return D{Size: gtx.Constraints.Min}
}

// bg draws w on a col background.
func bg(gtx C, col color.NRGBA, w layout.Widget) D {
	return layout.Background{}.Layout(gtx, func(gtx C) D { return fill(gtx, col) }, w)
}

func hline(gtx C) D {
	gtx.Constraints.Min = image.Pt(gtx.Constraints.Max.X, gtx.Dp(1))
	return fill(gtx, colLine)
}

func vline(gtx C) D {
	gtx.Constraints.Min = image.Pt(gtx.Dp(1), gtx.Constraints.Max.Y)
	return fill(gtx, colLine)
}

// flexSpace takes the remaining room in a Flex.
var flexSpace = layout.Flexed(1, func(gtx C) D { return D{Size: gtx.Constraints.Min} })

func gap(dp unit.Dp) layout.FlexChild {
	return layout.Rigid(layout.Spacer{Width: dp, Height: dp}.Layout)
}

// badge is a small rounded tag.
func badge(th *material.Theme, s string, col color.NRGBA) layout.Widget {
	return func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				r := gtx.Dp(3)
				defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, r).Push(gtx.Ops).Pop()
				c := col
				c.A = 0x22
				paint.Fill(gtx.Ops, c)
				return D{Size: gtx.Constraints.Min}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.Inset{Left: 6, Right: 6, Top: 1, Bottom: 1}.Layout(gtx, lbl(th, 11, s, col).Layout)
			}))
	}
}

// btn is a button in a Flex; primary ones use the accent colour. Disabled
// buttons ignore clicks and have grey text; the keyboard focus gets an
// outline.
func btn(th *material.Theme, c *widget.Clickable, s string, primary, enabled bool) layout.FlexChild {
	return layout.Rigid(func(gtx C) D {
		b := material.Button(th, c, s)
		b.TextSize, b.CornerRadius = 14, 4
		b.Inset = layout.Inset{Top: 8, Bottom: 8, Left: 16, Right: 16}
		if !primary {
			b.Background, b.Color = colButton2, colInk
		}
		if !enabled {
			gtx = gtx.Disabled()
			b.Color = rgb(0x9e9e9e)
		}
		if !enabled && primary {
			// Gio greys a disabled button's background out, which turns the
			// red muddy: draw the light red ourselves (no clickable, no events).
			return layout.Background{}.Layout(gtx,
				func(gtx C) D {
					r := gtx.Dp(4)
					defer clip.UniformRRect(image.Rectangle{Max: gtx.Constraints.Min}, r).Push(gtx.Ops).Pop()
					return fill(gtx, colLight)
				},
				func(gtx C) D {
					l := lbl(th, 14, s, colWhite)
					l.Font.Weight = font.Medium
					return b.Inset.Layout(gtx, l.Layout)
				})
		}
		border := widget.Border{CornerRadius: 4}
		if gtx.Focused(c) {
			border.Color, border.Width = colBlue, 2
		}
		return border.Layout(gtx, b.Layout)
	})
}

// hrow and vcol are Flex shorthands.
func hrow(gtx C, kids ...layout.FlexChild) D {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx, kids...)
}

func vcol(gtx C, kids ...layout.FlexChild) D {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, kids...)
}

// row is hrow as a widget.
func row(kids ...layout.FlexChild) layout.Widget {
	return func(gtx C) D { return hrow(gtx, kids...) }
}

// focused lays w out with an outline while c has the keyboard focus (Tab
// moves it; Gio draws no indicator itself). Call it inside c.Layout.
func focused(gtx C, c *widget.Clickable, w layout.Widget) D {
	if !gtx.Focused(c) {
		return w(gtx)
	}
	return widget.Border{Color: colBlue, Width: 2, CornerRadius: 3}.Layout(gtx, w)
}

// checkbox draws a (tri-state) checkbox icon in a clickable.
func checkbox(gtx C, c *widget.Clickable, s tri, enabled bool) D {
	if !enabled {
		gtx = gtx.Disabled()
	}
	return c.Layout(gtx, func(gtx C) D {
		return focused(gtx, c, func(gtx C) D { return checkIcon(gtx, s, enabled) })
	})
}

// checkIcon is the icon of a checkbox.
func checkIcon(gtx C, s tri, enabled bool) D {
	ic, col := icUnchecked, colMuted
	switch s {
	case checked:
		ic, col = icChecked, colAccent
	case partial:
		ic, col = icPartial, colAccent
	}
	if !enabled {
		col = rgb(0xb0b0b0)
	}
	sz := gtx.Dp(20)
	gtx.Constraints = layout.Exact(image.Pt(sz, sz))
	return ic.Layout(gtx, col)
}

// fixed lays w out at exactly dp width.
func fixed(dp unit.Dp, w layout.Widget) layout.FlexChild {
	return layout.Rigid(func(gtx C) D {
		px := gtx.Dp(dp)
		gtx.Constraints.Min.X, gtx.Constraints.Max.X = px, px
		return w(gtx)
	})
}

func rightAligned(l material.LabelStyle) layout.Widget {
	l.Alignment = text.End
	return l.Layout
}

// textField is a single-line editor with a border.
func textField(th *material.Theme, ed *widget.Editor, hint string) layout.Widget {
	return func(gtx C) D {
		return widget.Border{Color: rgb(0xbdbdbd), CornerRadius: 4, Width: 1}.Layout(gtx, func(gtx C) D {
			return layout.Inset{Top: 7, Bottom: 7, Left: 8, Right: 8}.Layout(gtx, func(gtx C) D {
				gtx.Constraints.Min.X = gtx.Constraints.Max.X
				e := material.Editor(th, ed, hint)
				e.HintColor = rgb(0x9e9e9e)
				return e.Layout(gtx)
			})
		})
	}
}

// size formats bytes for display: "999 B", "1.0 KB", "412.3 MB".
func size(n int64) string {
	if n > -1024 && n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	v, unit := float64(n)/1024, "KB"
	for _, u := range []string{"MB", "GB"} {
		if math.Abs(v) < 1024 {
			break
		}
		v, unit = v/1024, u
	}
	return fmt.Sprintf("%.1f %s", v, unit)
}

func signedSize(n int64) string {
	if n > 0 {
		return "+" + size(n)
	}
	return size(n)
}

func plural(n int, word string) string {
	if n == 1 {
		return "1 " + word
	}
	if strings.HasSuffix(word, "y") {
		return fmt.Sprintf("%d %sies", n, word[:len(word)-1])
	}
	return fmt.Sprintf("%d %ss", n, word)
}

func date(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2 Jan 2006")
}

// winPath shows a manifest path (forward slashes) the Windows way.
func winPath(s string) string { return strings.ReplaceAll(s, "/", `\`) }
