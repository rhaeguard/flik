//go:build release

package main

// init - the whole purpose of this is to make sure that the executable files by default are in fullscreen mode
func init() {
	IsFullscreen = true
}
