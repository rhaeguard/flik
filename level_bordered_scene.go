package main

import rl "github.com/gen2brain/raylib-go/raylib"

type SceneLevelsBordered struct {
	level          Level
	levelSettings  LevelSettings
	playerSettings map[Player]PlayerSettings
}

func NewSceneLevelsBordered(window *Window) SceneLevelsBordered {
	return SceneLevelsBordered{
		levelSettings: LevelSettings{
			sceneId:         LevelBordered,
			stonesPerPlayer: 4,
			backgroundColor: BG_COLOR,
			isBordered:      true,
			boundary:        window.GetScreenBoundary(),
		},
		playerSettings: map[Player]PlayerSettings{
			PlayerOne: {
				label:          "you",
				primaryColor:   rl.NewColor(55, 113, 142, 255),
				outerRingColor: rl.NewColor(37, 78, 112, 255),
				lifeColor:      rl.NewColor(255, 250, 255, 255),
				rocketColor:    rl.SkyBlue,
				isCpu:          false,
			},
			PlayerTwo: {
				label:          "cpu",
				primaryColor:   rl.NewColor(133, 90, 92, 255),
				outerRingColor: rl.NewColor(102, 16, 31, 255),
				lifeColor:      rl.NewColor(255, 250, 255, 255),
				rocketColor:    rl.NewColor(129, 13, 32, 255),
				isCpu:          true,
			},
		},
	}
}

func (scene *SceneLevelsBordered) Init(data any, window *Window) {
	// init
	level := newLevel(scene.levelSettings, scene.playerSettings)
	level.init(window)
	scene.level = level
}

func (scene *SceneLevelsBordered) GetId() SceneId {
	return LevelBordered
}

func (scene *SceneLevelsBordered) HandleUserInput(window *Window) {
	scene.level.handleUserInput(window)
}

func (scene *SceneLevelsBordered) Update(window *Window) (SceneId, any) {
	nextSceneId := scene.GetId()
	var levelData any = nil
	if scene.level.status != Stopped {
		scene.level.update(window)
		if scene.level.status == Finished {
			nextSceneId = Transition
			levelData = &scene.level
		}
	}
	return nextSceneId, levelData
}

func (scene *SceneLevelsBordered) Draw(window *Window) {
	scene.level.draw(window)
}

func (scene *SceneLevelsBordered) Teardown(window *Window) {

}
