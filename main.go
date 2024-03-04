package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var bgColor = rl.NewColor(139, 212, 195, 255)
var teal = rl.NewColor(80, 114, 137, 255)
var tealDarker = rl.NewColor(28, 71, 99, 255)
var pinkish = rl.NewColor(255, 211, 193, 255)

type stone struct {
	pos      rl.Vector2
	color    rl.Color
	velocity rl.Vector2
	mass     float32
	radius   float32
}

func newStone(w, h float64, color rl.Color, radius float32) stone {
	return stone{
		pos:      rl.NewVector2(float32(w), float32(h)),
		color:    color,
		velocity: rl.NewVector2(0, 0),
		mass:     1,
		radius:   radius,
	}
}

type gameStatus = int8
type actionEnum = int8

const (
	Uninitialized gameStatus = iota
	Initialized   gameStatus = iota
	Stopped       gameStatus = iota
	// action enums
	NoAction   actionEnum = iota
	StoneAimed actionEnum = iota
	StoneHit   actionEnum = iota
	// magic numbers
	VelocityDampingFactor   float32 = 0.987
	VelocityThresholdToStop float32 = 0.07
	MaxPullLengthAllowed    float32 = 250.0 // TODO: might need to be adjusted based on the dynamic screen size
	MaxPushVelocityAllowed  float32 = 15.0
)

type startupConfig struct {
	fullscreen bool
	width      int32
	height     int32
}

func (c *startupConfig) GetScreenDimensions() (int32, int32) {
	if c.fullscreen {
		w := int32(rl.GetScreenWidth())
		h := int32(rl.GetScreenHeight())
		return w, h
	} else {
		return c.width, c.height
	}
}

type game struct {
	status          gameStatus
	lastTimeUpdated float64
	stones          []stone
	selectedStone   *stone
	action          actionEnum
	startupConfig   startupConfig
}

func generateStones(screenWidth, screenHeight int32) []stone {
	stones := []stone{}

	width := float64(screenWidth)
	height := float64(screenHeight)

	stoneRadius := float32(screenHeight) * 0.069

	// left side gen
	for h := 0.25 * height; h < height; h += 0.25 * height {
		for w := 0.1 * width; w < 0.5*width; w += 0.2 * width {
			stones = append(stones, newStone(w, h, teal, stoneRadius))
		}
	}

	// right side gen
	for h := 0.25 * height; h < height; h += 0.25 * height {
		for w := 0.9 * width; w > 0.5*width; w -= 0.2 * width {
			stones = append(stones, newStone(w, h, pinkish, stoneRadius))
		}
	}

	return stones
}

