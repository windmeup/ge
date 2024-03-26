package ge

import (
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/quasilyte/gmath"
	"golang.org/x/image/font"
)

type AlignVertical uint8

const (
	AlignVerticalTop AlignVertical = iota
	AlignVerticalCenter
	AlignVerticalBottom
)

type AlignHorizontal uint8

const (
	AlignHorizontalLeft AlignHorizontal = iota
	AlignHorizontalCenter
	AlignHorizontalRight
)

type GrowVertical uint8

const (
	GrowVerticalDown GrowVertical = iota
	GrowVerticalUp
	GrowVerticalBoth
	GrowVerticalNone
)

type GrowHorizontal uint8

const (
	GrowHorizontalRight GrowHorizontal = iota
	GrowHorizontalLeft
	GrowHorizontalBoth
	GrowHorizontalNone
)

type Label struct {
	Text string

	colorScale       ColorScale
	ebitenColorScale ebiten.ColorScale

	Pos    Pos
	Width  float64
	Height float64

	Visible bool

	AlignVertical   AlignVertical
	AlignHorizontal AlignHorizontal
	GrowVertical    GrowVertical
	GrowHorizontal  GrowHorizontal

	face       font.Face
	capHeight  float64
	lineHeight float64
}

func NewLabel(ff font.Face) *Label {
	m := ff.Metrics()
	capHeight := math.Abs(float64(m.CapHeight.Floor()))
	lineHeight := float64(m.Height.Floor())
	label := &Label{
		face:             ff,
		capHeight:        capHeight,
		lineHeight:       lineHeight,
		colorScale:       defaultColorScale,
		ebitenColorScale: defaultColorScale.toEbitenColorScale(),
		Visible:          true,
		AlignHorizontal:  AlignHorizontalLeft,
		AlignVertical:    AlignVerticalTop,
	}
	return label
}

func (l *Label) IsDisposed() bool {
	return l.face == nil
}

func (l *Label) Dispose() {
	l.face = nil
}

func (l *Label) GetColorScale() ColorScale {
	return l.colorScale
}

func (l *Label) GetAlpha() float32 {
	return l.colorScale.A
}

func (l *Label) SetAlpha(a float32) {
	if l.colorScale.A == a {
		return
	}
	l.colorScale.A = a
	l.ebitenColorScale = l.colorScale.toEbitenColorScale()
}

func (l *Label) SetColorScaleRGBA(r, g, b, a uint8) {
	var scale ColorScale
	scale.SetRGBA(r, g, b, a)
	l.SetColorScale(scale)
}

func (l *Label) SetColorScale(colorScale ColorScale) {
	if l.colorScale == colorScale {
		return
	}
	l.colorScale = colorScale
	l.ebitenColorScale = l.colorScale.toEbitenColorScale()
}

