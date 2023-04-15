package main

import (
	"math"

	"github.com/pkg/errors"
	"github.com/veandco/go-sdl2/sdl"
)

type renderer struct {
	window   *sdl.Window
	renderer *sdl.Renderer
	font     *sdl.Texture

	dirty    bool
	bgColor  color
	waveform [screenWidth]sdl.Point
}

func newRenderer(width, height int32) (*renderer, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, errors.Wrap(err, "error initializing sdl")
	}

	window, err := sdl.CreateWindow(
		"M8",
		sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED,
		width, height,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		return nil, errors.Wrap(err, "error creating window")
	}

	sdlRenderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
	if err != nil {
		return nil, errors.Wrap(err, "error creating renderer")
	}

	if err := sdlRenderer.SetLogicalSize(screenWidth, screenHeight); err != nil {
		return nil, errors.Wrap(err, "error setting renderer logical size")
	}

	font, err := createFont(sdlRenderer)
	if err != nil {
		return nil, errors.Wrap(err, "error initializing font for renderer")
	}

	return &renderer{
		window,
		sdlRenderer,
		font,
		false,
		color{},
		[screenWidth]sdl.Point{},
	}, nil
}

func createFont(renderer *sdl.Renderer) (*sdl.Texture, error) {
	surface, err := sdl.CreateRGBSurfaceWithFormat(0, fontWidth, fontHeight, 32, sdl.PIXELFORMAT_ARGB8888)
	if err != nil {
		return nil, errors.Wrap(err, "error creating surface for font")
	}
	defer surface.Free()

	pixels := surface.Pixels()

	l := int(surface.W*surface.H) / 8
	for i := 0; i < l; i++ {
		p := fontData[i]
		for j := 0; j < 8; j++ {
			var c byte
			if p&(1<<j) == 0 {
				c = math.MaxUint8
			}

			// Set all 4 color components (ARGB)
			for k := 0; k < 4; k++ {
				pixels[(i*8+j)*4+k] = c
			}
		}
	}

	font, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return nil, errors.Wrap(err, "error creating texture for font")
	}

	if err := font.SetBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, errors.Wrap(err, "error setting blendmode on font")
	}

	return font, nil
}

func (r *renderer) toggleFullscreen() {
	// TODO: impl
	panic("not implemented")
}

func (r *renderer) render() error {
	if !r.dirty {
		return nil
	}

	r.renderer.Present()
	return nil
}
