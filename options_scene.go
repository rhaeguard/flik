package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type SceneOptions struct {
	nextSceneId        SceneId
	screenSizes        []string
	screenSizesIx      int32
	screenSizesEnabled bool

	fullscreen        []string
	fullscreenIx      int32
	fullscreenEnabled bool

	musicVolume float32
	sfxVolume   float32

	saveClicked bool
	backClicked bool
}

func NewSceneOptions() SceneOptions {

	return SceneOptions{
		screenSizes: []string{
			" 1280x720",
			" 1440x810",
			" 1920x1080",
		},
		screenSizesEnabled: false,
		fullscreen: []string{
			" no",
			" yes",
		},
		fullscreenEnabled: false,
	}
}

func (scene *SceneOptions) GetId() SceneId {
	return Options
}

func (scene *SceneOptions) Init(data any, window *Window) {
	scene.nextSceneId = scene.GetId()
	fullscreenIx := int32(0) // no
	if window.fullscreen {
		fullscreenIx = 1 // yes
	}
	scene.fullscreenIx = fullscreenIx

	w, h := window.GetScreenDimensions()

	current := fmt.Sprintf(" %dx%d", int(w), int(h))

	ix := slices.Index(scene.screenSizes, current)

	if ix == -1 {
		scene.screenSizesIx = int32(len(scene.screenSizes))
		scene.screenSizes = append(scene.screenSizes, current)
	} else {
		scene.screenSizesIx = int32(ix)
	}

	maxSize := fmt.Sprintf(" %dx%d", window.maxScreenWidth, window.maxScreenHeight)

	if !slices.Contains(scene.screenSizes, maxSize) {
		scene.screenSizes = append(scene.screenSizes, maxSize)
	}

	scene.musicVolume = window.musicVolume
	scene.sfxVolume = window.sfxVolume
}

func (scene *SceneOptions) HandleUserInput(window *Window) {

}

func (scene *SceneOptions) Update(window *Window) (SceneId, any) {
	if scene.saveClicked {
		fullscreen := strings.Trim(scene.fullscreen[scene.fullscreenIx], " ")
		shouldBeFullscreen := fullscreen == "yes"
		if shouldBeFullscreen {
			rl.SetWindowState(rl.FlagFullscreenMode)
		} else {
			rl.ClearWindowState(rl.FlagFullscreenMode)
		}
		window.fullscreen = shouldBeFullscreen

		pieces := strings.Split(strings.Trim(scene.screenSizes[scene.screenSizesIx], " "), "x")
		newWidth, _ := strconv.Atoi(pieces[0])
		newHeight, _ := strconv.Atoi(pieces[1])
		rl.SetWindowSize(newWidth, newHeight)
		rl.SetWindowPosition(100, 100)
		window.width = int32(newWidth)
		window.height = int32(newHeight)

		//
		window.musicVolume = scene.musicVolume
		window.sfxVolume = scene.sfxVolume

		game.status = GameUninitialized
	}

	if scene.backClicked {
		return Main, nil
	}

	return scene.nextSceneId, nil
}

func (scene *SceneOptions) Draw(window *Window) {
	// draw background
	rl.ClearBackground(BG_COLOR)

	ScreenWidth, ScreenHeight := window.GetScreenDimensions()

	gui.SetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR, colorToInt64(BG_COLOR))
	gui.SetStyle(gui.DEFAULT, gui.BASE_COLOR_NORMAL, colorToInt64(BG_COLOR))
	gui.SetStyle(gui.DEFAULT, gui.BASE_COLOR_FOCUSED, colorToInt64(BG_COLOR))
	gui.SetStyle(gui.DEFAULT, gui.BASE_COLOR_PRESSED, colorToInt64(dimWhite(120)))

	gui.SetStyle(gui.DEFAULT, gui.BORDER_COLOR_NORMAL, colorToInt64(BG_COLOR))
	gui.SetStyle(gui.DEFAULT, gui.BORDER_COLOR_FOCUSED, colorToInt64(dimWhite(200)))
	gui.SetStyle(gui.DEFAULT, gui.BORDER_COLOR_PRESSED, colorToInt64(dimWhite(255)))

	gui.SetStyle(gui.DEFAULT, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.DEFAULT, gui.TEXT_COLOR_FOCUSED, colorToInt64(dimWhite(200)))
	gui.SetStyle(gui.DEFAULT, gui.TEXT_COLOR_PRESSED, colorToInt64(dimWhite(255)))

	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_CENTER))

	gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, int64(FontSize/3))
	gui.SetStyle(gui.DEFAULT, gui.TEXT_SPACING, int64(FontSize/60))
	gui.Label(rl.NewRectangle(0, ScreenHeight*0.125, ScreenWidth, ScreenHeight/5), "OPTIONS")

	// reset it back to the original
	gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, int64(FontSize/10))
	gui.SetStyle(gui.DEFAULT, gui.TEXT_SPACING, int64(FontSize/200))

	yAxis := ScreenHeight / 3
	gui.SetStyle(gui.DROPDOWNBOX, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_LEFT))
	x := strings.Join(scene.screenSizes, ";")
	if gui.DropdownBox(rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20), x, &scene.screenSizesIx, scene.screenSizesEnabled) {
		scene.screenSizesEnabled = !scene.screenSizesEnabled
	}

	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_RIGHT))
	gui.Label(rl.NewRectangle(0, yAxis, ScreenWidth*0.45, ScreenHeight/20), "resolution")

	yAxis += ScreenHeight / 20

	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_RIGHT))
	gui.Label(rl.NewRectangle(0, yAxis, ScreenWidth*0.45, ScreenHeight/20), "fullscreen")

	if !scene.screenSizesEnabled {
		gui.SetStyle(gui.DROPDOWNBOX, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_LEFT))
		x = strings.Join(scene.fullscreen, ";")
		if gui.DropdownBox(rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20), x, &scene.fullscreenIx, scene.fullscreenEnabled) {
			scene.fullscreenEnabled = !scene.fullscreenEnabled
		}
	}

	yAxis += ScreenHeight / 20

	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_RIGHT))
	gui.Label(rl.NewRectangle(0, yAxis, ScreenWidth*0.45, ScreenHeight/20), "music")

	if !scene.screenSizesEnabled && !scene.fullscreenEnabled {
		scene.musicVolume = gui.SliderBar(
			rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20),
			"",
			"",
			scene.musicVolume,
			0,
			1,
		)
	}

	yAxis += ScreenHeight / 20

	gui.SetStyle(gui.LABEL, gui.TEXT_COLOR_NORMAL, colorToInt64(dimWhite(120)))
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, int64(gui.TEXT_ALIGN_RIGHT))
	gui.Label(rl.NewRectangle(0, yAxis, ScreenWidth*0.45, ScreenHeight/20), "sound")

	if !scene.screenSizesEnabled && !scene.fullscreenEnabled {
		scene.sfxVolume = gui.SliderBar(
			rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20),
			"",
			"",
			scene.sfxVolume,
			0,
			1,
		)
	}

	yAxis += ScreenHeight / 5

	scene.saveClicked = gui.Button(
		rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20),
		"save",
	)

	yAxis += ScreenHeight / 20

	scene.backClicked = gui.Button(
		rl.NewRectangle(ScreenWidth/2, yAxis, ScreenWidth*0.25, ScreenHeight/20),
		"back",
	)
}

func (scene *SceneOptions) Teardown(window *Window) {
}
