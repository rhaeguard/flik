package main

import (
	"fmt"
	"image/color"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var bgColor = rl.NewColor(139, 212, 195, 255)
var teal = rl.NewColor(80, 114, 137, 255)
var tealDarker = rl.NewColor(28, 71, 99, 255)
var pinkish = rl.NewColor(255, 211, 193, 255)
var shardCollisionColor = rl.NewColor(255, 192, 113, 255)

type gameStatus = int8
type actionEnum = int8
type player = int8

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
	GameOver      gameStatus = iota
	// action enums
	NoAction   actionEnum = iota
	StoneAimed actionEnum = iota
	StoneHit   actionEnum = iota
	// player turn
	PlayerOne player = iota
	PlayerTwo player = iota
)

type stone struct {
	pos      rl.Vector2
	color    rl.Color
	velocity rl.Vector2
	mass     float32
	radius   float32
	life     float32
	isDead   bool
	playerId player
}

func dimWhite(alpha uint8) color.RGBA {
	return rl.NewColor(255, 255, 255, alpha)
}

func newStone(w, h float64, color rl.Color, radius, mass float32, p player) stone {
	return stone{
		pos:      rl.NewVector2(float32(w), float32(h)),
		color:    color,
		velocity: rl.NewVector2(0, 0),
		mass:     mass,
		radius:   radius,
		life:     100,
		isDead:   false,
		playerId: p,
	}
}

type Window struct {
	fullscreen bool
	width      int32
	height     int32
}

func (c *Window) GetScreenDimensions() (int32, int32) {
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

type Game struct {
	status                         gameStatus
	lastTimeUpdated                float64
	stones                         []stone
	selectedStone                  *stone
	selectedStoneRotAnimationAngle float32
	hitStoneMoving                 *stone
	action                         actionEnum
	allParticles                   []particle
	allShards                      []shard
	score                          score
	playerTurn                     player
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

func generateStones(window *Window) []stone {
	stones := []stone{}

	screenWidth, screenHeight := window.GetScreenDimensions()

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
				stones = append(stones, newStone(w1, h, teal, StoneRadius, 1, PlayerOne))
			}

			if f2[pos] {
				w2 := width*float64(x)*0.125 + width*0.5
				stones = append(stones, newStone(w2, h, pinkish, StoneRadius, 1, PlayerTwo))
			}
		}
	}

	return stones
}

func newGame() Game {
	return Game{
		status:                         Uninitialized,
		lastTimeUpdated:                0.0,
		stones:                         []stone{},
		selectedStone:                  nil,
		selectedStoneRotAnimationAngle: 0.0,
		hitStoneMoving:                 nil,
		action:                         NoAction,
		allParticles:                   []particle{},
		allShards:                      []shard{},
		score: score{
			teal: 6,
			pink: 6,
		},
		playerTurn: PlayerOne,
	}
}

func (g *Game) init(w *Window) {
	g.stones = generateStones(w)
	g.status = Initialized
}

