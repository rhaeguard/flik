package main

import (
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type shard struct {
	pos        rl.Vector2
	velocity   rl.Vector2
	life       float32
	color      rl.Color
	radius     float32
	angles     []float32
	polyPoints []rl.Vector2
}

func getPoint(angle float32, eRadius rl.Vector2) rl.Vector2 {
	theta := float64(angle * rl.Deg2rad)

	x := float64(eRadius.X) * math.Cos(theta)
	y := float64(eRadius.Y) * math.Sin(theta)

	return rl.NewVector2(float32(x), float32(y))
}

func getAngles() []float32 {
	n := rand.Intn(4) + 3

	angles := []float32{}

	for i := 0; i < n; i++ {
		angle := rand.Float32() * 355.23
		angles = append(angles, angle)
	}

	slices.Sort(angles)

	return angles
}

func getPolygon(center rl.Vector2, angles []float32, eRadius rl.Vector2) []rl.Vector2 {
	slices.Sort(angles)

	polygonPoints := []rl.Vector2{}

	for _, angle := range angles {
		pt := getPoint(angle, eRadius)
		pt = rl.Vector2Add(pt, center)
		polygonPoints = append(polygonPoints, pt)
	}

	return polygonPoints
}

func NewShard(
	pos rl.Vector2,
	angle float32,
	speed float32,
	life float32,
	radius float32,
	color rl.Color,
) shard {

	angleInRadians := float64(angle * rl.Deg2rad)
	vx := float32(math.Cos(angleInRadians)) * speed
	vy := -float32(math.Sin(angleInRadians)) * speed

	return shard{
		pos:  pos,
		life: life,
		velocity: rl.NewVector2(
			vx, vy,
		),
		color:  color,
		radius: radius,
		angles: getAngles(),
	}
}

func (p *shard) update() {
	p.life -= 0.0167 * 2

	if p.life > 0 {
		p.pos = rl.Vector2Add(p.pos, p.velocity)

		for i := range p.angles {
			p.angles[i] += 10
		}

		eRadius := rl.NewVector2(
			p.radius*1.5,
			p.radius,
		)

		p.polyPoints = getPolygon(p.pos, p.angles, eRadius)
	}
}

func (p *shard) render() {
	if p.life > 0 {
		alpha := 255 * p.life / 2.0

		l := len(p.polyPoints)
		p0 := p.polyPoints[0]
		for i := 1; i < l-1; i++ {
			p1 := p.polyPoints[i%l]
			p2 := p.polyPoints[i+1]
			rl.DrawTriangle(p2, p1, p0, rl.NewColor(p.color.R, p.color.G, p.color.B, uint8(alpha)))
		}
	}
}
