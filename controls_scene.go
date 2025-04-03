package main

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

type sceneElement struct {
	text         string
	rectangle    rl.Rectangle
	fontSize     float32
	interactable bool
	active       bool
	targetScene  SceneId
}

type SceneControls struct {
	nextSceneId     SceneId
	texture         rl.Texture2D
	colorSwapShader rl.Shader
	sceneElements   []sceneElement
}

func NewSceneControls() SceneControls {
	return SceneControls{}
}

func (scene *SceneControls) GetId() SceneId {
	return Controls
}

func (scene *SceneControls) Init(data any, window *Window) {
	scene.nextSceneId = scene.GetId()
	screenWidth, screenHeight := window.GetScreenDimensions()

	measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), "controls", FontSize/3, 10)
	w := (screenWidth - measuredSize.X) / 2
	h := (screenHeight - measuredSize.Y) / 3

	scene.sceneElements = append(scene.sceneElements, sceneElement{
		text:         "controls",
		fontSize:     FontSize / 3,
		rectangle:    rl.NewRectangle(w, h, measuredSize.X, measuredSize.Y),
		interactable: false,
	})

	instructionTextMeasured := rl.MeasureTextEx(rl.GetFontDefault(), "right-click to pick a circle", FontSize/10, 10)
	w = (screenWidth - instructionTextMeasured.X) / 2
	h = h + measuredSize.Y*1.2

	scene.sceneElements = append(scene.sceneElements, sceneElement{
		text:         "right-click to pick a circle",
		fontSize:     FontSize / 10,
		rectangle:    rl.NewRectangle(w, h, instructionTextMeasured.X, instructionTextMeasured.Y),
		interactable: false,
	})

	/////

	instruction2TextMeasured := rl.MeasureTextEx(rl.GetFontDefault(), "left-click to attack", FontSize/10, 10)
	w = (screenWidth - instruction2TextMeasured.X) / 2
	h = h + instructionTextMeasured.Y*1.2

	scene.sceneElements = append(scene.sceneElements, sceneElement{
		text:         "left-click to attack",
		fontSize:     FontSize / 10,
		rectangle:    rl.NewRectangle(w, h, instruction2TextMeasured.X, instruction2TextMeasured.Y),
		interactable: false,
	})

	////

	backButton := rl.MeasureTextEx(rl.GetFontDefault(), "back", FontSize/5, 10)
	w = (screenWidth - backButton.X) / 2
	h = h + instruction2TextMeasured.Y*2

	scene.sceneElements = append(scene.sceneElements, sceneElement{
		text:         "back",
		fontSize:     FontSize / 5,
		rectangle:    rl.NewRectangle(w, h, backButton.X, backButton.Y),
		interactable: true,
		targetScene:  Main,
	})

	// 	scene.texture = rl.LoadTexture("mouse.png")

	// 	scene.colorSwapShader = rl.LoadShaderFromMemory(``, `#version 330
	// // Input vertex attributes (from vertex shader)
	// in vec2 fragTexCoord;
	// in vec4 fragColor;

	// // Input uniform values
	// uniform sampler2D texture0;

	// out vec4 finalColor;

	// void main()
	// {
	// 	vec4 texelColor = texture(texture0, fragTexCoord);

	//		if (texelColor == vec4(1.0, 1.0, 1.0, 1.0)) {
	//			discard;
	//		} else {
	//			finalColor = vec4(1.0, 1.0, 1.0, 0.5);
	//		}
	//	}
	//
	// `)
}

func (scene *SceneControls) HandleUserInput(window *Window) {
	if rl.IsMouseButtonReleased(rl.MouseButtonLeft) {
		for _, se := range scene.sceneElements {
			if se.active {
				scene.nextSceneId = se.targetScene
			}
		}
	}
}

func (scene *SceneControls) Update(window *Window) (SceneId, any) {
	mousePosition := rl.GetMousePosition()

	for bi, se := range scene.sceneElements {
		if !se.interactable {
			continue
		}
		scene.sceneElements[bi].active = rl.CheckCollisionPointRec(mousePosition, se.rectangle)
	}
	return scene.nextSceneId, nil
}

func (scene *SceneControls) Draw(window *Window) {
	// draw background
	rl.ClearBackground(BG_COLOR)

	for _, e := range scene.sceneElements {
		dimValue := 100
		if e.active {
			dimValue = 255
		}
		rl.DrawTextEx(
			rl.GetFontDefault(),
			e.text,
			rl.NewVector2(e.rectangle.X, e.rectangle.Y),
			e.fontSize,
			10,
			dimWhite(uint8(dimValue)),
		)
	}

	// mouseLeft := rl.NewVector2(
	// 	(screenWidth/2-float32(scene.texture.Width)/2)/2,
	// 	h+measuredSize.Y*1.2,
	// )

	// mouseRight := rl.NewVector2(
	// 	(screenWidth/2-float32(scene.texture.Width)/2)/2+screenWidth/2,
	// 	h+measuredSize.Y*1.2,
	// )

	// originalSize := rl.NewVector2(float32(scene.texture.Width)/2, float32(scene.texture.Height))
	// adjustedSize := rl.NewVector2(
	// 	(originalSize.X/originalSize.Y)*screenHeight*0.4,
	// 	screenHeight*0.4,
	// )

	// rl.BeginShaderMode(scene.colorSwapShader)

	// rl.DrawTexturePro(
	// 	scene.texture,
	// 	rl.NewRectangle(0, 0, originalSize.X, originalSize.Y),
	// 	rl.NewRectangle(mouseLeft.X, mouseLeft.Y, adjustedSize.X, adjustedSize.Y),
	// 	rl.NewVector2(0, 0),
	// 	0,
	// 	rl.White,
	// )

	// rl.DrawTexturePro(
	// 	scene.texture,
	// 	rl.NewRectangle(float32(scene.texture.Width)/2, 0, float32(scene.texture.Width)/2, float32(scene.texture.Height)),
	// 	rl.NewRectangle(mouseRight.X, mouseRight.Y, float32(scene.texture.Width)/2, float32(scene.texture.Height)),
	// 	rl.NewVector2(0, 0),
	// 	0,
	// 	rl.White,
	// )

	// rl.EndShaderMode()
}

func (scene *SceneControls) Teardown(window *Window) {
	// rl.UnloadTexture(scene.texture)
	// rl.UnloadShader(scene.colorSwapShader)
}
