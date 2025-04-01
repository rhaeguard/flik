package main

import rl "github.com/gen2brain/raylib-go/raylib"

type SceneLevelsBasic struct {
	level Level
}

func NewSceneLevelsBasic() SceneLevelsBasic {
	return SceneLevelsBasic{}
}

func (scene *SceneLevelsBasic) Init(window *Window) {
	// init
	level := newLevel()
	level.init(window)
	scene.level = level
}

func (scene *SceneLevelsBasic) GetId() SceneId {
	return Levels
}
func (scene *SceneLevelsBasic) HandleUserInput(window *Window) {
	if rl.IsKeyDown(rl.KeyS) {
		if scene.level.status == Stopped {
			scene.level.status = Initialized
		} else {
			scene.level.status = Stopped
		}
	}

	scene.level.stonesAreStill = areStonesStill(&scene.level)

	if scene.level.status != Stopped {
		if scene.level.playerTurn == PlayerOne {
			handleMouseMove(&scene.level)
		} else {
			handleCpuMove(&scene.level, window)
		}
	}
}
func (scene *SceneLevelsBasic) Update(window *Window) {
	if scene.level.status != Stopped {
		update(&scene.level, window)
	}
}
func (scene *SceneLevelsBasic) Draw(window *Window) {
	draw(&scene.level, window)

}
