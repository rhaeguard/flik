package main

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Window struct {
	fullscreen bool
	width      int32
	height     int32
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

func (c *Window) GetScreenDiagonal() float32 {
	w, h := c.GetScreenDimensions()
	res := math.Sqrt(float64(w*w + h*h))
	return float32(res)
}
