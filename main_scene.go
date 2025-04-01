package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type buttonRectangle struct {
	text        string
	rectangle   rl.Rectangle
	fontSize    float32
	active      bool
	targetScene SceneId
}

type SceneMain struct {
	nextSceneId      SceneId
	logoText         string
	logoBoundingBox  rl.Rectangle
	logoFontSize     float32
	buttonRectangles []buttonRectangle
}

func NewSceneMain() SceneMain {
	return SceneMain{}
}

func (scene *SceneMain) GetId() SceneId {
	return Main
}

func (scene *SceneMain) Init(data any, window *Window) {
	scene.nextSceneId = scene.GetId()

	screenWidth, screenHeight := window.GetScreenDimensions()
	defaultFont := rl.GetFontDefault()

	text := "...flik..."
	measuredSize := rl.MeasureTextEx(defaultFont, text, FontSize, 10)
	w := (screenWidth - measuredSize.X) / 2
	h := (screenHeight - measuredSize.Y) / 4

	scene.logoText = text
	scene.logoFontSize = FontSize
	scene.logoBoundingBox = rl.NewRectangle(w, h, measuredSize.X, measuredSize.Y)

	playText := rl.MeasureTextEx(defaultFont, "play", FontSize/5, 10)
	w = (screenWidth - playText.X) / 2
	h = h + measuredSize.Y*1.02 // 2% gap

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text:        "play",
		rectangle:   rl.NewRectangle(w, h, playText.X, playText.Y),
		fontSize:    FontSize / 5,
		targetScene: Levels,
	})

	controls := rl.MeasureTextEx(defaultFont, "controls", FontSize/5, 10)
	w = (screenWidth - controls.X) / 2
	h = h + playText.Y*1.02 // 2% gap

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text:        "controls",
		rectangle:   rl.NewRectangle(w, h, controls.X, controls.Y),
		fontSize:    FontSize / 5,
		targetScene: Main,
	})

	quitText := rl.MeasureTextEx(defaultFont, "quit", FontSize/5, 10)
	w = (screenWidth - quitText.X) / 2
	h = h + playText.Y*1.02 // 2% gap

	scene.buttonRectangles = append(scene.buttonRectangles, buttonRectangle{
		text:        "quit",
		rectangle:   rl.NewRectangle(w, h, quitText.X, quitText.Y),
		fontSize:    FontSize / 5,
		targetScene: Main,
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
}

func (scene *SceneMain) Update(window *Window) (SceneId, any) {
	mousePosition := rl.GetMousePosition()

	for bi, buttonConfig := range scene.buttonRectangles {
		scene.buttonRectangles[bi].active = rl.CheckCollisionPointRec(mousePosition, buttonConfig.rectangle)
	}

	return scene.nextSceneId, nil
}

func (scene *SceneMain) Draw(window *Window) {
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

		if buttonConfig.active {
			dimLevel = 255
		}

		rl.DrawTextEx(
			rl.GetFontDefault(),
			buttonConfig.text,
			rl.NewVector2(buttonConfig.rectangle.X, buttonConfig.rectangle.Y),
			buttonConfig.fontSize,
			10,
			dimWhite(uint8(dimLevel)),
		)
	}

}
