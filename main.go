package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

var BG_COLOR = rl.NewColor(139, 212, 195, 255)
var STONE_COLLISION_SHARD_COLOR = rl.NewColor(255, 192, 113, 255)

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

// default values
var IsFullscreen bool = false

type GameStatus uint8

const (
	// game status
	GameUninitialized GameStatus = iota
	GameInitialized   GameStatus = iota
)

type GameSettings struct{}

type Game struct {
	status       GameStatus
	scenes       map[SceneId]Scene
	settings     GameSettings
	currentScene SceneId
}

func NewGame() Game {
	return Game{
		status:   GameUninitialized,
		scenes:   map[SceneId]Scene{},
		settings: GameSettings{},
	}
}

func (g *Game) Init(window *Window) {
	screenWidth, screenHeight := window.GetScreenDimensions()
	// magic numbers
	// ratio is computed based on 2560 x 1440
	VelocityDampingFactor = 0.987
	VelocityThresholdToStop = screenWidth / 6_000
	MaxPullLengthAllowed = 0.1 * screenWidth
	MaxPushVelocityAllowed = 0.008 * screenWidth
	MaxParticleSpeed = 0.008 * screenWidth
	MaxShardRadius = screenWidth / 256
	StoneRadius = screenHeight * 0.06
	FontSize = screenWidth * 0.25

	// initialize the gameLevelScene
	mainScene := NewSceneMain()
	g.scenes[Main] = &mainScene

	levelBasic := NewSceneLevelsBasic()
	g.scenes[LevelBasic] = &levelBasic

	levelBordered := NewSceneLevelsBordered(window)
	g.scenes[LevelBordered] = &levelBordered

	levelTimed := NewSceneLevelsTimeLimit(window)
	g.scenes[LevelTimeLimit] = &levelTimed

	gameOverScene := NewSceneTransition()
	g.scenes[Transition] = &gameOverScene

	controlsScene := NewSceneControls()
	g.scenes[Controls] = &controlsScene

	// set the init status
	g.currentScene = Main
	g.scenes[g.currentScene].Init(nil, window)
	g.status = GameInitialized
}

func (g *Game) Update(window *Window) uint8 {
	scene := g.scenes[g.currentScene]

	scene.HandleUserInput(window)

	nextSceneId, data := scene.Update(window)

	if nextSceneId == Quit {
		return 1
	}

	if rl.IsKeyDown(rl.KeyZero) {
		nextSceneId = Main
	}

	if rl.IsKeyDown(rl.KeyOne) {
		nextSceneId = LevelBasic
	}

	if rl.IsKeyDown(rl.KeyTwo) {
		nextSceneId = LevelBordered
	}

	if rl.IsKeyDown(rl.KeyThree) {
		nextSceneId = LevelTimeLimit
	}

	if g.currentScene != nextSceneId {
		// fmt.Printf("Scene change [%d => %d]\n", g.currentScene, nextSceneId)
		// TODO: it is possible that if we go to the Game Over screen twice
		// TODO: it will just append to the existing struct instance instead of creating a totally new screen
		// TODO: this can be bad because it can result in weird errors.
		g.scenes[nextSceneId].Init(data, window)
		g.currentScene = nextSceneId
	}

	return 0
}

func (g *Game) Draw(window *Window) {
	scene := g.scenes[g.currentScene]
	scene.Draw(window)
}

func (g *Game) Teardown(window *Window) {
	g.scenes[Main].Teardown(window)
	g.scenes[LevelBasic].Teardown(window)
	g.scenes[LevelBordered].Teardown(window)
	g.scenes[Controls].Teardown(window)
	g.scenes[Transition].Teardown(window)
}

func main() {
	game := NewGame()
	window := Window{
		fullscreen: IsFullscreen,
		width:      1920,
		height:     1080,
	}

	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	if window.fullscreen {
		rl.InitWindow(0, 0, "flik")
		rl.ToggleFullscreen()
	} else {
		rl.InitWindow(window.width, window.height, "flik")
	}

	rl.SetTargetFPS(60)

	defer rl.CloseWindow()

	for !rl.WindowShouldClose() {
		if game.status == GameUninitialized {
			(&game).Init(&window)
		}

		if game.Update(&window) != 0 {
			break
		}

		rl.BeginDrawing()

		game.Draw(&window)

		rl.EndDrawing()
	}

	game.Teardown(&window)
}