func (l *Label) DrawWithOffset(screen *ebiten.Image, offset gmath.Vec) {
	if !l.Visible || l.Text == "" {
		return
	}

	pos := l.Pos.Resolve()

	// Adjust the pos, since "dot position" (baseline) is not a top-left corner.
	pos.Y += l.capHeight

	numLines := strings.Count(l.Text, "\n") + 1

	var containerRect gmath.Rect
	bounds, _ := font.BoundString(l.face, l.Text) // assume bounds is well-formed and its width and height is integer
	boundsWidth := float64((bounds.Max.X - bounds.Min.X).Floor())
	boundsHeight := float64((bounds.Max.Y - bounds.Min.Y).Floor())
	if l.Width == 0 && l.Height == 0 {
		// Auto-sized container.
		switch l.GrowHorizontal {
		case GrowHorizontalRight:
			containerRect.Min.X = pos.X
			containerRect.Max.X = pos.X + boundsWidth
		case GrowHorizontalLeft:
			containerRect.Min.X = pos.X - boundsWidth
			containerRect.Max.X = pos.X
			pos.X -= boundsWidth
		case GrowHorizontalBoth:
			containerRect.Min.X = pos.X - boundsWidth/2
			containerRect.Max.X = pos.X + boundsWidth/2
			pos.X -= boundsWidth / 2
		}
		switch l.GrowVertical {
		case GrowVerticalDown:
			containerRect.Min.Y = pos.Y
			containerRect.Max.Y = pos.Y + boundsHeight
		case GrowVerticalUp:
			containerRect.Min.Y = pos.Y - boundsHeight
			containerRect.Max.Y = pos.Y
			pos.Y -= boundsHeight
		case GrowVerticalBoth:
			containerRect.Min.Y = pos.Y - boundsHeight/2
			containerRect.Max.Y = pos.Y + boundsHeight/2
			pos.Y -= boundsHeight / 2
		}
	} else {
		containerRect = gmath.Rect{
			Min: pos,
			Max: pos.Add(gmath.Vec{X: l.Width, Y: l.Height}),
		}
		if delta := boundsWidth - l.Width; delta > 0 {
			switch l.GrowHorizontal {
			case GrowHorizontalRight:
				containerRect.Max.X += delta
			case GrowHorizontalLeft:
				containerRect.Min.X -= delta
			case GrowHorizontalBoth:
				containerRect.Min.X -= delta / 2
				containerRect.Max.X += delta / 2
			case GrowHorizontalNone:
				// Do nothing.
			}
		}
		if delta := boundsHeight - l.Height; delta > 0 {
			switch l.GrowVertical {
			case GrowVerticalDown:
				containerRect.Min.Y += delta
			case GrowVerticalUp:
				containerRect.Min.Y -= delta
				pos.Y -= delta
			case GrowVerticalBoth:
				containerRect.Min.Y -= delta / 2
				containerRect.Max.Y += delta / 2
				pos.Y -= delta / 2
			case GrowVerticalNone:
				// Do nothing.
			}
		}
	}

	switch l.AlignVertical {
	case AlignVerticalTop:
		// Do nothing.
	case AlignVerticalCenter:
		pos.Y += (containerRect.Height() - l.estimateHeight(numLines)) / 2
	case AlignVerticalBottom:
		pos.Y += containerRect.Height() - l.estimateHeight(numLines)
	}

	var drawOptions ebiten.DrawImageOptions
	drawOptions.ColorScale = l.ebitenColorScale
	drawOptions.Filter = ebiten.FilterLinear

	if l.AlignHorizontal == AlignHorizontalLeft {
		drawOptions.GeoM.Translate(math.Round(pos.X), math.Round(pos.Y))
		drawOptions.GeoM.Translate(offset.X, offset.Y)
		text.DrawWithOptions(screen, l.Text, l.face, &drawOptions)
		return
	}

	textRemaining := l.Text
	offsetY := 0.0
	for {
		nextLine := strings.IndexByte(textRemaining, '\n')
		lineText := textRemaining
		if nextLine != -1 {
			lineText = textRemaining[:nextLine]
			textRemaining = textRemaining[nextLine+len("\n"):]
		}
		lineBounds, _ := font.BoundString(l.face, l.Text) // assume bounds is well-formed and its width and height is integer
		lineBoundsWidth := float64((lineBounds.Max.X - lineBounds.Min.X).Floor())
		offsetX := 0.0
		switch l.AlignHorizontal {
		case AlignHorizontalCenter:
			offsetX = (containerRect.Width() - lineBoundsWidth) / 2
		case AlignHorizontalRight:
			offsetX = containerRect.Width() - lineBoundsWidth
		}
		drawOptions.GeoM.Reset()
		drawOptions.GeoM.Translate(math.Round(pos.X+offsetX), math.Round(pos.Y+offsetY))
		drawOptions.GeoM.Translate(offset.X, offset.Y)
		text.DrawWithOptions(screen, lineText, l.face, &drawOptions)
		if nextLine == -1 {
			break
		}
		offsetY += l.lineHeight
	}
}

func (l *Label) Draw(screen *ebiten.Image) {
	l.DrawWithOffset(screen, gmath.Vec{})
}

func (l *Label) estimateHeight(numLines int) float64 {
	estimatedHeight := l.capHeight
	if numLines >= 2 {
		estimatedHeight += (float64(numLines) - 1) * l.lineHeight
	}
	return estimatedHeight
}
