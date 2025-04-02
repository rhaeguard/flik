package main

import (
	"fmt"
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LevelStatus = int8
type ActionEnum = int8
type Player = int8

const (
	Uninitialized LevelStatus = iota
	Initialized   LevelStatus = iota
	Stopped       LevelStatus = iota
	Finished      LevelStatus = iota
	// action enums
	NoAction   ActionEnum = iota
	StoneAimed ActionEnum = iota
	StoneHit   ActionEnum = iota
	// player turn
	PlayerOne Player = iota
	PlayerTwo Player = iota
)

// level is a scene
// it will have sublevels
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

func newStone(stoneId int, x, y float32, radius, mass float32, playerId Player) Stone {
	return Stone{
		id:       stoneId,
		pos:      rl.NewVector2(x, y),
		velocity: rl.NewVector2(0, 0),
		mass:     mass,
		radius:   radius,
		life:     100,
		isDead:   false,
		playerId: playerId,
	}
}

type PlayerSettings struct {
	label          string
	primaryColor   rl.Color
	outerRingColor rl.Color
	lifeColor      rl.Color
	rocketColor    rl.Color
}

type LevelSettings struct {
	backgroundColor rl.Color
}

type Level struct {
	status                         LevelStatus
	lastTimeUpdated                float64
	totalTimeRunning               float32
	stones                         []Stone
	selectedStone                  *Stone
	selectedStoneRotAnimationAngle float32
	hitStoneMoving                 *Stone
	aimVectorStart                 rl.Vector2
	aimVectorForwardExtensionEnd   rl.Vector2
	action                         ActionEnum
	allParticles                   []Particle
	allShards                      []Shard
	score                          map[Player]uint8
	playerTurn                     Player
	stonesAreStill                 bool
	playerSettings                 map[Player]PlayerSettings
	levelSettings                  LevelSettings
}

func newLevel(levelSettings LevelSettings) Level {
	return Level{
		status:                         Uninitialized,
		lastTimeUpdated:                0.0,
		totalTimeRunning:               0.0,
		stones:                         []Stone{},
		selectedStone:                  nil,
		selectedStoneRotAnimationAngle: 0.0,
		hitStoneMoving:                 nil,
		action:                         NoAction,
		allParticles:                   []Particle{},
		allShards:                      []Shard{},
		stonesAreStill:                 true,
		score: map[Player]uint8{
			PlayerOne: 0,
			PlayerTwo: 0,
		},
		playerTurn: PlayerOne,
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
		levelSettings: levelSettings,
	}
}

func (level *Level) init(window *Window) {
	level.stones = generateStones(window)
	level.status = Initialized
}

// generates a random formation of 6 stones in a 3x4 matrix
func generateFormation() [12]bool {
	a := [12]bool{
		!true, !true, !true, !true, true, true,
		false, false, false, false, false, false,
	}
	rand.Shuffle(12, func(i, j int) { a[i], a[j] = a[j], a[i] })
	return a
}

func generateStones(window *Window) []Stone {
	stones := []Stone{}

	screenWidth, screenHeight := window.GetScreenDimensions()

	f1 := generateFormation()
	f2 := generateFormation()

	ids := 0

	for x := 1; x <= 3; x += 1 {
		for y := 1; y <= 4; y += 1 {
			h := screenHeight * float32(y) * 0.2

			pos := 3*(y-1) + (x - 1)

			if f1[pos] {
				w1 := screenWidth * float32(x) * 0.125
				stones = append(stones, newStone(ids, w1, h, StoneRadius, 1, PlayerOne))
				ids++
			}

			if f2[pos] {
				w2 := screenWidth*float32(x)*0.125 + screenWidth*0.5
				stones = append(stones, newStone(ids, w2, h, StoneRadius, 1, PlayerTwo))
				ids++
			}
		}
	}

	return stones
}

// do not allow objects to penetrate into each other
// this algorithm basically identifies the penetration depth
// and moves the objects back half the distance in the direction they are coming in.
func resolvePenetrationDepth(a, b *Stone) {
	direction := rl.Vector2Subtract(a.pos, b.pos)
	penetrationDepth := (a.radius + b.radius) - rl.Vector2Length(direction)

	direction = rl.Vector2Scale(rl.Vector2Normalize(direction), penetrationDepth/2)

	a.pos = rl.Vector2Add(a.pos, direction)
	b.pos = rl.Vector2Add(b.pos, rl.Vector2Negate(direction))
}

func resolveCollision(a, b *Stone) {
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

func calcVelocity(s *Stone) {
	s.velocity = rl.Vector2Scale(s.velocity, VelocityDampingFactor)
	if rl.Vector2Length(s.velocity) < VelocityThresholdToStop {
		s.velocity.X = 0
		s.velocity.Y = 0
	}
}

func (level *Level) checkStonesForMovements() {
	for _, stone := range level.stones {
		if stone.isDead {
			continue
		}
		if rl.Vector2Length(stone.velocity) != 0 {
			level.stonesAreStill = false
			return
		}
	}
	level.stonesAreStill = true
}

type collisionPair struct {
	a, b           *Stone
	collisionPoint rl.Vector2
	magnitude      float32
}

func (level *Level) update(window *Window) {
	collidingPairs := []collisionPair{}
	allStonesCount := len(level.stones)
	for i := range allStonesCount {
		a := &level.stones[i]
		if a.isDead {
			continue
		}
		for j := i + 1; j < allStonesCount; j++ {
			b := &level.stones[j]

			if b.isDead {
				continue
			}

			if rl.CheckCollisionCircles(a.pos, a.radius, b.pos, b.radius) {
				collisionPoint := rl.Vector2Scale(rl.Vector2Add(a.pos, b.pos), 0.5)

				combinedVelocity := rl.Vector2Add(a.velocity, b.velocity)

				collisionMagnitude := (2 * rl.Vector2Length(combinedVelocity)) / MaxPushVelocityAllowed

				collidingPairs = append(collidingPairs, collisionPair{a, b, collisionPoint, collisionMagnitude})
			}
		}
	}

	for _, p := range collidingPairs {
		speedDiff := rl.Vector2Length(rl.Vector2Subtract(p.a.velocity, p.b.velocity))
		aIsFaster := rl.Vector2Length(p.a.velocity) > rl.Vector2Length(p.b.velocity)
		amount := rl.Clamp(speedDiff, 0, MaxPushVelocityAllowed) * 2

		resolvePenetrationDepth(p.a, p.b)
		resolveCollision(p.a, p.b)
		level.hitStoneMoving = nil

		if aIsFaster {
			p.b.life -= amount
			p.a.life -= amount * 0.2
		} else {
			p.a.life -= amount
			p.b.life -= amount * 0.2
		}

		for i := float32(0.0); i < 100; i += 0.5 {
			// TODO: shard size should depend on the screen size
			shardColor := level.playerSettings[p.a.playerId].primaryColor
			if rand.Float32() > 0.5 {
				shardColor = level.playerSettings[p.b.playerId].primaryColor
			}
			part := NewShard(
				p.collisionPoint,
				3.6*i,
				MaxParticleSpeed*rand.Float32(),
				p.magnitude,
				MaxShardRadius*(rand.Float32()+0.5),
				shardColor,
				true,
			)

			level.allShards = append(level.allShards, part)
		}
	}

	screenRect := window.GetScreenBoundary()

	newlyDeadStonesIx := []int{}

	for i := range level.stones {
		stone := &level.stones[i]
		if stone.isDead {
			continue
		}
		stone.pos = rl.Vector2Add(stone.pos, stone.velocity)
		calcVelocity(stone)

		if !rl.CheckCollisionPointRec(stone.pos, screenRect) || stone.life <= 0 {
			stone.isDead = true
			newlyDeadStonesIx = append(newlyDeadStonesIx, i)
			if level.hitStoneMoving == stone {
				level.hitStoneMoving = nil
			}
		}
	}

	if level.selectedStone != nil {
		strength := rl.Vector2Distance(level.aimVectorStart, level.selectedStone.pos)
		strength = rl.Clamp(MaxPullLengthAllowed, 0, strength)
		level.selectedStoneRotAnimationAngle += rl.GetFrameTime() * 3 * strength
	}

	{ // creates the shards at the position of the dead stone
		for _, ix := range newlyDeadStonesIx {
			stone := &level.stones[ix]

			shardColor := level.playerSettings[stone.playerId].primaryColor

			for i := float32(0.0); i < 300; i += 0.5 {
				part := NewShard(
					stone.pos,
					3.6*i,
					MaxParticleSpeed*rand.Float32(),
					2,
					MaxShardRadius*(rand.Float32()+0.5),
					shardColor,
					false,
				)

				level.allShards = append(level.allShards, part)
			}
		}
	}

	if level.action == StoneHit {
		// find the diff between the selected stone and where the mouse is
		diff := rl.Vector2Subtract(level.selectedStone.pos, level.aimVectorStart)
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

		level.selectedStone.velocity = v

		level.action = NoAction
		level.hitStoneMoving = level.selectedStone
		level.selectedStone = nil
		level.selectedStoneRotAnimationAngle = 0
		if level.playerTurn == PlayerOne {
			level.playerTurn = PlayerTwo
		} else {
			level.playerTurn = PlayerOne
		}
	}

	{
		// rocket exhaust
		stone := level.hitStoneMoving

		if stone != nil {
			rocketColor := level.playerSettings[stone.playerId].rocketColor

			generalAngle := (rl.Vector2Angle(
				rl.Vector2Normalize(stone.velocity),
				rl.NewVector2(1, 0),
			) * rl.Rad2deg) - 180

			life := 0.3 * (rl.Vector2Length(stone.velocity)) / 15

			for range 25 {
				angle := generalAngle + float32((rand.Intn(20) - 10))
				part := NewParticle(
					stone.pos,
					angle,
					MaxParticleSpeed*rand.Float32(),
					life,
					stone.radius*1.02,
					rocketColor,
				)

				level.allParticles = append(level.allParticles, part)
			}

			if rl.Vector2Length(stone.velocity) == 0 {
				level.hitStoneMoving = nil
			}

		}
	}

	for i := range level.allParticles {
		level.allParticles[i].update()
	}

	for i := range level.allShards {
		level.allShards[i].update()
	}

	{
		// filter out the dead particles
		newAllParticles := []Particle{}
		for _, p := range level.allParticles {
			if p.life > 0 {
				newAllParticles = append(newAllParticles, p)
			}
		}

		level.allParticles = newAllParticles
	}

	{
		// filter out the dead shards
		newShards := []Shard{}
		for _, p := range level.allShards {
			if p.life > 0 {
				newShards = append(newShards, p)
			}
		}

		level.allShards = newShards
	}

	if level.status == Initialized {
		// scoring calculation
		scorePlayerOne := 0
		scorePlayerTwo := 0

		for _, stone := range level.stones {
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

		level.score[PlayerOne] = uint8(scorePlayerOne)
		level.score[PlayerTwo] = uint8(scorePlayerTwo)

		if scorePlayerOne*scorePlayerTwo == 0 {
			level.status = Finished
			level.playerTurn = PlayerOne
		}
	}

	level.checkStonesForMovements()
	level.lastTimeUpdated = rl.GetTime()
	level.totalTimeRunning += rl.GetFrameTime()
}

func (level *Level) setAimVectorStart(aimVectorStart rl.Vector2) {
	level.aimVectorStart = aimVectorStart
	if level.selectedStone != nil {
		level.aimVectorForwardExtensionEnd = rl.Vector2Add(level.selectedStone.pos, rl.Vector2Negate(rl.Vector2Subtract(level.aimVectorStart, level.selectedStone.pos)))
	}
}

func (level *Level) handleUserInput(window *Window) {
	if rl.IsKeyDown(rl.KeyS) {
		if level.status == Stopped {
			level.status = Initialized
		} else {
			level.status = Stopped
		}
	}

	if level.status != Stopped {
		if level.playerTurn == PlayerOne {
			level.handleMouseMove()
		} else {
			level.handleCpuMove(window)
		}
	}
}

func (level *Level) handleMouseMove() {
	level.setAimVectorStart(rl.GetMousePosition())

	if rl.IsMouseButtonDown(rl.MouseButtonRight) && level.stonesAreStill {
		for i, stone := range level.stones {
			if stone.isDead {
				continue
			}
			if level.playerTurn == stone.playerId && rl.CheckCollisionPointCircle(level.aimVectorStart, stone.pos, stone.radius) {
				level.selectedStone = &level.stones[i]
				level.action = StoneAimed
				break
			}
		}
	}

	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && level.action == StoneAimed {
		level.action = StoneHit
	}
}

func (level *Level) handleCpuMove(window *Window) {
	if !level.stonesAreStill || level.status == Finished {
		return
	}

	actor, target := cpuSearchBestOption(level, window)

	if actor == nil || target == nil {
		return
	}

	level.selectedStone = actor
	level.action = StoneHit
	clampedV := rl.Vector2Subtract(target.pos, actor.pos)
	clampedV = rl.Vector2ClampValue(clampedV, 0.0, MaxPullLengthAllowed)
	clampedV = rl.Vector2Negate(clampedV)
	clampedV = rl.Vector2Add(actor.pos, clampedV)

	screenBoundary := window.GetScreenBoundary()
	boundaryLines := window.GetScreenBoundaryLines()

	if !rl.CheckCollisionPointRec(clampedV, screenBoundary) {
		for _, line := range boundaryLines {
			point, ok := getLineToLineIntersectionPoint(line[0], line[1], clampedV, actor.pos)
			if ok {
				clampedV = point
				break
			}
		}
	}

	level.setAimVectorStart(clampedV)
}

func (level *Level) draw(window *Window) {
	rl.ClearBackground(level.levelSettings.backgroundColor)

	screenWidth, screenHeight := window.GetScreenDimensions()

	if level.status != Finished {
		drawScore(screenWidth, screenHeight, level)

		// draw the vertical centre line
		rl.DrawLineEx(
			rl.NewVector2(screenWidth/2, 0),
			rl.NewVector2(screenWidth/2, screenHeight),
			screenWidth/256,
			dimWhite(125),
		)

		// draw the aim bubbles
		if level.action == StoneAimed {
			for i := float32(0.0); i <= 1.0; i += 0.1 {
				amount := i + level.totalTimeRunning/10
				amount = amount - float32(int(amount))
				point := rl.Vector2Lerp(level.selectedStone.pos, level.aimVectorForwardExtensionEnd, amount)
				rl.DrawCircleV(point, StoneRadius*0.4*(1-amount), dimWhite(50))
			}
		}

		// draw the stones
		for i := range level.stones {
			stone := &(level.stones[i])
			drawStone(stone, level)
		}

		// draw the aim line
		if level.action == StoneAimed {
			rl.DrawCircleV(level.selectedStone.pos, StoneRadius*0.1, dimWhite(60))
			rl.DrawLineEx(
				level.aimVectorStart,
				level.selectedStone.pos,
				3.0,
				dimWhite(60),
			)
		}
	}

	{
		// draw particles
		for _, p := range level.allParticles {
			rl.BeginBlendMode(rl.BlendAdditive)
			p.render()
			rl.EndBlendMode()
		}
	}

	{
		// draw shards
		for _, p := range level.allShards {
			p.render()
		}
	}
}

func drawStone(s *Stone, level *Level) {
	if s.isDead {
		return
	}
	playerSettings := level.playerSettings[s.playerId]

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

	if level.stonesAreStill && s.playerId == level.playerTurn && level.playerTurn == PlayerOne {
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

		if level.selectedStone == s {
			rl.DrawRing(
				s.pos,
				s.radius*1.1,
				s.radius*1.4,
				0.0+level.selectedStoneRotAnimationAngle,
				40.0+level.selectedStoneRotAnimationAngle,
				0,
				dimWhite(100),
			)
		} else {
			rl.DrawRing(
				s.pos,
				s.radius*1.1,
				s.radius*1.4,
				0.0+level.totalTimeRunning*10,
				40.0+level.totalTimeRunning*10,
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
		360.0*s.life/100,
		0,
		playerSettings.lifeColor,
	)
}

func drawScore(screenWidth, screenHeight float32, level *Level) {
	color := dimWhite(60)
	labelP1 := level.playerSettings[PlayerOne].label
	labelP2 := level.playerSettings[PlayerTwo].label

	defaultFont := rl.GetFontDefault()

	p1Score := fmt.Sprintf("0%d", level.score[PlayerOne])
	p2Score := fmt.Sprintf("0%d", level.score[PlayerTwo])

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
