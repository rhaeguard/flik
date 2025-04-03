package main

type SceneLevelsBasic struct {
	level         Level
	levelSettings LevelSettings
}

func NewSceneLevelsBasic() SceneLevelsBasic {
	return SceneLevelsBasic{
		levelSettings: LevelSettings{
			backgroundColor: BG_COLOR,
		},
	}
}

func (scene *SceneLevelsBasic) Init(data any, window *Window) {
	// init
	level := newLevel(scene.levelSettings)
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
			nextSceneId = LevelBordered
			levelData = scene.level
		}
	}
	return nextSceneId, levelData
}

func (scene *SceneLevelsBasic) Draw(window *Window) {
	scene.level.draw(window)
}

func (scene *SceneLevelsBasic) Teardown(window *Window) {

}
