package main

type SceneLevelsBasic struct {
	level Level
}

func NewSceneLevelsBasic() SceneLevelsBasic {
	return SceneLevelsBasic{}
}

func (scene *SceneLevelsBasic) Init(data any, window *Window) {
	// init
	level := newLevel()
	level.init(window)
	scene.level = level
}

func (scene *SceneLevelsBasic) GetId() SceneId {
	return Levels
}
func (scene *SceneLevelsBasic) HandleUserInput(window *Window) {
	handleUserInput(&scene.level, window)
}
func (scene *SceneLevelsBasic) Update(window *Window) (SceneId, any) {
	nextSceneId := scene.GetId()
	var levelData any = nil
	if scene.level.status != Stopped {
		nextSceneId, levelData = update(&scene.level, window)
	}
	return nextSceneId, levelData
}
func (scene *SceneLevelsBasic) Draw(window *Window) {
	draw(&scene.level, window)

}
func (scene *SceneLevelsBasic) Teardown(window *Window) {

}
