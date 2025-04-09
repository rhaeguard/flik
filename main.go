package main

import (
	_ "embed"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var BG_COLOR = rl.NewColor(139, 212, 195, 255)

// magic numbers
var VelocityDampingFactor float32
var VelocityThresholdToStop float32
var MaxPullLengthAllowed float32
var MaxPushVelocityAllowed float32
var StoneRadius float32
var StoneSelectionCancelCircleRadius float32
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

type Game struct {
	status       GameStatus
	currentScene SceneId
	scenes       [TotalSceneCount]Scene
}

func NewGame() Game {
	return Game{
		status: GameUninitialized,
		scenes: [TotalSceneCount]Scene{},
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
	StoneSelectionCancelCircleRadius = StoneRadius * 0.2
	FontSize = screenWidth * 0.25

	// initialize the gameLevelScene
	mainScene := NewSceneMain(window)
	g.scenes[Main] = &mainScene

	levelBasic := NewSceneLevelsBasic()
	g.scenes[LevelBasic] = &levelBasic

	levelBordered := NewSceneLevelsBordered(window)
	g.scenes[LevelBordered] = &levelBordered

	levelTimed := NewSceneLevelsTimeLimit(window)
	g.scenes[LevelTimeLimit] = &levelTimed

	gameOverScene := NewSceneTransition()
	g.scenes[Transition] = &gameOverScene

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
	g.scenes[LevelBasic].Teardown(window)
	g.scenes[LevelBordered].Teardown(window)
	g.scenes[Transition].Teardown(window)

	for i := range int(TotalSceneCount) {
		if s := g.scenes[i]; s != nil {
			s.Teardown(window)
		}
	}
}

//go:embed bin/assets/bg.ogg
var backgroundMusic []byte

//go:embed bin/assets/icon.png
var iconImage []byte

func main() {
	game := NewGame()
	window := Window{
		fullscreen: IsFullscreen,
		width:      1920,
		height:     1080,
		title:      "flik",
	}

	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	if window.fullscreen {
		rl.InitWindow(0, 0, window.title)
		rl.ToggleFullscreen()
	} else {
		rl.InitWindow(window.width, window.height, window.title)
	}
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	icon := rl.LoadImageFromMemory(".png", iconImage, int32(len(iconImage)))
	defer rl.UnloadImage(icon)

	rl.SetWindowIcon(*icon)

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	bgMusic := rl.LoadMusicStreamFromMemory(".ogg", backgroundMusic, int32(len(backgroundMusic)))
	rl.SetMusicVolume(bgMusic, 0.125)
	defer rl.UnloadMusicStream(bgMusic)

	// starts playing the music
	rl.PlayMusicStream(bgMusic)

	for !rl.WindowShouldClose() {
		if game.status == GameUninitialized {
			(&game).Init(&window)
		}

		// loops the music
		rl.UpdateMusicStream(bgMusic)

		if game.Update(&window) != 0 {
			break
		}

		rl.BeginDrawing()

		game.Draw(&window)

		rl.EndDrawing()
	}

	game.Teardown(&window)
}
