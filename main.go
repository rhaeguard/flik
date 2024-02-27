package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var bgColor = rl.NewColor(54, 89, 74, 255)

type stone struct {
	pos      rl.Vector2
	color    rl.Color
	velocity rl.Vector2
	mass     float32
	radius   float32
}

func main() {
	lastTimeUpdated := 0.0

	mouseLeftStart := rl.NewVector2(0, 0)

	rl.InitWindow(1000, 600, "flik")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	stones := []stone{
		{
			pos:      rl.NewVector2(300, 300),
			color:    rl.Black,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},

		{
			pos:      rl.NewVector2(680, 320),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},

		{
			pos:      rl.NewVector2(730, 280),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},

		{
			pos:      rl.NewVector2(730, 320),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},

		{
			pos:      rl.NewVector2(760, 260),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},
		{
			pos:      rl.NewVector2(760, 300),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},
		{
			pos:      rl.NewVector2(760, 340),
			color:    rl.White,
			velocity: rl.NewVector2(0, 0),
			mass:     1,
			radius:   15,
		},
	}

	resolveCollision := func(a, b *stone) {
		// 1. find unit normal and unit tangent
		normalV := rl.NewVector2(a.pos.X-b.pos.X, a.pos.Y-b.pos.Y)
		unitNormal := rl.Vector2Scale(normalV, 1/rl.Vector2Length(normalV))
		unitTangent := rl.NewVector2(-unitNormal.Y, unitNormal.X)

		// 2. initial velocity vectors

		// 3.
		van := rl.Vector2DotProduct(unitNormal, a.velocity)
		vbn := rl.Vector2DotProduct(unitNormal, b.velocity)
		vat := rl.Vector2DotProduct(unitTangent, a.velocity)
		vbt := rl.Vector2DotProduct(unitTangent, b.velocity)

		// 4. find new tangential velocities - after collision
		vatp := vat
		vbtp := vbt

		// 5. find new normal velocities
		masses := a.mass + b.mass
		vanp := (van*(a.mass-b.mass) + 2*b.mass*vbn) / masses
		vbnp := (vbn*(b.mass-a.mass) + 2*a.mass*van) / masses

		// 6. scalar normal and tangential velocities to vectors
		vanpV := rl.Vector2Scale(unitNormal, vanp)
		vbnpV := rl.Vector2Scale(unitNormal, vbnp)
		vatpV := rl.Vector2Scale(unitTangent, vatp)
		vbtpV := rl.Vector2Scale(unitTangent, vbtp)

		// 7. find the final velocity vectors
		vaV := rl.Vector2Add(vanpV, vatpV)
		vbV := rl.Vector2Add(vbnpV, vbtpV)

		fmt.Printf("%p\n", a)
		fmt.Printf("%-v\n", a.velocity)
		fmt.Printf("%-v\n---------\n", vaV)

		a.velocity = vaV
		b.velocity = vbV
	}

	doStonesCollide := func(a, b *stone) bool {
		return rl.CheckCollisionCircles(a.pos, a.radius, b.pos, b.radius)
	}

	clamp := func(v rl.Vector2) rl.Vector2 {
		maxValue := float32(250)
		x := v.X
		y := v.Y

		if v.X > maxValue {
			x = maxValue
		}
		if v.X < -maxValue {
			x = -maxValue
		}

		if v.Y > maxValue {
			y = maxValue
		}
		if v.Y < -maxValue {
			y = -maxValue
		}

		return rl.NewVector2(x, y)
	}

	handleMouseMove := func() {
		fmt.Printf("%-v\n", stones[0].velocity)
		hasStopped := stones[0].velocity == rl.NewVector2(0, 0)

		if rl.IsMouseButtonDown(rl.MouseButtonRight) && hasStopped {
			(&stones[0]).pos = rl.GetMousePosition()
		}

		if rl.IsMouseButtonDown(rl.MouseButtonLeft) && hasStopped {
			mouseLeftStart = rl.GetMousePosition()
		}

		if !rl.IsMouseButtonDown(rl.MouseButtonLeft) && mouseLeftStart != rl.NewVector2(0, 0) {
			diff := rl.Vector2Subtract(stones[0].pos, mouseLeftStart)
			diff = clamp(diff)

			diff = rl.Vector2Scale(diff, 15.0)
			v := rl.Vector2Scale(diff, 1/250.0)
			fmt.Printf("%-v\n", v)
			(&stones[0]).velocity = v
			mouseLeftStart = rl.NewVector2(0, 0)
		}
	}

	calcVelocity := func(s *stone) {
		velocityDampingFactor := float32(0.987)
		s.velocity = rl.Vector2Scale(s.velocity, velocityDampingFactor)
		if rl.Vector2Length(s.velocity) < 0.07 {
			s.velocity = rl.NewVector2(0, 0)
		}
	}

	update := func() {
		if rl.GetTime()-lastTimeUpdated > 0.1667 {
			seen := map[string]bool{}
			for i := 0; i < len(stones); i++ {
				for j := 0; j < len(stones); j++ {
					a := &stones[i]
					b := &stones[j]
					key := fmt.Sprintf("%p-%p", a, b)
					if _, ok := seen[key]; i == j || ok {
						if ok {
							fmt.Printf("skipping because already seen\n")
						}
						continue
					}
					if doStonesCollide(a, b) {
						seen[fmt.Sprintf("%p-%p", a, b)] = true
						seen[fmt.Sprintf("%p-%p", b, a)] = true
						resolveCollision(a, b)
					}
					if len(seen) > 0 {
						fmt.Printf("%-v\n", seen)
					}
				}
			}
			for i := range stones {
				stone := &stones[i]
				stone.pos = rl.Vector2Add(stone.pos, stone.velocity)
				calcVelocity(stone)
			}

		}
	}

	draw := func() {
		for _, stone := range stones {
			rl.DrawCircleV(stone.pos, stone.radius, stone.color)
		}
	}

	for !rl.WindowShouldClose() {
		handleMouseMove()

		update()

		rl.BeginDrawing()

		rl.ClearBackground(bgColor)

		draw()

		rl.EndDrawing()
	}
}
