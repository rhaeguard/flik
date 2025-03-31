package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Particle struct {
	pos      rl.Vector2
	velocity rl.Vector2
	life     float32
	color    rl.Color
	radius   float32
}

func NewParticle(
	pos rl.Vector2,
	angle float32,
	speed float32,
	life float32,
	radius float32,
	color rl.Color,
) Particle {

	angleInRadians := float64(angle * rl.Deg2rad)
	vx := float32(math.Cos(angleInRadians)) * speed
	vy := -float32(math.Sin(angleInRadians)) * speed

	return Particle{
		pos:  pos,
		life: life,
		velocity: rl.NewVector2(
			vx, vy,
		),
		color:  color,
		radius: radius,
	}
}

func (p *Particle) update() {
	p.life -= 0.0167 * 2

	if p.life > 0 {
		p.pos = rl.Vector2Add(p.pos, p.velocity)
	}
}

func (p *Particle) render() {
	if p.life > 0 {
		alpha := 255 * p.life / 2.0
		rl.DrawCircleV(p.pos, p.radius, rl.NewColor(p.color.R, p.color.G, p.color.B, uint8(alpha)))
	}
}
