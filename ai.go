package main

import (
	"cmp"
	"slices"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type searchPair struct {
	actor, target *Stone
	score         float32
}

func compareSearchPairs(p1, p2 searchPair) int {
	return cmp.Compare(p2.score, p1.score)
}

// / sufficiently "smart" AI
// / searches the options based on:
// / - proximity
// / - life state of the target stone
// / - life state of the hitting stone
// / - whether own stone will be hit in the process
// / - whether stone will richochet
func cpuSearchBestOption(level *Level, window *Window) (*Stone, *Stone) {
	searchPairs := []searchPair{}
	for i := range level.stones {
		actor := &level.stones[i]
		if actor.isDead || actor.playerId == PlayerOne { // TODO: we should have better ways to indicate the opponent
			continue
		}
		for j := range level.stones {
			if i == j {
				continue
			}

			target := &level.stones[j]
			if target.isDead || target.playerId == PlayerTwo { // TODO: better way to indicate the attacking player
				continue
			}

			searchPairs = append(searchPairs, searchPair{
				actor:  actor,
				target: target,
			})
		}
	}

	screenDiagonalSize := window.GetScreenDiagonal()

	for pi := range searchPairs {
		pair := &(searchPairs[pi])
		actor, target := pair.actor, pair.target

		ssOrigin := rl.Vector2Subtract(actor.pos, target.pos)
		angle := 2 * rl.Vector2LineAngle(rl.Vector2Normalize(ssOrigin), rl.NewVector2(1, 0))

		aTop := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, -actor.radius), -angle), actor.pos)
		aBottom := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, actor.radius), -angle), actor.pos)

		tTop := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, -target.radius), -angle), target.pos)
		tBottom := rl.Vector2Add(rl.Vector2Rotate(rl.NewVector2(0, target.radius), -angle), target.pos)

		hitsOwn := false
		richochets := false

		for i := range level.stones {
			stone := &level.stones[i]
			if stone.isDead || (stone == actor || stone == target) {
				continue
			}
			// line 1 check aTop - tTop
			if rl.CheckCollisionCircleLine(stone.pos, stone.radius, aTop, tTop) {
				hitsOwn = stone.playerId == PlayerTwo // better way needed?
				richochets = true
			}
			// line 2 check aBottom - tBottom
			if rl.CheckCollisionCircleLine(stone.pos, stone.radius, aBottom, tBottom) {
				hitsOwn = stone.playerId == PlayerTwo
				richochets = true
			}

			// line 3 check center to center
			if rl.CheckCollisionCircleLine(stone.pos, stone.radius, actor.pos, target.pos) {
				hitsOwn = stone.playerId == PlayerTwo
				richochets = true
			}

			if hitsOwn && richochets {
				break
			}
		}

		distance := rl.Vector2Distance(actor.pos, target.pos) / screenDiagonalSize

		pair.score -= distance
		if hitsOwn {
			pair.score += -1
		}

		if richochets {
			pair.score += -0.5
		}

		if actor.life <= 5 {
			pair.score += -0.5
		}

		if target.life <= 10 {
			pair.score += 1
		}
	}

	slices.SortFunc(searchPairs, compareSearchPairs)

	if len(searchPairs) > 0 {
		pair := searchPairs[0]
		return pair.actor, pair.target
	}
	return nil, nil
}