func main() {
	game := game{
		status:          Uninitialized,
		lastTimeUpdated: 0.0,
		stones:          []stone{},
		selectedStone:   nil,
		action:          NoAction,
		startupConfig: startupConfig{
			fullscreen: true,
			width:      800,
			height:     600,
		},
	}

	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	if game.startupConfig.fullscreen {
		rl.InitWindow(0, 0, "flik")
		rl.ToggleFullscreen()
	} else {
		rl.InitWindow(game.startupConfig.width, game.startupConfig.height, "flik")
	}

	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	resolveCollision := func(a, b *stone) {
		// 1. find unit normal and unit tangent
		normalV := rl.NewVector2(a.pos.X-b.pos.X, a.pos.Y-b.pos.Y)
		unitNormal := rl.Vector2Scale(normalV, 1/rl.Vector2Length(normalV))
		unitTangent := rl.NewVector2(-unitNormal.Y, unitNormal.X)

		// 2. initial velocity vectors
		// everything stays as-is

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

		a.velocity = vaV
		b.velocity = vbV
	}

	calcVelocity := func(s *stone) {
		s.velocity = rl.Vector2Scale(s.velocity, VelocityDampingFactor)
		if rl.Vector2Length(s.velocity) < VelocityThresholdToStop {
			s.velocity = rl.NewVector2(0, 0)
		}
	}

	type pair struct {
		a, b *stone
	}
	update := func() {
		seen := map[string]bool{}
		collidingPairs := []pair{}
		for i := 0; i < len(game.stones); i++ {
			for j := 0; j < len(game.stones); j++ {
				if i == j {
					continue
				}

				a := &game.stones[i]
				b := &game.stones[j]
				key := fmt.Sprintf("%d-%d", i, j)
				if _, ok := seen[key]; ok {
					continue
				}

				if rl.CheckCollisionCircles(a.pos, a.radius, b.pos, b.radius) {
					seen[fmt.Sprintf("%d-%d", i, j)] = true
					seen[fmt.Sprintf("%d-%d", j, i)] = true
					collidingPairs = append(collidingPairs, pair{a, b})
				}
			}
		}

		for _, p := range collidingPairs {
			resolveCollision(p.a, p.b)
		}

		for i := range game.stones {
			stone := &game.stones[i]
			stone.pos = rl.Vector2Add(stone.pos, stone.velocity)
			calcVelocity(stone)
		}

		if game.action == StoneHit {
			// find the diff between the selected stone and where the mouse is
			diff := rl.Vector2Subtract(game.selectedStone.pos, rl.GetMousePosition())
			// find the length of the diff vector
			length := rl.Vector2Length(diff)
			// make sure the length is bounded
			length = rl.Clamp(length, 0, MaxPullLengthAllowed)
			// the max speed we allow is 15,
			// so we calculate the speed based on the distance from the selected stone
			speed := MaxPushVelocityAllowed * (length / MaxPullLengthAllowed)
			// normalize the diff vector
			// scale it up based on the speed
			v := rl.Vector2Scale(rl.Vector2Normalize(diff), speed)

			game.selectedStone.velocity = v

			game.action = NoAction
			game.selectedStone = nil
		}

		game.lastTimeUpdated = rl.GetTime()
	}

	handleMouseMove := func() {
		mousePos := rl.GetMousePosition()
		hasStopped := game.selectedStone == nil || game.selectedStone.velocity == rl.NewVector2(0, 0)

		if rl.IsMouseButtonDown(rl.MouseButtonRight) && hasStopped {
			for i, stone := range game.stones {
				if rl.CheckCollisionPointCircle(mousePos, stone.pos, stone.radius) {
					game.selectedStone = &game.stones[i]
					game.action = StoneAimed
					break
				}
			}
		}

		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && game.action == StoneAimed {
			game.action = StoneHit
		}
	}

	drawStone := func(s *stone) {
		rl.DrawCircleV(s.pos, s.radius, s.color)

		rl.DrawRing(
			s.pos,
			s.radius*0.8,
			s.radius*1.01,
			0.0,
			360.0,
			0,
			tealDarker,
		)
	}

	draw := func() {
		screenWidth, screenHeight := game.startupConfig.GetScreenDimensions()

		// draw background
		rl.ClearBackground(bgColor)

		// measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), "00", 600, 0)
		// width := (SCREEN_WIDTH/2 - int32(measuredSize.X)) / 2
		// height := (SCREEN_HEIGHT - int32(measuredSize.Y)) / 2
		// rl.DrawText("01", width, height, 600, rl.NewColor(255, 255, 255, 60))
		// rl.DrawText("07", SCREEN_WIDTH-width-int32(measuredSize.X), height, 600, rl.NewColor(255, 255, 255, 60))

		rl.DrawLineEx(
			rl.NewVector2(float32(screenWidth/2), 0),
			rl.NewVector2(float32(screenWidth/2), float32(screenHeight)),
			10.0,
			rl.NewColor(255, 255, 255, 125),
		)

		for i := 0; i < len(game.stones); i++ {
			stone := &(game.stones[i])
			drawStone(stone)
		}

		if game.action == StoneAimed {
			rl.DrawLineEx(
				rl.GetMousePosition(),
				game.selectedStone.pos,
				3.0,
				rl.Yellow,
			)

			{
				// TODO: this should not really be calculated here
				mouseLeftStart := rl.GetMousePosition()

				diff := rl.Vector2Subtract(game.selectedStone.pos, mouseLeftStart)

				length := rl.Vector2Length(diff)
				length = rl.Clamp(length, 0, MaxPullLengthAllowed)
				normalizedSpeed := length / MaxPullLengthAllowed
				// given the normalized speed, calculate the angle
				angle := normalizedSpeed * 360.0

				rl.DrawRing(
					game.selectedStone.pos,
					game.selectedStone.radius,
					game.selectedStone.radius*1.5,
					0.0,
					angle,
					0,
					rl.NewColor(255, 255, 255, 60),
				)
			}
		}
	}

	for !rl.WindowShouldClose() {
		if game.status == Uninitialized {
			screenWidth, screenHeight := game.startupConfig.GetScreenDimensions()
			game.stones = generateStones(screenWidth, screenHeight)
			game.status = Initialized
		}

		handleMouseMove()

		update()

		rl.BeginDrawing()

		draw()

		rl.EndDrawing()
	}
}
