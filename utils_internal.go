package ge

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
)

func applyColorScale(c ColorScale, ec *ebiten.ColorScale) {
	if c == defaultColorScale {
		return
	}
	ec.ScaleWithColorScale(c.toEbitenColorScale())
}

func applyColorScale2(c ColorScale, m *colorm.ColorM) {
	if c == defaultColorScale {
		return
	}
	m.Scale(float64(c.R), float64(c.G), float64(c.B), float64(c.A))
}

func assignColors(vertices []ebiten.Vertex, c ColorScale) {
	colorR := float32(c.R)
	colorG := float32(c.G)
	colorB := float32(c.B)
	colorA := float32(c.A)
	for i := range vertices {
		v := &vertices[i]
		v.ColorR = colorR
		v.ColorG = colorG
		v.ColorB = colorB
		v.ColorA = colorA
	}
}
