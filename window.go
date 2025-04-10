package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Window struct {
	title           string
	width           int32
	height          int32
	fullscreen      bool
	musicVolume     float32
	sfxVolume       float32
	maxScreenWidth  int32
	maxScreenHeight int32
}

func (c *Window) GetScreenDimensions() (float32, float32) {
	if c.fullscreen {
		w := float32(rl.GetScreenWidth())
		h := float32(rl.GetScreenHeight())
		return w, h
	} else {
		return float32(c.width), float32(c.height)
	}
}

func (c *Window) GetScreenBoundary() rl.Rectangle {
	screenWidth, screenHeight := c.GetScreenDimensions()
	screenRect := rl.NewRectangle(0, 0, screenWidth, screenHeight)
	return screenRect
}

// GetScreenBoundaryLines - returns the boundary lines in the order of
// TOP, LEFT, BOTTOM, RIGHT
func (c *Window) GetScreenBoundaryLines(screenBoundary rl.Rectangle) [4][2]rl.Vector2 {
	topLeft := rl.NewVector2(screenBoundary.X, screenBoundary.Y)
	topRight := rl.NewVector2(screenBoundary.X+screenBoundary.Width, screenBoundary.Y)
	bottomLeft := rl.NewVector2(screenBoundary.X, screenBoundary.Y+screenBoundary.Height)
	bottomRight := rl.NewVector2(screenBoundary.X+screenBoundary.Width, screenBoundary.Y+screenBoundary.Height)

	lines := [4][2]rl.Vector2{
		{topLeft, topRight},
		{topLeft, bottomLeft},
		{bottomLeft, bottomRight},
		{topRight, bottomRight},
	}

	return lines
}

func (c *Window) GetScreenDiagonal() float32 {
	w, h := c.GetScreenDimensions()
	res := math.Sqrt(float64(w*w + h*h))
	return float32(res)
}
