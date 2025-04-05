package main

import rl "github.com/gen2brain/raylib-go/raylib"

type SceneLevelsBasic struct {
	level          Level
	levelSettings  LevelSettings
	playerSettings map[Player]PlayerSettings
}

func NewSceneLevelsBasic() SceneLevelsBasic {
	return SceneLevelsBasic{
		levelSettings: LevelSettings{
			sceneId:         LevelBasic,
			stonesPerPlayer: 6,
			backgroundColor: BG_COLOR,
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

func (scene *SceneLevelsBasic) Init(data any, window *Window) {
	// init
	level := newLevel(
		scene.levelSettings,
		scene.playerSettings,
	)
	level.init(window)
	scene.level = level
}

func (scene *SceneLevelsBasic) GetId() SceneId {
	return LevelBasic
}

func (scene *SceneLevelsBasic) HandleUserInput(window *Window) {
	scene.level.handleUserInput(window)
}

func (scene *SceneLevelsBasic) Update(window *Window) (SceneId, any) {
	nextSceneId := scene.GetId()
	var levelData any = nil
	if scene.level.status != Stopped {
		scene.level.update(window)
		if scene.level.status == Finished { // TODO: this needs to be elaborate - is it a win, is it a loss?
			nextSceneId = Transition
			levelData = &scene.level
		}
	}
	return nextSceneId, levelData
}

func (scene *SceneLevelsBasic) Draw(window *Window) {
	scene.level.draw(window)
}

func (scene *SceneLevelsBasic) Teardown(window *Window) {

}
