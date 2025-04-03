package main

var InitialLevel = LevelBasic

var LevelProgression = map[SceneId]SceneId{
	LevelBasic:    LevelBordered,
	LevelBordered: LevelBasic,
}
