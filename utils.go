package main

import (
	"image/color"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// line a and line b intersection
// reference: https://en.wikipedia.org/wiki/Line%E2%80%93line_intersection
func getLineToLineIntersectionPoint(as, ae, bs, be rl.Vector2) (rl.Vector2, bool) {
	uA := ((be.X-bs.X)*(as.Y-bs.Y) - (be.Y-bs.Y)*(as.X-bs.X)) / ((be.Y-bs.Y)*(ae.X-as.X) - (be.X-bs.X)*(ae.Y-as.Y))
	uB := ((ae.X-as.X)*(as.Y-bs.Y) - (ae.Y-as.Y)*(as.X-bs.X)) / ((be.Y-bs.Y)*(ae.X-as.X) - (be.X-bs.X)*(ae.Y-as.Y))

	if uA >= 0 && uA <= 1 && uB >= 0 && uB <= 1 {
		// intersection points
		x := as.X + (uA * (ae.X - as.X))
		y := as.Y + (uA * (ae.Y - as.Y))
		return rl.NewVector2(x, y), true
	}
	return rl.NewVector2(0, 0), false
}

func dimWhite(alpha uint8) color.RGBA {
	return rl.NewColor(255, 255, 255, alpha)
}
