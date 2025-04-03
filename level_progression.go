package main

var LevelProgression = map[SceneId]SceneId{
	LevelBasic:    LevelBordered,
	LevelBordered: LevelBasic,
}
