package ge

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/ge/audio"
	"github.com/quasilyte/ge/gemath"
	"github.com/quasilyte/ge/loader"
)

type Context struct {
	Loader   *loader.Cache
	Renderer *Renderer

	Input Input
	Audio audio.System

	Rand gemath.Rand

	CurrentScene *Scene

	OnCriticalError func(err error)

	WindowTitle  string
	WindowWidth  float64
	WindowHeight float64
}

func NewContext() *Context {
	ctx := &Context{
		WindowTitle: "GE Game",
	}
	ctx.Loader = loader.NewCache()
	ctx.Renderer = NewRenderer()
	ctx.Rand.SetSeed(0)
	ctx.Input.init()
	ctx.Audio.Init(ctx.Loader)
	ctx.OnCriticalError = func(err error) {
		panic(err)
	}
	ctx.Loader.WAVDecoder = &ctx.Audio
	return ctx
}

func (ctx *Context) NewScene(name string, controller SceneController) *Scene {
	scene := newScene()
	scene.Name = name
	scene.context = ctx

	scene.controller = controller
	controller.Init(scene)
	scene.addQueuedObjects()

	return scene
}

func (ctx *Context) LoadSprite(path string) *Sprite {
	return NewSprite(ctx.Loader.GetImage(path))
}

func (ctx *Context) Draw(screen *ebiten.Image) {
	ctx.CurrentScene.graphics = ctx.Renderer.Draw(screen, ctx.CurrentScene.graphics)
}

func (ctx *Context) WindowRect() gemath.Rect {
	return gemath.Rect{
		Max: gemath.Vec{X: ctx.WindowWidth, Y: ctx.WindowHeight},
	}
}
