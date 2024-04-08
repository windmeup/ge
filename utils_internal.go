package ge

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func applyColorScale(c ColorScale, ec *ebiten.ColorScale) {
	if c == defaultColorScale {
		return
	}
	ec.ScaleWithColorScale(c.toEbitenColorScale())
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
