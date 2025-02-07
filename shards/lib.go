package main

import (
	"math"
	"math/rand"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func getPoint(angle float32) rl.Vector2 {
	eRadiusX := 250.0
	eRadiusY := 80.0

	theta := float64(angle * rl.Deg2rad)

	x := eRadiusX * math.Cos(theta)
	y := eRadiusY * math.Sin(theta)

	return rl.NewVector2(float32(x), float32(y))
}

func getAngles() []float32 {
	n := rand.Intn(4) + 3

	angles := []float32{}

	for i := 0; i < n; i++ {
		angle := rand.Float32() * 355.23
		angles = append(angles, angle)
	}

	slices.Sort(angles)

	return angles
}

func getPolygon(center rl.Vector2, angles []float32) []rl.Vector2 {
	slices.Sort(angles)

	polygonPoints := []rl.Vector2{}

	for _, angle := range angles {
		pt := getPoint(angle)
		pt = rl.Vector2Add(pt, center)
		polygonPoints = append(polygonPoints, pt)
	}

	return polygonPoints
}

func main() {
	rl.InitWindow(800, 450, "raylib [core] example - basic window")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	angles := getAngles()

	var polygon []rl.Vector2

	for !rl.WindowShouldClose() {
		if rl.IsKeyPressed(rl.KeySpace) {
			angles = getAngles()
		}

		for i := range angles {
			angles[i] += 10
		}
		polygon = getPolygon(rl.NewVector2(400, 225), angles)

		rl.BeginDrawing()

		rl.ClearBackground(rl.Black)

		rl.DrawEllipseLines(400, 225, 250, 80, rl.White)

		for _, p := range polygon {
			rl.DrawCircleV(p, 5, rl.Red)
		}

		l := len(polygon)

		for i := 0; i < l; i++ {
			p1 := polygon[i%l]
			p2 := polygon[(i+1)%l]
			rl.DrawLineV(p1, p2, rl.Red)
		}
		p0 := polygon[0]
		for i := 1; i < l-1; i++ {
			p1 := polygon[i%l]
			p2 := polygon[i+1]
			rl.DrawTriangle(p2, p1, p0, rl.Green)
		}

		rl.EndDrawing()
	}
}
