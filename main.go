package main

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var BG_COLOR = rl.NewColor(139, 212, 195, 255)
var STONE_COLLISION_SHARD_COLOR = rl.NewColor(255, 192, 113, 255)
var AIM_VECTOR_COLOR = rl.Yellow

type GameStatus = int8
type ActionEnum = int8
type Player = int8

// magic numbers
var VelocityDampingFactor float32
var VelocityThresholdToStop float32
var MaxPullLengthAllowed float32
var MaxPushVelocityAllowed float32
var StoneRadius float32
var FontSize float32

// shards and particles
var MaxParticleSpeed float32
var MaxShardRadius float32

const (
	Uninitialized GameStatus = iota
	Initialized   GameStatus = iota
	Stopped       GameStatus = iota
	GameOver      GameStatus = iota
	// action enums
	NoAction   ActionEnum = iota
	StoneAimed ActionEnum = iota
	StoneHit   ActionEnum = iota
	// player turn
	PlayerOne Player = iota
	PlayerTwo Player = iota
)

type PlayerSettings struct {
	label          string
	primaryColor   rl.Color
	outerRingColor rl.Color
	lifeColor      rl.Color
	rocketColor    rl.Color
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

func (c *Window) GetScreenDiagonal() float32 {
	w, h := c.GetScreenDimensions()
	res := math.Sqrt(float64(w*w) + float64(h*h))
	return float32(res)
}

type Stone struct {
	id       int
	pos      rl.Vector2
	velocity rl.Vector2
	mass     float32
	radius   float32
	life     float32
	isDead   bool
	playerId Player
}

func newStone(stoneId int, x, y float64, radius, mass float32, playerId Player) Stone {
	return Stone{
		id:       stoneId,
		pos:      rl.NewVector2(float32(x), float32(y)),
		velocity: rl.NewVector2(0, 0),
		mass:     mass,
		radius:   radius,
		life:     100,
		isDead:   false,
		playerId: playerId,
	}
}

func dimWhite(alpha uint8) color.RGBA {
	return rl.NewColor(255, 255, 255, alpha)
}

type Game struct {
	status                         GameStatus
	lastTimeUpdated                float64
	totalTimeRunning               float32
	stones                         []Stone
	selectedStone                  *Stone
	selectedStoneRotAnimationAngle float32
	hitStoneMoving                 *Stone
	stoneHitPosition               rl.Vector2
	action                         ActionEnum
	allParticles                   []particle
	allShards                      []shard
	score                          map[Player]uint8
	playerTurn                     Player
	stonesAreStill                 bool
	playerSettings                 map[Player]PlayerSettings
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

func generateStones(window *Window) []Stone {
	stones := []Stone{}

	screenWidth, screenHeight := window.GetScreenDimensions()

	width := float64(screenWidth)
	height := float64(screenHeight)

	f1 := generateFormation()
	f2 := generateFormation()

	ids := 0

	for x := 1; x <= 3; x += 1 {
		for y := 1; y <= 4; y += 1 {
			h := height * float64(y) * 0.2

			pos := 3*(y-1) + (x - 1)

			if f1[pos] {
				w1 := width * float64(x) * 0.125
				stones = append(stones, newStone(ids, w1, h, StoneRadius, 1, PlayerOne))
				ids++
			}

			if f2[pos] {
				w2 := width*float64(x)*0.125 + width*0.5
				stones = append(stones, newStone(ids, w2, h, StoneRadius, 1, PlayerTwo))
				ids++
			}
		}
	}

	return stones
}

func newGame() Game {
	return Game{
		status:                         Uninitialized,
		lastTimeUpdated:                0.0,
		totalTimeRunning:               0.0,
		stones:                         []Stone{},
		selectedStone:                  nil,
		selectedStoneRotAnimationAngle: 0.0,
		hitStoneMoving:                 nil,
		action:                         NoAction,
		allParticles:                   []particle{},
		allShards:                      []shard{},
		score: map[Player]uint8{
			PlayerOne: 6,
			PlayerTwo: 6,
		},
		playerTurn:     PlayerOne,
		stonesAreStill: true,
		playerSettings: map[Player]PlayerSettings{
			PlayerOne: {
				label:          "you",
				primaryColor:   rl.NewColor(55, 113, 142, 255),
				outerRingColor: rl.NewColor(37, 78, 112, 255),
				lifeColor:      rl.NewColor(255, 250, 255, 255),
				rocketColor:    rl.SkyBlue,
			},
			PlayerTwo: {
				label:          "cpu",
				primaryColor:   rl.NewColor(133, 90, 92, 255),
				outerRingColor: rl.NewColor(102, 16, 31, 255),
				lifeColor:      rl.NewColor(255, 250, 255, 255),
				rocketColor:    rl.NewColor(129, 13, 32, 255),
			},
		},
	}
}

func (g *Game) init(w *Window) {
	g.stones = generateStones(w)
	g.status = Initialized
}

func main() {
	window := Window{
		fullscreen: true,
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

	defer rl.CloseWindow()

	// do not allow objects to penetrate into each other
	// this algorithm basically identifies the penetration depth
	// and moves the objects back half the distance in the direction they are coming in.
	resolvePenetrationDepth := func(a, b *Stone) {
		direction := rl.Vector2Subtract(a.pos, b.pos)
		penetrationDepth := (a.radius + b.radius) - rl.Vector2Length(direction)

		direction = rl.Vector2Scale(rl.Vector2Normalize(direction), penetrationDepth/2)

		a.pos = rl.Vector2Add(a.pos, direction)
		b.pos = rl.Vector2Add(b.pos, rl.Vector2Negate(direction))
	}

	resolveCollision := func(a, b *Stone) {
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

	calcVelocity := func(s *Stone) {
		s.velocity = rl.Vector2Scale(s.velocity, VelocityDampingFactor)
		if rl.Vector2Length(s.velocity) < VelocityThresholdToStop {
			s.velocity.X = 0
			s.velocity.Y = 0
		}
	}

	type collisionPair struct {
		a, b           *Stone
		collisionPoint rl.Vector2
		magnitude      float32
	}
	update := func() {
		seen := map[string]bool{}
		collidingPairs := []collisionPair{}
		for i := range game.stones {
			a := &game.stones[i]
			if a.isDead {
				continue
			}
			for j := range game.stones {
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
					intersection := rl.Vector2Scale(rl.Vector2Add(a.pos, b.pos), 0.5)

					combinedVelocity := rl.Vector2Add(a.velocity, b.velocity)

					collisionMagnitude := (2 * rl.Vector2Length(combinedVelocity)) / MaxPushVelocityAllowed

					collidingPairs = append(collidingPairs, collisionPair{a, b, intersection, collisionMagnitude})
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
					p.collisionPoint,
					float32(3.6*float32(i)),
					MaxParticleSpeed*rand.Float32(),
					p.magnitude,
					MaxShardRadius*(rand.Float32()+0.5),
					STONE_COLLISION_SHARD_COLOR,
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
			strength := rl.Vector2Distance(game.stoneHitPosition, game.selectedStone.pos)
			strength = rl.Clamp(MaxPullLengthAllowed, 0, strength)
			game.selectedStoneRotAnimationAngle += rl.GetFrameTime() * 3 * strength
		}

		{ // creates the shards at the position of the dead stone
			for _, ix := range newlyDeadStonesIx {
				stone := &game.stones[ix]

				shardColor := game.playerSettings[stone.playerId].primaryColor

				for i := 0.0; i < 300; i += 0.5 {
					part := NewShard(
						stone.pos,
						float32(3.6*float32(i)),
						MaxParticleSpeed*rand.Float32(),
						2,
						MaxShardRadius*(rand.Float32()+0.5),
						shardColor,
						false,
					)

					game.allShards = append(game.allShards, part)
				}
			}
		}

		if game.action == StoneHit {
			// find the diff between the selected stone and where the mouse is
			diff := rl.Vector2Subtract(game.selectedStone.pos, game.stoneHitPosition)
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
				rocketColor := game.playerSettings[stone.playerId].rocketColor

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

		for i := range game.allParticles {
			game.allParticles[i].update()
		}

		for i := range game.allShards {
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
			scorePlayerOne := 0
			scorePlayerTwo := 0

			for _, stone := range game.stones {
				if stone.isDead {
					continue
				}

				if stone.playerId == PlayerOne {
					scorePlayerOne += 1
				}

				if stone.playerId == PlayerTwo {
					scorePlayerTwo += 1
				}
			}

			game.score[PlayerOne] = uint8(scorePlayerOne)
			game.score[PlayerTwo] = uint8(scorePlayerTwo)

			if scorePlayerOne*scorePlayerTwo == 0 {
				game.status = GameOver
				game.playerTurn = PlayerOne
			}
		}

		game.lastTimeUpdated = rl.GetTime()
		game.totalTimeRunning += rl.GetFrameTime()
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
		game.stoneHitPosition = rl.GetMousePosition()

		if rl.IsMouseButtonDown(rl.MouseButtonRight) && game.stonesAreStill {
			for i, stone := range game.stones {
				if stone.isDead {
					continue
				}
				if game.playerTurn == stone.playerId && rl.CheckCollisionPointCircle(game.stoneHitPosition, stone.pos, stone.radius) {
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

	type searchPair struct {
		actor, target *Stone
		score         float32
	}

	compareSearchPairs := func(p1, p2 searchPair) int {
		return cmp.Compare(p2.score, p1.score)
	}

	/// sufficiently "smart" AI
	/// searches the options based on:
	/// - proximity
	/// - life state of the target stone
	/// - life state of the hitting stone
	/// - whether own stone will be hit in the process
	/// - whether stone will richochet
	cpuSearchBestOption := func() (*Stone, *Stone) {
		searchPairs := []searchPair{}
		for i := range game.stones {
			actor := &game.stones[i]
			if actor.isDead || actor.playerId == PlayerOne { // TODO: we should have better ways to indicate the opponent
				continue
			}
			for j := range game.stones {
				if i == j {
					continue
				}

				target := &game.stones[j]
				if target.isDead || target.playerId == PlayerTwo { // TODO: better way to indicate the attacking player
					continue
				}

				searchPairs = append(searchPairs, searchPair{
					actor:  actor,
					target: target,
				})
			}
		}

		screenDiagonalSize := window.GetScreenDiagonal()

		for pi := range searchPairs {
			pair := &(searchPairs[pi])
			actor, target := pair.actor, pair.target

			ssOrigin := rl.Vector2Subtract(actor.pos, target.pos)
			angle := 2 * rl.Vector2LineAngle(rl.Vector2Normalize(ssOrigin), rl.NewVector2(1, 0))

			aTop := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, -actor.radius), -angle), actor.pos)
			aBottom := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, actor.radius), -angle), actor.pos)

			tTop := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, -target.radius), -angle), target.pos)
			tBottom := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, target.radius), -angle), target.pos)

			hitsOwn := false
			richochets := false

			for i := range game.stones {
				stone := &game.stones[i]
				if stone.isDead || (stone == actor || stone == target) {
					continue
				}
				// line 1 check aTop - tTop
				if rl.CheckCollisionCircleLine(stone.pos, stone.radius, aTop, tTop) {
					hitsOwn = stone.playerId == PlayerTwo // better way needed?
					richochets = true
				}
				// line 2 check aBottom - tBottom
				if rl.CheckCollisionCircleLine(stone.pos, stone.radius, aBottom, tBottom) {
					hitsOwn = stone.playerId == PlayerTwo
					richochets = true
				}

				// line 3 check center to center
				if rl.CheckCollisionCircleLine(stone.pos, stone.radius, actor.pos, target.pos) {
					hitsOwn = stone.playerId == PlayerTwo
					richochets = true
				}

				if hitsOwn && richochets {
					break
				}
			}

			distance := rl.Vector2Distance(actor.pos, target.pos) / screenDiagonalSize

			pair.score -= distance
			if hitsOwn {
				pair.score += -1
			}

			if richochets {
				pair.score += -0.5
			}

			if actor.life <= 5 {
				pair.score += -0.5
			}

			if target.life <= 10 {
				pair.score += 1
			}
		}

		slices.SortFunc(searchPairs, compareSearchPairs)

		if len(searchPairs) > 0 {
			pair := searchPairs[0]
			return pair.actor, pair.target
		}
		return nil, nil
	}

	handleCpuMove := func() {
		if !game.stonesAreStill || game.status == GameOver {
			return
		}

		actor, target := cpuSearchBestOption()

		if actor == nil || target == nil {
			return
		}

		game.selectedStone = actor
		game.action = StoneHit
		game.stoneHitPosition = rl.Vector2Add(actor.pos, rl.Vector2Negate(rl.Vector2Subtract(target.pos, actor.pos)))
	}

	drawStone := func(s *Stone) {
		if s.isDead {
			return
		}
		playerSettings := game.playerSettings[s.playerId]

		rl.DrawCircleV(s.pos, s.radius, playerSettings.primaryColor)

		// the outer/border ring
		rl.DrawRing(
			s.pos,
			s.radius*0.8,
			s.radius*1.01,
			0.0,
			360.0,
			0,
			playerSettings.outerRingColor,
		)

		if game.stonesAreStill && s.playerId == game.playerTurn && game.playerTurn == PlayerOne {
			// the "active player" ring
			rl.DrawRing(
				s.pos,
				s.radius*1.1,
				s.radius*1.4,
				0.0,
				360.0,
				0,
				dimWhite(50),
			)

			if game.selectedStone == s {
				rl.DrawRing(
					s.pos,
					s.radius*1.1,
					s.radius*1.4,
					0.0+game.selectedStoneRotAnimationAngle,
					40.0+game.selectedStoneRotAnimationAngle,
					0,
					dimWhite(100),
				)
			} else {
				rl.DrawRing(
					s.pos,
					s.radius*1.1,
					s.radius*1.4,
					0.0+game.totalTimeRunning*10,
					40.0+game.totalTimeRunning*10,
					0,
					dimWhite(100),
				)
			}

		}

		rl.DrawRing(
			s.pos,
			s.radius*0.5,
			s.radius*0.8,
			0.0,
			360.0*float32(s.life)/100,
			0,
			playerSettings.lifeColor,
		)
	}

	drawScore := func(screenWidth, screenHeight float32) {
		color := dimWhite(60)
		labelP1 := game.playerSettings[PlayerOne].label
		labelP2 := game.playerSettings[PlayerTwo].label

		defaultFont := rl.GetFontDefault()

		p1Score := fmt.Sprintf("0%d", game.score[PlayerOne])
		p2Score := fmt.Sprintf("0%d", game.score[PlayerTwo])

		measuredSize := rl.MeasureTextEx(defaultFont, "00", FontSize, FontSize/10)

		offsetX := (screenWidth/2 - measuredSize.X) / 2
		offsetY := (screenHeight - measuredSize.Y) / 2

		rl.DrawTextEx(defaultFont, p1Score, rl.NewVector2(offsetX, offsetY), FontSize, FontSize/10, color)
		rl.DrawTextEx(defaultFont, p2Score, rl.NewVector2(screenWidth-offsetX-measuredSize.X, offsetY), FontSize, FontSize/10, color)

		labelP1Width := rl.MeasureTextEx(defaultFont, labelP1, FontSize/3, FontSize/30).X
		labelP2Width := rl.MeasureTextEx(defaultFont, labelP2, FontSize/3, FontSize/30).X
		p1OffsetX := (screenWidth/2 - labelP1Width) / 2
		p2OffsetX := ((screenWidth/2 - labelP2Width) / 2) + screenWidth/2
		labelsOffsetY := offsetY + measuredSize.Y*0.8

		rl.DrawTextEx(defaultFont, labelP1, rl.NewVector2(p1OffsetX, labelsOffsetY), FontSize/3, FontSize/30, color)
		rl.DrawTextEx(defaultFont, labelP2, rl.NewVector2(p2OffsetX, labelsOffsetY), FontSize/3, FontSize/30, color)
	}

	draw := func() {
		screenWidth, screenHeight := window.GetScreenDimensions()

		if game.status == GameOver {
			whoWon := game.playerSettings[PlayerOne].label
			if game.score[PlayerOne] == 0 {
				whoWon = game.playerSettings[PlayerTwo].label
			}
			whoWon = fmt.Sprintf("%s won!", whoWon)
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
			drawScore(float32(screenWidth), float32(screenHeight))

			rl.DrawLineEx(
				rl.NewVector2(float32(screenWidth/2), 0),
				rl.NewVector2(float32(screenWidth/2), float32(screenHeight)),
				float32(screenWidth)/256,
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
				game.stoneHitPosition,
				game.selectedStone.pos,
				3.0,
				AIM_VECTOR_COLOR,
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
			FontSize = float32(screenWidth) * 0.25
			// init
			game = newGame()
			game.init(&window)
		}

		if rl.IsKeyDown(rl.KeyS) {
			if game.status == Stopped {
				game.status = Initialized
			} else {
				game.status = Stopped
			}
		}

		game.stonesAreStill = areStonesStill()

		if game.status != Stopped {
			if game.playerTurn == PlayerOne {
				handleMouseMove()
			} else {
				handleCpuMove()
			}
			update()
		}

		rl.BeginDrawing()

		// draw background
		rl.ClearBackground(BG_COLOR)

		draw()

		rl.EndDrawing()
	}
}
