package main

type SceneId uint8

const (
	// scenes
	Main           SceneId = iota
	LevelBasic     SceneId = iota
	LevelBordered  SceneId = iota
	LevelTimeLimit SceneId = iota
	Controls       SceneId = iota
	Transition     SceneId = iota
	Quit           SceneId = iota
)

type Scene interface {
	GetId() SceneId
	Init(data any, window *Window)
	HandleUserInput(window *Window)
	Update(window *Window) (SceneId, any)
	Draw(window *Window)
	Teardown(window *Window)
}
