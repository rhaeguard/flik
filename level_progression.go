package main

var InitialLevel = LevelBasic

var LevelProgression = map[SceneId]SceneId{
	LevelBasic:     LevelBordered,
	LevelBordered:  LevelTimeLimit,
	LevelTimeLimit: LevelBasic,
}
