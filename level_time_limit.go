package main

type SceneLevelsTimeLimit struct {
	level          Level
	levelSettings  LevelSettings
	playerSettings [TotalPlayerCount]PlayerSettings
}

func NewSceneLevelsTimeLimit(window *Window) SceneLevelsTimeLimit {
	return SceneLevelsTimeLimit{
		levelSettings: LevelSettings{
			sceneId:             LevelTimeLimit,
			stonesPerPlayer:     5,
			isTimed:             true,
			totalSecondsAllowed: 45,
			backgroundColor:     BG_COLOR,
		}, playerSettings: [TotalPlayerCount]PlayerSettings{
			PlayerOne: getPlayer("you", HumanPlayerPalette1, false),
			PlayerTwo: getPlayer("cpu", CpuPlayerPalette1, true),
		},
	}
}

func (scene *SceneLevelsTimeLimit) Init(data any, window *Window) {
	// init
	level := newLevel(scene.levelSettings, scene.playerSettings)
	level.init(window)
	scene.level = level
}

func (scene *SceneLevelsTimeLimit) GetId() SceneId {
	return LevelTimeLimit
}

func (scene *SceneLevelsTimeLimit) HandleUserInput(window *Window) {
	scene.level.handleUserInput(window)
}

func (scene *SceneLevelsTimeLimit) Update(window *Window) (SceneId, any) {
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

func (scene *SceneLevelsTimeLimit) Draw(window *Window) {
	scene.level.draw(window)
}

func (scene *SceneLevelsTimeLimit) Teardown(window *Window) {

}
