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
		status:       GameUninitialized,
		scenes:       map[SceneId]Scene{},
		settings:     GameSettings{},
		currentScene: Main,
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

	levelBordered := NewSceneLevelsBordered()
	g.scenes[LevelBordered] = &levelBordered

	gameOverScene := NewSceneGameOver()
	g.scenes[GameOver] = &gameOverScene

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

	if g.currentScene != nextSceneId {
		// fmt.Printf("Scene change [%d => %d]\n", g.currentScene, nextSceneId)
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
	g.scenes[GameOver].Teardown(window)
}

func main() {
	game := NewGame()
	window := Window{
		fullscreen: true,
		width:      14401280,
		height:     810,
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
