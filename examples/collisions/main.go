package main

import (
	"image/color"

	"github.com/quasilyte/ge"
	"github.com/quasilyte/ge/gedebug"
	"github.com/quasilyte/ge/input"
	"github.com/quasilyte/ge/physics"
	"github.com/quasilyte/gmath"

	_ "image/png"
)

const (
	ActionLeft input.Action = iota
	ActionRight
	ActionUp
	ActionDown
	ActionSolidToggle
	ActionNextShape
)

func main() {
	ctx := ge.NewContext(ge.ContextConfig{})
	ctx.WindowTitle = "Collisions"
	ctx.WindowWidth = 800
	ctx.WindowHeight = 640

	// Bind controls.
	keymap := input.Keymap{
		ActionLeft:        {input.KeyA},
		ActionRight:       {input.KeyD},
		ActionUp:          {input.KeyW},
		ActionDown:        {input.KeyS},
		ActionSolidToggle: {input.KeySpace},
		ActionNextShape:   {input.KeyEnter},
	}
	inputHandler := ctx.Input.NewHandler(0, keymap)

	if err := ge.RunGame(ctx, &controller{input: inputHandler}); err != nil {
		panic(err)
	}
}

type controller struct {
	input *input.Handler
}

func (c *controller) Init(scene *ge.Scene) {
	p := &player{input: c.input}
	p.body.Pos = gmath.Vec{X: 400, Y: 320}
	scene.AddObject(p)

	{
		var b physics.Body
		o := &obstacle{}
		b.InitCircle(o, 20)
		o.body = b
		o.body.Pos = gmath.Vec{X: 300, Y: 200}
		scene.AddObject(o)
	}

	{
		var b physics.Body
		o := &obstacle{}
		b.InitRotatedRect(o, 100, 20)
		o.body = b
		o.body.Pos = gmath.Vec{X: 450, Y: 400}
		scene.AddObject(o)
	}
	{
		var b physics.Body
		o := &obstacle{}
		b.InitRotatedRect(o, 20, 100)
		o.body = b
		o.body.Pos = gmath.Vec{X: 650, Y: 400}
		scene.AddObject(o)
	}
	{
		var b physics.Body
		o := &obstacle{rotates: true}
		b.InitRotatedRect(o, 160, 60)
		o.body = b
		o.body.Pos = gmath.Vec{X: 500, Y: 200}
		scene.AddObject(o)
	}
}

func (c *controller) Update(float64) {
	// do nothing
}

type player struct {
	body    physics.Body
	scene   *ge.Scene
	aura    *gedebug.BodyAura
	input   *input.Handler
	isSolid bool
}

func (p *player) Init(scene *ge.Scene) {
	p.scene = scene

	p.body.InitCircle(p, 32)
	scene.AddBody(&p.body)

	p.aura = &gedebug.BodyAura{Body: &p.body}
	scene.AddGraphics(p.aura)
}

func (p *player) IsDisposed() bool { return false }

func (p *player) Update(delta float64) {
	if p.input.ActionIsJustPressed(ActionSolidToggle) {
		p.isSolid = !p.isSolid
	}

	if p.input.ActionIsJustPressed(ActionNextShape) {
		if p.body.IsCircle() {
			p.body.InitRotatedRect(p, 64, 40)
		} else if p.body.IsRotatedRect() {
			p.body.InitCircle(p, 32)
		}
	}

	const movementSpeed = 128
	const rotationSpeed = 2
	var velocity gmath.Vec
	if p.input.ActionIsPressed(ActionLeft) {
		switch {
		case p.body.IsCircle():
			velocity.X = -movementSpeed * delta
		case p.body.IsRotatedRect():
			p.body.Rotation -= rotationSpeed * gmath.Rad(delta)
		}
	}
	if p.input.ActionIsPressed(ActionRight) {
		switch {
		case p.body.IsCircle():
			velocity.X = movementSpeed * delta
		case p.body.IsRotatedRect():
			p.body.Rotation += rotationSpeed * gmath.Rad(delta)
		}
	}
	if p.input.ActionIsPressed(ActionUp) {
		switch {
		case p.body.IsCircle():
			velocity.Y = -movementSpeed * delta
		case p.body.IsRotatedRect():
			velocity = gmath.RadToVec(p.body.Rotation).Mulf(movementSpeed * delta)
		}
	}
	if p.input.ActionIsPressed(ActionDown) {
		switch {
		case p.body.IsCircle():
			velocity.Y = movementSpeed * delta
		case p.body.IsRotatedRect():
			velocity = gmath.RadToVec(p.body.Rotation).Mulf(-movementSpeed * delta)
		}
	}

	alpha := uint8(255)
	if !p.isSolid {
		alpha = 150
	}
	collision := p.scene.GetMovementCollision(&p.body, velocity)
	if collision != nil {
		p.aura.Color = color.RGBA{R: 255, G: 150, B: 150, A: alpha}
		if p.isSolid {
			p.body.Pos = p.body.Pos.Add(collision.Normal.Mulf(collision.Depth + 0.1))
		} else {
			p.body.Pos = p.body.Pos.Add(velocity)
		}
	} else {
		p.aura.Color = color.RGBA{R: 150, G: 255, B: 150, A: alpha}
		p.body.Pos = p.body.Pos.Add(velocity)
	}
}

type obstacle struct {
	body    physics.Body
	rotates bool
}

func (o *obstacle) Init(scene *ge.Scene) {
	scene.AddBody(&o.body)
	scene.AddGraphics(&gedebug.BodyAura{Body: &o.body})
}

func (o *obstacle) IsDisposed() bool { return o.body.IsDisposed() }

func (o *obstacle) Update(delta float64) {
	if o.rotates {
		o.body.Rotation += gmath.Rad(delta / 2)
	}
}
