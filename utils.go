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

var CpuPlayerPalette1 = PlayerColorPalette{
	primaryColor:   rl.NewColor(133, 90, 92, 255),
	outerRingColor: rl.NewColor(102, 16, 31, 255),
	lifeColor:      rl.NewColor(255, 250, 255, 255),
	rocketColor:    rl.NewColor(129, 13, 32, 255),
}

var HumanPlayerPalette1 = PlayerColorPalette{
	primaryColor:   rl.NewColor(55, 113, 142, 255),
	outerRingColor: rl.NewColor(37, 78, 112, 255),
	lifeColor:      rl.NewColor(255, 250, 255, 255),
	rocketColor:    rl.SkyBlue,
}

func getPlayer(label string, palette PlayerColorPalette, isCpu bool) PlayerSettings {
	return PlayerSettings{
		label:          label,
		primaryColor:   palette.primaryColor,
		outerRingColor: palette.outerRingColor,
		lifeColor:      palette.lifeColor,
		rocketColor:    palette.rocketColor,
		isCpu:          isCpu,
	}
}

func colorToInt64(color rl.Color) int64 {
	r, g, b, a := color.RGBA()
	// these individual colors are alpha-premultiplied
	// where we have 2 bytes per color.
	// (255, 0, 0, 255) is for example: 0xffff
	// we won't utilize this alpha value, but it can be as big as the color value itself.
	// this mask will basically get rid of the alpha byte
	mask := uint32(0xff00)

	r = (r & mask) << 16 // two bytes to the left, because red color is in 3rd byte from left (assuming 4 bytes), so it moves to the first byte
	g = (g & mask) << 8  // a byte to the left, 3rd byte => 2nd byte
	b = (b & mask)       // same position, the 3rd byte
	a = (a & mask) >> 8  // move right, 3rd => 4th

	return int64(r | g | b | a)
}
