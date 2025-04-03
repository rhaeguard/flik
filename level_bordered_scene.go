package main

type SceneLevelsBordered struct {
	level         Level
	levelSettings LevelSettings
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
	}
}

func (scene *SceneLevelsBordered) Init(data any, window *Window) {
	// init
	level := newLevel(scene.levelSettings)
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