func main() {
	window := Window{
		fullscreen: !true,
		width:      1280,
		height:     720,
	}
	game := newGame()

	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	if window.fullscreen {
		rl.InitWindow(0, 0, "flik")
		rl.ToggleFullscreen()
	} else {
		rl.InitWindow(window.width, window.height, "flik")
	}

	rl.SetTargetFPS(60)

	camera := rl.NewCamera2D(rl.NewVector2(0, 0), rl.NewVector2(0, 0), 0, 1)

	defer rl.CloseWindow()

	// do not allow objects to penetrate into each other
	// this algorithm basically identifies the penetration depth
	// and moves the objects back half the distance in the direction they are coming in.
	resolvePenetrationDepth := func(a, b *stone) {
		direction := rl.Vector2Subtract(a.pos, b.pos)
		penetrationDepth := (a.radius + b.radius) - rl.Vector2Length(direction)

		direction = rl.Vector2Scale(rl.Vector2Normalize(direction), penetrationDepth/2)

		a.pos = rl.Vector2Add(a.pos, direction)
		b.pos = rl.Vector2Add(b.pos, rl.Vector2Negate(direction))
	}

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
			speedDiff := rl.Vector2Length(rl.Vector2Subtract(p.a.velocity, p.b.velocity))
			aIsFaster := rl.Vector2Length(p.a.velocity) > rl.Vector2Length(p.b.velocity)
			amount := rl.Clamp(speedDiff, 0, MaxPushVelocityAllowed) * 2

			resolvePenetrationDepth(p.a, p.b)
			resolveCollision(p.a, p.b)
			game.hitStoneMoving = nil

			if aIsFaster {
				p.b.life -= amount
				p.a.life -= amount * 0.2
			} else {
				p.a.life -= amount
				p.b.life -= amount * 0.2
			}

			for i := 0.0; i < 100; i += 0.5 {
				// TODO: shard size should depend on the screen size
				part := NewShard(
					p.p,
					float32(3.6*float32(i)),
					MaxParticleSpeed*rand.Float32(),
					p.life,
					MaxShardRadius*(rand.Float32()+0.5),
					shardCollisionColor,
					true,
				)

				game.allShards = append(game.allShards, part)
			}
		}

		screenWidth, screenHeight := window.GetScreenDimensions()
		screenRect := rl.NewRectangle(0, 0, float32(screenWidth), float32(screenHeight))

		newlyDeadStonesIx := []int{}

		for i := range game.stones {
			stone := &game.stones[i]
			if stone.isDead {
				continue
			}
			stone.pos = rl.Vector2Add(stone.pos, stone.velocity)
			calcVelocity(stone)

			if !rl.CheckCollisionPointRec(stone.pos, screenRect) || stone.life <= 0 {
				stone.isDead = true
				newlyDeadStonesIx = append(newlyDeadStonesIx, i)
				if game.hitStoneMoving == stone {
					game.hitStoneMoving = nil
				}
			}
		}

		if game.selectedStone != nil {
			strength := rl.Vector2Distance(rl.GetMousePosition(), game.selectedStone.pos)
			strength = rl.Clamp(MaxPullLengthAllowed, 0, strength) * 3
			game.selectedStoneRotAnimationAngle += rl.GetFrameTime() * strength
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
			game.selectedStoneRotAnimationAngle = 0
			if game.playerTurn == PlayerOne {
				game.playerTurn = PlayerTwo
			} else {
				game.playerTurn = PlayerOne
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

		if game.status == Initialized {
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

			if game.score.pink*game.score.teal == 0 {
				game.status = GameOver
			}
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
				if stone.isDead {
					continue
				}
				if game.playerTurn == stone.playerId && rl.CheckCollisionPointCircle(mousePos, stone.pos, stone.radius) {
					game.selectedStone = &game.stones[i]
					game.action = StoneAimed
					break
				}
			}
		}

		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && game.action == StoneAimed {
			game.action = StoneHit
		}

		if game.status == GameOver && rl.IsKeyDown(rl.KeySpace) {
			game.status = Uninitialized
		}
	}

	drawStone := func(s *stone) {
		if s.isDead {
			return
		}
		rl.DrawCircleV(s.pos, s.radius, s.color)

		// the outer/border ring
		rl.DrawRing(
			s.pos,
			s.radius*0.8,
			s.radius*1.01,
			0.0,
			360.0,
			0,
			tealDarker,
		)

		if s.playerId == game.playerTurn {
			// the "active player" ring
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

		rl.DrawRing(
			s.pos,
			s.radius*0.5,
			s.radius*0.8,
			0.0,
			360.0*float32(s.life)/100,
			0,
			rl.Green,
		)

		if game.selectedStone == s {
			// this section draws the spinning wheel
			// when the player is aiming
			rl.DrawRing(
				s.pos,
				s.radius*1.1,
				s.radius*1.5,
				0.0,
				360.0,
				0,
				dimWhite(50),
			)

			rl.DrawRing(
				s.pos,
				s.radius*1.1,
				s.radius*1.5,
				0.0+game.selectedStoneRotAnimationAngle,
				40.0+game.selectedStoneRotAnimationAngle,
				0,
				dimWhite(100),
			)
		}
	}

	drawScore := func(screenWidth, screenHeight int32) {
		dimmedWhiteColor := dimWhite(60)

		measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), "00", 600, 0)

		width := (screenWidth/2 - int32(measuredSize.X)) / 2
		height := (screenHeight - int32(measuredSize.Y)) / 2

		rl.DrawText(fmt.Sprintf("0%d", game.score.teal), width, height, 600, dimmedWhiteColor)
		rl.DrawText("teal", width+int32(measuredSize.X)/4, height+4*int32(measuredSize.Y)/5, 200, dimmedWhiteColor)

		rl.DrawText(fmt.Sprintf("0%d", game.score.pink), screenWidth-width-int32(measuredSize.X), height, 600, dimmedWhiteColor)
		rl.DrawText("pink", screenWidth-width-int32(measuredSize.X)+int32(measuredSize.X)/4, height+4*int32(measuredSize.Y)/5, 200, dimmedWhiteColor)
	}

	draw := func() {
		screenWidth, screenHeight := window.GetScreenDimensions()

		if game.status == GameOver {
			whoWon := "teal won!"
			if game.score.teal == 0 {
				whoWon = "pink won!"
			}
			measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), whoWon, 200, 10)
			w := (float32(screenWidth) - measuredSize.X) / 2
			h := (float32(screenHeight) - measuredSize.Y) / 2

			rl.DrawTextEx(
				rl.GetFontDefault(),
				whoWon,
				rl.NewVector2(w, h),
				200,
				10,
				dimWhite(60),
			)

			message2 := rl.MeasureTextEx(rl.GetFontDefault(), "press space to restart", 50, 10)
			w = (float32(screenWidth) - message2.X) / 2
			h = h + measuredSize.Y*1.5
			rl.DrawTextEx(
				rl.GetFontDefault(),
				"press space to restart",
				rl.NewVector2(w, h),
				50,
				10,
				dimWhite(60),
			)
		} else {
			drawScore(screenWidth, screenHeight)

			rl.DrawLineEx(
				rl.NewVector2(float32(screenWidth/2), 0),
				rl.NewVector2(float32(screenWidth/2), float32(screenHeight)),
				10.0,
				dimWhite(125),
			)

			for i := range game.stones {
				stone := &(game.stones[i])
				drawStone(stone)
			}
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
		}
	}

	for !rl.WindowShouldClose() {
		if game.status == Uninitialized {
			screenWidth, screenHeight := window.GetScreenDimensions()
			// magic numbers
			// ratio is computed based on 2560 x 1440
			VelocityDampingFactor = 0.987
			VelocityThresholdToStop = float32(screenWidth) / 6_000
			MaxPullLengthAllowed = 0.1 * float32(screenWidth)
			MaxPushVelocityAllowed = 0.008 * float32(screenWidth)
			MaxParticleSpeed = 0.008 * float32(screenWidth)
			MaxShardRadius = float32(screenWidth) / 256
			StoneRadius = float32(screenHeight) * 0.06
			// init
			game = newGame()
			game.init(&window)
		}
		handleMouseMove()

		update()

		rl.BeginDrawing()

		// draw background
		rl.ClearBackground(bgColor)

		rl.BeginMode2D(camera)

		draw()

		rl.EndMode2D()

		rl.EndDrawing()
	}
}
