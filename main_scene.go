package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const GAME_INSTRUCTIONS = "> Click & drag a circle to aim and charge\n> Release to attack\n> Drag back to center to cancel"

type buttonRectangle struct {
	text         string
	rectangle    rl.Rectangle
	fontSize     float32
	fontSpacing  float32
	interactable bool
	active       bool
	targetScene  SceneId
}

type SceneMain struct {
	nextSceneId      SceneId
	logoText         string
	logoBoundingBox  rl.Rectangle
	logoFontSize     float32
	buttonRectangles []buttonRectangle

	//
	level          Level
	levelSettings  LevelSettings
	playerSettings [TotalPlayerCount]PlayerSettings
}

func NewSceneMain(window *Window) SceneMain {
	b := window.GetScreenBoundary()

	frameWidth := b.Width * 0.5
	frameHeight := b.Height * 0.6

	bb := rl.NewRectangle(
		frameWidth,
		(b.Height-frameHeight)*0.3,
		frameWidth,
		frameHeight,
	)

	return SceneMain{
		levelSettings: LevelSettings{
			sceneId:         LevelBordered,
			stonesPerPlayer: 1,
			backgroundColor: BG_COLOR,
			isBordered:      true,
			boundary:        bb,
		},
		playerSettings: [TotalPlayerCount]PlayerSettings{
			PlayerOne: getPlayer("p1", HumanPlayerPalette1, true),
			PlayerTwo: getPlayer("p2", CpuPlayerPalette1, true),
		},
	}
}

func (scene *SceneMain) GetId() SceneId {
	return Main
}

func (scene *SceneMain) Init(data any, window *Window) {
	scene.nextSceneId = scene.GetId()

	// initialize the tutorial game
	level := newLevel(scene.levelSettings, scene.playerSettings)
	scene.level = level
	scene.level.status = Initialized

	ww := level.levelSettings.boundary.Width
	hh := level.levelSettings.boundary.Height

	playerOneStone := newStone(0, level.levelSettings.boundary.X+ww*0.25, level.levelSettings.boundary.Y+0.75*hh, StoneRadius, 1, PlayerOne)
	playerTwoStone := newStone(1, level.levelSettings.boundary.X+ww*0.75, level.levelSettings.boundary.Y+0.25*hh, StoneRadius, 1, PlayerTwo)

	scene.level.stones = []Stone{
		playerOneStone, playerTwoStone,
	}

	// the buttons, the text and the logo

	screenWidth, screenHeight := window.GetScreenDimensions()
	defaultFont := rl.GetFontDefault()

	text := window.title
	measuredSize := rl.MeasureTextEx(defaultFont, text, FontSize/2, 10)
	w := (screenWidth/2 - measuredSize.X) / 2
	h := (screenHeight - measuredSize.Y) / 4

	scene.logoText = text
	scene.logoFontSize = FontSize / 2
	scene.logoBoundingBox = rl.NewRectangle(w, h, measuredSize.X, measuredSize.Y)

	playText := rl.MeasureTextEx(defaultFont, "play", FontSize/5, 10)
	h = h + measuredSize.Y*1.05 // 5% gap

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text:         "play",
		rectangle:    rl.NewRectangle(w, h, playText.X, playText.Y),
		fontSize:     FontSize / 5,
		targetScene:  InitialLevel,
		interactable: true,
	})

	quitText := rl.MeasureTextEx(defaultFont, "quit", FontSize/5, 10)
	h = h + playText.Y*1.02 // 2% gap

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text:         "quit",
		rectangle:    rl.NewRectangle(w, h, quitText.X, quitText.Y),
		fontSize:     FontSize / 5,
		targetScene:  Quit,
		interactable: true,
	})

	praticeText := rl.MeasureTextEx(rl.GetFontDefault(), "practice", FontSize/5, 10)

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text: "pratice",
		rectangle: rl.NewRectangle(
			scene.level.levelSettings.boundary.X+(scene.level.levelSettings.boundary.X-praticeText.X)/2,
			(scene.level.levelSettings.boundary.Y-praticeText.Y)/2,
			praticeText.X,
			praticeText.Y,
		),
		fontSize: FontSize / 5,
	})

	h = (window.GetScreenBoundary().Height - scene.level.levelSettings.boundary.Y - scene.level.levelSettings.boundary.Height)

	instructionsText := rl.MeasureTextEx(rl.GetFontDefault(), GAME_INSTRUCTIONS, FontSize/12, 5)

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text: GAME_INSTRUCTIONS,
		rectangle: rl.NewRectangle(
			scene.level.levelSettings.boundary.X+(scene.level.levelSettings.boundary.X-instructionsText.X)/2,
			scene.level.levelSettings.boundary.Y+scene.level.levelSettings.boundary.Height+(h-instructionsText.Y)/2,
			instructionsText.X,
			instructionsText.Y,
		),
		fontSize:    FontSize / 12,
		fontSpacing: 5,
	})
}

func (scene *SceneMain) HandleUserInput(window *Window) {
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		for _, buttonConfig := range scene.buttonRectangles {
			if buttonConfig.active {
				scene.nextSceneId = buttonConfig.targetScene
			}
		}
	}

	scene.level.handleUserInput(window)
}

func (scene *SceneMain) Update(window *Window) (SceneId, any) {
	mousePosition := rl.GetMousePosition()

	for bi, buttonConfig := range scene.buttonRectangles {
		scene.buttonRectangles[bi].active = buttonConfig.interactable && rl.CheckCollisionPointRec(mousePosition, buttonConfig.rectangle)
	}

	if rl.CheckCollisionPointRec(mousePosition, scene.level.levelSettings.boundary) {
		scene.level.playerSettings[PlayerOne].isCpu = false
		scene.level.playerSettings[PlayerTwo].isCpu = false
	}

	if scene.level.status != Stopped {
		scene.level.update(window)

		if scene.level.status == Finished {
			// reinit
			scene.level.stones[0].isDead = false
			scene.level.stones[0].life = 100

			scene.level.stones[1].isDead = false
			scene.level.stones[1].life = 100

			scene.level.score[PlayerOne] = 1
			scene.level.score[PlayerTwo] = 1
			scene.level.status = Initialized
		}
	}

	return scene.nextSceneId, nil
}

func (scene *SceneMain) Draw(window *Window) {
	// draw background
	rl.ClearBackground(BG_COLOR)

	{
		rl.DrawRectangleLinesEx(
			scene.level.levelSettings.boundary,
			scene.level.levelSettings.boundary.Width/255,
			dimWhite(125),
		)
	}

	scene.level.drawObjects()

	rl.DrawTextEx(
		rl.GetFontDefault(),
		scene.logoText,
		rl.NewVector2(scene.logoBoundingBox.X, scene.logoBoundingBox.Y),
		scene.logoFontSize,
		10,
		dimWhite(120),
	)

	for _, buttonConfig := range scene.buttonRectangles {
		dimLevel := 60

		if !buttonConfig.interactable {
			dimLevel = 120
		}

		if buttonConfig.active {
			dimLevel = 255
		}

		fontSpacing := buttonConfig.fontSpacing
		if fontSpacing == 0.0 {
			fontSpacing = 10
		}

		rl.DrawTextEx(
			rl.GetFontDefault(),
			buttonConfig.text,
			rl.NewVector2(buttonConfig.rectangle.X, buttonConfig.rectangle.Y),
			buttonConfig.fontSize,
			fontSpacing,
			dimWhite(uint8(dimLevel)),
		)
	}

}

func (scene *SceneMain) Teardown(window *Window) {

}
