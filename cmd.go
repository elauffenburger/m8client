package main

import (
	"fmt"
	"math"

	"github.com/veandco/go-sdl2/sdl"
)

type cmd interface {
	execute(*controllerContext) error
}

type position struct {
	x, y int16
}

type size struct {
	width, height int16
}

type color struct {
	r, g, b uint8
}

const (
	drawRectOpCode                 = 0xFE
	drawCharacteOpCode             = 0xFD
	drawOscilloscopeWaveformOpCode = 0xFC
	joypadKeyPressedStateOpCode    = 0xFB
)

type errUnknownCmd struct {
	command byte
}

func (e errUnknownCmd) Error() string {
	return fmt.Sprintf("unknown command byte: 0x%x", e.command)
}

type errInvalidCmdLen struct {
	cmdName     string
	expectedLen int
	buf         []byte
}

func (e errInvalidCmdLen) Error() string {
	return fmt.Sprintf("invalid %s packet; expected %d bytes; actual %d bytes; buffer: %x", e.cmdName, e.expectedLen, len(e.buf), e.buf)
}

type NoOpCmd struct{}

func (c NoOpCmd) execute(ctrlCtx *controllerContext) error {
	return nil
}

type DrawRectCmd struct {
	pos   position
	size  size
	color color
}

func (c DrawRectCmd) execute(ctrlCtx *controllerContext) error {
	var (
		renderer    = ctrlCtx.renderer
		sdlRenderer = ctrlCtx.renderer.renderer
	)

	if c.pos.x == 0 && c.pos.y == 0 && c.size.width == int16(m8ScreenWidth) && c.size.height == int16(m8ScreenHeight) {
		renderer.bgColor = c.color
	}

	if err := sdlRenderer.SetDrawColor(c.color.r, c.color.g, c.color.b, 0xff); err != nil {
		return err
	}

	renderRect := sdl.Rect{
		X: int32(c.pos.x),
		Y: int32(c.pos.y),
		W: int32(c.size.width),
		H: int32(c.size.height),
	}

	if err := sdlRenderer.FillRect(&renderRect); err != nil {
		return err
	}

	return nil
}

type DrawCharCmd struct {
	ch         byte
	pos        position
	foreground color
	background color
}

func (c DrawCharCmd) execute(ctrlCtx *controllerContext) error {
	var (
		renderer    = ctrlCtx.renderer
		sdlRenderer = renderer.renderer
		x           = int32(c.pos.x)
		y           = int32(c.pos.y)
	)

	if c.background != c.foreground {
		if err := sdlRenderer.SetDrawColor(c.background.r, c.background.g, c.background.b, math.MaxUint8); err != nil {
			return err
		}

		if err := sdlRenderer.FillRect(&sdl.Rect{
			X: x - 1,
			Y: y + 2,
			W: fontChWidth - 1,
			H: fontChHeight + 1,
		}); err != nil {
			return err
		}
	}

	if err := ctrlCtx.renderer.font.SetColorMod(c.foreground.r, c.foreground.g, c.foreground.b); err != nil {
		return err
	}

	var (
		row    = c.ch / fontChsPerRow
		column = c.ch % fontChsPerRow

		sourceRect = sdl.Rect{
			X: int32(column * 8),
			Y: int32(row * 8),
			W: fontChWidth,
			H: fontChHeight,
		}

		renderRect = sdl.Rect{
			X: x,
			Y: y + 3,
			W: fontChWidth,
			H: fontChHeight,
		}
	)

	if err := sdlRenderer.Copy(renderer.font, &sourceRect, &renderRect); err != nil {
		return err
	}

	return nil
}

type DrawOscWaveformCmd struct {
	color    color
	waveform []byte
}

func (c DrawOscWaveformCmd) execute(ctrlCtx *controllerContext) error {
	var (
		renderer    = ctrlCtx.renderer
		sdlRenderer = renderer.renderer
	)

	renderRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: m8ScreenWidth,
		H: m8ScreenHeight / 8,
	}

	if err := sdlRenderer.SetDrawColor(renderer.bgColor.r, renderer.bgColor.g, renderer.bgColor.b, math.MaxUint8); err != nil {
		return err
	}

	if err := sdlRenderer.FillRect(&renderRect); err != nil {
		return err
	}

	if err := sdlRenderer.DrawRect(&renderRect); err != nil {
		return err
	}

	if len(c.waveform) == 0 {
		return nil
	}

	if err := sdlRenderer.SetDrawColor(c.color.r, c.color.g, c.color.b, math.MaxUint8); err != nil {
		return err
	}

	for x, y := range c.waveform {
		renderer.waveform[x] = sdl.Point{X: int32(x), Y: int32(y)}
	}

	if err := sdlRenderer.DrawPoints(renderer.waveform[:]); err != nil {
		return err
	}

	return nil
}

type JoypadKeyPressedCmd struct {
	key byte
}

func (c JoypadKeyPressedCmd) execute(ctrlCtx *controllerContext) error {
	// TODO: impl
	return nil
}
