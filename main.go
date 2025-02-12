package main

import (
	"fmt"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var bgColor = rl.NewColor(139, 212, 195, 255)
var teal = rl.NewColor(80, 114, 137, 255)
var tealDarker = rl.NewColor(28, 71, 99, 255)
var pinkish = rl.NewColor(255, 211, 193, 255)

type gameStatus = int8
type actionEnum = int8

// magic numbers
var VelocityDampingFactor float32
var VelocityThresholdToStop float32
var MaxPullLengthAllowed float32
var MaxPushVelocityAllowed float32
var StoneRadius float32

// shards and particles
var MaxParticleSpeed float32
var MaxShardRadius float32

const (
	Uninitialized gameStatus = iota
	Initialized   gameStatus = iota
	Stopped       gameStatus = iota
	// action enums
	NoAction   actionEnum = iota
	StoneAimed actionEnum = iota
	StoneHit   actionEnum = iota
)

type stone struct {
	pos      rl.Vector2
	color    rl.Color
	velocity rl.Vector2
	mass     float32
	radius   float32
	isDead   bool
}

func newStone(w, h float64, color rl.Color, radius float32) stone {
	return stone{
		pos:      rl.NewVector2(float32(w), float32(h)),
		color:    color,
		velocity: rl.NewVector2(0, 0),
		mass:     1,
		radius:   radius,
		isDead:   false,
	}
}

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

type score struct {
	teal uint8
	pink uint8
}

type game struct {
	status          gameStatus
	lastTimeUpdated float64
	stones          []stone
	selectedStone   *stone
	hitStoneMoving  *stone
	action          actionEnum
	startupConfig   startupConfig
	allParticles    []particle
	allShards       []shard
	score           *score
	colorTurn       rl.Color
}

// generates a random formation of 6 stones in a 3x4 matrix
func generateFormation() [12]bool {
	a := [12]bool{
		true, true, true, true, true, true,
		false, false, false, false, false, false,
	}
	rand.Shuffle(12, func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}

func generateStones(screenWidth, screenHeight int32) []stone {
	stones := []stone{}

	width := float64(screenWidth)
	height := float64(screenHeight)

	f1 := generateFormation()
	f2 := generateFormation()

	for x := 1; x <= 3; x += 1 {
		for y := 1; y <= 4; y += 1 {
			h := height * float64(y) * 0.2

			pos := 3*(y-1) + (x - 1)

			if f1[pos] {
				w1 := width * float64(x) * 0.125
				stones = append(stones, newStone(w1, h, teal, StoneRadius))
			}

			if f2[pos] {
				w2 := width*float64(x)*0.125 + width*0.5
				stones = append(stones, newStone(w2, h, pinkish, StoneRadius))
			}
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
		hitStoneMoving:  nil,
		action:          NoAction,
		startupConfig: startupConfig{
			fullscreen: true,
			width:      640,
			height:     360,
		},
		allParticles: []particle{},
		allShards:    []shard{},
		score: &score{
			teal: 6,
			pink: 6,
		},
		colorTurn: teal,
	}

	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	if game.startupConfig.fullscreen {
		rl.InitWindow(0, 0, "flik")
		rl.ToggleFullscreen()
	} else {
		rl.InitWindow(game.startupConfig.width, game.startupConfig.height, "flik")
	}

	rl.SetTargetFPS(60)

	defer rl.CloseWindow()

	resolveCollision := func(a, b *stone) {
		// 1. find unit normal and unit tangent
		unitNormal := rl.Vector2Normalize(rl.Vector2Subtract(a.pos, b.pos))
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
			s.velocity.X = 0
			s.velocity.Y = 0
		}
	}

	type pair struct {
		a, b *stone
		p    rl.Vector2
		life float32
	}
	update := func() {
		seen := map[string]bool{}
		collidingPairs := []pair{}
		for i := 0; i < len(game.stones); i++ {
			a := &game.stones[i]
			if a.isDead {
				continue
			}
			for j := 0; j < len(game.stones); j++ {
				if i == j {
					continue
				}

				b := &game.stones[j]

				if b.isDead {
					continue
				}

				key := fmt.Sprintf("%d-%d", i, j)
				if _, ok := seen[key]; ok {
					continue
				}

				if rl.CheckCollisionCircles(a.pos, a.radius, b.pos, b.radius) {
					seen[fmt.Sprintf("%d-%d", i, j)] = true
					seen[fmt.Sprintf("%d-%d", j, i)] = true
					intersection := circleIntersectionPoint(a, b)

					combinedVelocity := rl.Vector2Add(a.velocity, b.velocity)

					life := (2 * rl.Vector2Length(combinedVelocity)) / MaxPushVelocityAllowed

					collidingPairs = append(collidingPairs, pair{a, b, intersection, life})
				}
			}
		}

		for _, p := range collidingPairs {
			resolveCollision(p.a, p.b)
			game.hitStoneMoving = nil

			for i := 0.0; i < 100; i += 0.5 {
				// TODO: shard size should depend on the screen size
				part := NewShard(
					p.p,
					float32(3.6*float32(i)),
					MaxParticleSpeed*rand.Float32(),
					p.life,
					MaxShardRadius*(rand.Float32()+0.5),
					rl.NewColor(255, 192, 113, 255),
					true,
				)

				game.allShards = append(game.allShards, part)
			}
		}

		screenWidth, screenHeight := game.startupConfig.GetScreenDimensions()
		screenRect := rl.NewRectangle(0, 0, float32(screenWidth), float32(screenHeight))

		newlyDeadStonesIx := []int{}

		for i := range game.stones {
			stone := &game.stones[i]
			if stone.isDead {
				continue
			}
			stone.pos = rl.Vector2Add(stone.pos, stone.velocity)
			calcVelocity(stone)

			if !rl.CheckCollisionPointRec(stone.pos, screenRect) {
				stone.isDead = true
				newlyDeadStonesIx = append(newlyDeadStonesIx, i)
				if game.hitStoneMoving == stone {
					game.hitStoneMoving = nil
				}
			}
		}

		{
			for _, ix := range newlyDeadStonesIx {
				stone := &game.stones[ix]

				for i := 0.0; i < 300; i += 0.5 {
					part := NewShard(
						stone.pos,
						float32(3.6*float32(i)),
						MaxParticleSpeed*rand.Float32(),
						2,
						MaxShardRadius*(rand.Float32()+0.5),
						stone.color,
						false,
					)

					game.allShards = append(game.allShards, part)
				}
			}
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
			game.hitStoneMoving = game.selectedStone
			game.selectedStone = nil
			if game.colorTurn == teal {
				game.colorTurn = pinkish
			} else {
				game.colorTurn = teal
			}
		}

		{
			// rocket exhaust
			stone := game.hitStoneMoving

			if stone != nil {
				rocketColor := rl.Red

				if stone.color == teal {
					rocketColor = rl.SkyBlue
				}

				generalAngle := (rl.Vector2Angle(
					rl.Vector2Normalize(stone.velocity),
					rl.NewVector2(1, 0),
				) * rl.Rad2deg) - 180

				life := 0.3 * (rl.Vector2Length(stone.velocity)) / 15

				for i := 0; i < 25; i++ {
					angle := generalAngle + float32((rand.Intn(20) - 10))
					part := NewParticle(
						stone.pos,
						float32(angle),
						MaxParticleSpeed*rand.Float32(),
						life,
						stone.radius*1.02,
						rocketColor,
					)

					game.allParticles = append(game.allParticles, part)
				}

				if rl.Vector2Length(stone.velocity) == 0 {
					game.hitStoneMoving = nil
				}

			}
		}

		for i := 0; i < len(game.allParticles); i++ {
			game.allParticles[i].update()
		}

		for i := 0; i < len(game.allShards); i++ {
			game.allShards[i].update()
		}

		{
			// filter out the dead particles
			newAllParticles := []particle{}
			for _, p := range game.allParticles {
				if p.life > 0 {
					newAllParticles = append(newAllParticles, p)
				}
			}

			game.allParticles = newAllParticles
		}

		{
			// filter out the dead shards
			newShards := []shard{}
			for _, p := range game.allShards {
				if p.life > 0 {
					newShards = append(newShards, p)
				}
			}

			game.allShards = newShards
		}

		{
			// scoring calculation
			scoreTeal := 0
			scorePink := 0

			for _, stone := range game.stones {
				if stone.isDead {
					continue
				}

				if stone.color == teal {
					scoreTeal += 1
				}

				if stone.color == pinkish {
					scorePink += 1
				}
			}

			game.score.pink = uint8(scorePink)
			game.score.teal = uint8(scoreTeal)
		}

		game.lastTimeUpdated = rl.GetTime()
	}

	areStonesStill := func() bool {
		for _, stone := range game.stones {
			if stone.isDead {
				continue
			}
			if rl.Vector2Length(stone.velocity) != 0 {
				return false
			}
		}
		return true
	}

	handleMouseMove := func() {
		mousePos := rl.GetMousePosition()
		hasStopped := areStonesStill()

		if rl.IsMouseButtonDown(rl.MouseButtonRight) && hasStopped {
			for i, stone := range game.stones {
				if game.colorTurn == stone.color && rl.CheckCollisionPointCircle(mousePos, stone.pos, stone.radius) {
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
		if s.isDead {
			return
		}
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

		if s.color == game.colorTurn {
			rl.DrawRing(
				s.pos,
				s.radius*0.2,
				s.radius*0.5,
				0.0,
				360.0,
				0,
				rl.Red,
			)
		}

	}

	draw := func() {
		screenWidth, screenHeight := game.startupConfig.GetScreenDimensions()

		// draw background
		rl.ClearBackground(bgColor)

		measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), "00", 600, 0)
		width := (screenWidth/2 - int32(measuredSize.X)) / 2
		height := (screenHeight - int32(measuredSize.Y)) / 2
		rl.DrawText(fmt.Sprintf("0%d", game.score.teal), width, height, 600, rl.NewColor(255, 255, 255, 60))
		rl.DrawText("teal", width+int32(measuredSize.X)/4, height+4*int32(measuredSize.Y)/5, 200, rl.NewColor(255, 255, 255, 60))

		rl.DrawText(fmt.Sprintf("0%d", game.score.pink), screenWidth-width-int32(measuredSize.X), height, 600, rl.NewColor(255, 255, 255, 60))
		rl.DrawText("pink", screenWidth-width-int32(measuredSize.X)+int32(measuredSize.X)/4, height+4*int32(measuredSize.Y)/5, 200, rl.NewColor(255, 255, 255, 60))

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

		{
			// draw particles
			for _, p := range game.allParticles {
				rl.BeginBlendMode(rl.BlendAdditive)
				p.render()
				rl.EndBlendMode()
			}
		}

		{
			// draw shards
			for _, p := range game.allShards {
				mode := rl.BlendAlpha
				if p.fade {
					mode = rl.BlendAdditive
				}
				rl.BeginBlendMode(mode)
				p.render()
				rl.EndBlendMode()
			}
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
			// magic numbers
			// ratio is computed based on 2560 x 1440
			VelocityDampingFactor = 0.987
			VelocityThresholdToStop = float32(screenWidth) / 36_000
			MaxPullLengthAllowed = 0.1 * float32(screenWidth)
			MaxPushVelocityAllowed = 0.008 * float32(screenWidth)
			MaxParticleSpeed = 0.008 * float32(screenWidth)
			MaxShardRadius = float32(screenWidth) / 256
			StoneRadius = float32(screenHeight) * 0.069
			// init
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
