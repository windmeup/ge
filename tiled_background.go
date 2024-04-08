package ge

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	resource "github.com/quasilyte/ebitengine-resource"
	"github.com/quasilyte/ge/tiled"
	"github.com/quasilyte/gmath"
)

type TiledBackground struct {
	Pos Pos

	Visible bool

	ColorScale ColorScale

	Hue gmath.Rad

	disposed bool

	imageCache *imageCache

	combined *ebiten.Image
}

func NewTiledBackground(ctx *Context) *TiledBackground {
	return &TiledBackground{
		Visible:    true,
		ColorScale: defaultColorScale,
		imageCache: &ctx.imageCache,
	}
}

func (bg *TiledBackground) LoadTilesetWithRand(ctx *Context, rand *gmath.Rand, width, height float64, source resource.ImageID, tileset resource.RawID) {
	ts, err := tiled.UnmarshalTileset(ctx.Loader.LoadRaw(tileset).Data)
	if err != nil {
		panic(err)
	}

	spriteSheet := ctx.Loader.LoadImage(source)
	frames := make([]*ebiten.Image, 0, ts.NumTiles)
	for i := 0; i < ts.NumTiles; i++ {
		x := i * int(ts.TileWidth)
		frameRect := image.Rect(x, 0, x+int(ts.TileWidth), int(ts.TileHeight))
		frameImage := spriteSheet.Data.SubImage(frameRect).(*ebiten.Image)
		frames = append(frames, frameImage)
	}

	framePicker := gmath.NewRandPicker[int](rand)
	for i := 0; i < ts.NumTiles; i++ {
		framePicker.AddOption(i, *ts.Tiles[i].Probability)
	}

	combined := ebiten.NewImage(int(width), int(height))
	var op ebiten.DrawImageOptions
	applyColorScale(bg.ColorScale, &op.ColorScale)
	if bg.Hue != 0 {
		op.ColorM.RotateHue(float64(bg.Hue))
	}
	for y := float64(0); y < height; y += ts.TileHeight {
		for x := float64(0); x < width; x += ts.TileWidth {
			offset := gmath.Vec{X: x, Y: y}
			frameIndex := framePicker.Pick()
			img := frames[frameIndex]
			op.GeoM.Reset()
			op.GeoM.Translate(offset.X, offset.Y)
			combined.DrawImage(img, &op)
		}
	}
	bg.combined = combined
}

func (bg *TiledBackground) LoadTileset(ctx *Context, width, height float64, source resource.ImageID, tileset resource.RawID) {
	bg.LoadTilesetWithRand(ctx, &ctx.Rand, width, height, source, tileset)
}

func (bg *TiledBackground) IsDisposed() bool {
	return bg.disposed
}

func (bg *TiledBackground) Dispose() {
	bg.disposed = true
}

func (bg *TiledBackground) DrawPartial(screen *ebiten.Image, section gmath.Rect) {
	bg.DrawPartialWithOffset(screen, section, gmath.Vec{})
}

func (bg *TiledBackground) DrawPartialWithOffset(screen *ebiten.Image, section gmath.Rect, offset gmath.Vec) {
	if !bg.Visible {
		return
	}

	// TODO: handle pos too?

	pMin := gmath.Vec{X: math.Round(section.Min.X), Y: math.Round(section.Min.Y)}
	pMax := gmath.Vec{X: math.Round(section.Max.X), Y: math.Round(section.Max.Y)}
	unsafeSrc := toUnsafeImage(bg.combined)
	unsafeSubImage := bg.imageCache.UnsafeImageForSubImage()
	unsafeSubImage.original = unsafeSrc
	unsafeSubImage.bounds = image.Rectangle{
		Min: image.Point{X: int(pMin.X), Y: int(pMin.Y)},
		Max: image.Point{X: int(pMax.X), Y: int(pMax.Y)},
	}
	unsafeSubImage.image = unsafeSrc.image
	srcImage := toEbitenImage(unsafeSubImage)
	var op ebiten.DrawImageOptions
	op.GeoM.Translate(pMin.X, pMin.Y)
	op.GeoM.Translate(offset.X, offset.Y)
	screen.DrawImage(srcImage, &op)
}

func (bg *TiledBackground) Draw(screen *ebiten.Image) {
	if !bg.Visible {
		return
	}

	pos := bg.Pos.Resolve()

	var op ebiten.DrawImageOptions
	op.GeoM.Translate(pos.X, pos.Y)
	screen.DrawImage(bg.combined, &op)
}

func (bg *TiledBackground) DrawImage(img *ebiten.Image, options *ebiten.DrawImageOptions) {
	bg.combined.DrawImage(img, options)
}
