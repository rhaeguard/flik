package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SceneTransition struct {
	data             *Level
	winner           Player
	nextSceneId      SceneId
	message          buttonRectangle
	buttonRectangles []buttonRectangle
}

func NewSceneTransition() SceneTransition {
	return SceneTransition{}
}

func (scene *SceneTransition) GetId() SceneId {
	return Transition
}

func (scene *SceneTransition) Init(data any, window *Window) {
	scene.buttonRectangles = nil

	scene.data = data.(*Level)
	scene.nextSceneId = scene.GetId()

	scene.winner = PlayerOne
	if scene.data.score[PlayerOne] == 0 {
		scene.winner = PlayerTwo
	}

	{
		screenWidth, screenHeight := window.GetScreenDimensions()

		whoWon := fmt.Sprintf("%s won!", scene.data.playerSettings[scene.winner].label)
		offsetX := float32(0.0)
		if scene.winner == PlayerOne {
			offsetX = screenWidth / 2
		}
		measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), whoWon, FontSize/4, 10)
		w := offsetX + (screenWidth/2-measuredSize.X)/2
		h := (screenHeight - measuredSize.Y) / 2.5

		scene.message = buttonRectangle{
			text:      whoWon,
			rectangle: rl.NewRectangle(w, h, measuredSize.X, measuredSize.Y),
			fontSize:  FontSize / 4,
		}

		h = h + measuredSize.Y

		if scene.winner == PlayerOne { // TODO: we probably should not hardcode this cause it locks the player to the left half of the screen
			next := rl.MeasureTextEx(rl.GetFontDefault(), "next", FontSize/7, 10)

			w = offsetX + (screenWidth/2-next.X)/2
			h = h + next.Y*1.2

			targetScene := LevelProgression[scene.data.levelSettings.sceneId]

			scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
				text:        "next",
				rectangle:   rl.NewRectangle(w, h, next.X, next.Y),
				fontSize:    FontSize / 7,
				targetScene: targetScene,
			})
		}

		restart := rl.MeasureTextEx(rl.GetFontDefault(), "restart", FontSize/7, 10)

		w = offsetX + (screenWidth/2-restart.X)/2
		h = h + restart.Y*1.2

		scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
			text:        "restart",
			rectangle:   rl.NewRectangle(w, h, restart.X, restart.Y),
			fontSize:    FontSize / 7,
			targetScene: scene.data.levelSettings.sceneId,
		})

		mainMenu := rl.MeasureTextEx(rl.GetFontDefault(), "main menu", FontSize/7, 10) // TODO: should spacing be static????

		w = offsetX + (screenWidth/2-mainMenu.X)/2
		h = h + mainMenu.Y*1.2

		scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
			text:        "main menu",
			rectangle:   rl.NewRectangle(w, h, mainMenu.X, mainMenu.Y),
			fontSize:    FontSize / 7,
			targetScene: Main,
		})
	}
}

func (scene *SceneTransition) HandleUserInput(window *Window) {
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		for _, buttonConfig := range scene.buttonRectangles {
			if buttonConfig.active {
				scene.nextSceneId = buttonConfig.targetScene
			}
		}
	}
}

func (scene *SceneTransition) Update(window *Window) (SceneId, any) {
	for i := range scene.data.allShards {
		scene.data.allShards[i].update()
	}

	for i := range scene.data.allParticles {
		scene.data.allParticles[i].update()
	}

	mousePosition := rl.GetMousePosition()

	for bi, buttonConfig := range scene.buttonRectangles {
		scene.buttonRectangles[bi].active = rl.CheckCollisionPointRec(mousePosition, buttonConfig.rectangle)
	}

	return scene.nextSceneId, nil
}

func (scene *SceneTransition) Draw(window *Window) {
	// draw background
	rl.ClearBackground(BG_COLOR)
	// this prevents timebox from being rendered
	scene.data.levelSettings.isTimed = false
	scene.data.drawField(window) // TODO: might be excessive?

	screenWidth, screenHeight := window.GetScreenDimensions()

	offsetX := float32(0.0)
	if scene.winner == PlayerOne {
		offsetX = screenWidth / 2
	}

	// TODO: maybe account for the centre line thickness as well
	rl.DrawRectangleV(
		rl.NewVector2(offsetX, 0), rl.NewVector2(screenWidth/2, screenHeight), scene.data.levelSettings.backgroundColor,
	)

	rl.DrawTextEx(
		rl.GetFontDefault(),
		scene.message.text,
		rl.NewVector2(scene.message.rectangle.X, scene.message.rectangle.Y),
		scene.message.fontSize,
		10,
		dimWhite(60),
	)

	for _, btn := range scene.buttonRectangles {

		dimLevel := uint8(60)

		if btn.active {
			dimLevel = 255
		}

		rl.DrawTextEx(
			rl.GetFontDefault(),
			btn.text,
			rl.NewVector2(btn.rectangle.X, btn.rectangle.Y),
			btn.fontSize,
			10,
			dimWhite(dimLevel),
		)
	}

	{
		// draw shards
		// TODO: maybe the particles too?
		for _, p := range scene.data.allShards {
			p.render()
		}
	}
}

func (scene *SceneTransition) Teardown(window *Window) {

}
