package main

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type SceneGameOver struct {
	data        *Level
	nextSceneId SceneId
}

func NewSceneGameOver() SceneGameOver {
	return SceneGameOver{}
}

func (scene *SceneGameOver) GetId() SceneId {
	return GameOver
}

func (scene *SceneGameOver) Init(data any, window *Window) {
	scene.data = data.(*Level)
	scene.nextSceneId = scene.GetId()
}

func (scene *SceneGameOver) HandleUserInput(window *Window) {
	if rl.IsKeyDown(rl.KeySpace) {
		scene.nextSceneId = Levels
	}
}

func (scene *SceneGameOver) Update(window *Window) (SceneId, any) {
	for i := range scene.data.allShards {
		scene.data.allShards[i].update()
	}

	return scene.nextSceneId, nil
}

func (scene *SceneGameOver) Draw(window *Window) {
	screenWidth, screenHeight := window.GetScreenDimensions()

	whoWon := scene.data.playerSettings[PlayerOne].label
	if scene.data.score[PlayerOne] == 0 {
		whoWon = scene.data.playerSettings[PlayerTwo].label
	}
	whoWon = fmt.Sprintf("%s won!", whoWon)
	measuredSize := rl.MeasureTextEx(rl.GetFontDefault(), whoWon, FontSize/3, 10)
	w := (screenWidth - measuredSize.X) / 2
	h := (screenHeight - measuredSize.Y) / 3

	rl.DrawTextEx(
		rl.GetFontDefault(),
		whoWon,
		rl.NewVector2(w, h),
		FontSize/3,
		10,
		dimWhite(60),
	)

	message2 := rl.MeasureTextEx(rl.GetFontDefault(), "press space to restart", FontSize/5, 10)
	w = (screenWidth - message2.X) / 2
	h = h + measuredSize.Y*1.5
	rl.DrawTextEx(
		rl.GetFontDefault(),
		"press space to restart",
		rl.NewVector2(w, h),
		FontSize/5,
		10,
		dimWhite(60),
	)

	{
		// draw shards
		for _, p := range scene.data.allShards {
			p.render()
		}
	}
}

func (scene *SceneGameOver) Teardown(window *Window) {

}
