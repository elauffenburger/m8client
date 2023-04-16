package main

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/pkg/errors"
	"github.com/veandco/go-sdl2/sdl"
)

type cmd interface {
	execute(*controllerContext) error
}

func decodeInt16(data []byte) int16 {
	return int16(binary.LittleEndian.Uint16(data))
}

type position struct {
	x, y int16
}

func decodePosition(data []byte) position {
	return position{decodeInt16(data[0:2]), decodeInt16(data[2:4])}
}

type size struct {
	width, height int16
}

func decodeSize(data []byte) size {
	return size{decodeInt16(data[0:2]), decodeInt16(data[2:4])}
}

type color struct {
	r, g, b uint8
}

func decodeColor(data []byte) color {
	return color{data[0], data[1], data[2]}
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

// decodeCommand decodes the given M8 SLIP command packet
func decodeCommand(packet []byte) (cmd, error) {
	n := len(packet)
	if n == 0 {
		return nil, errors.New("empty packet")
	}

	opcode := packet[0]
	switch opcode {

	// 253 (0xFD) - Draw character command:
	//    12 bytes. char c, int16 x position, int16 y position, uint8 r, uint8 g, uint8 b, uint8 r_background, uint8 g_background, uint8 b_background
	case drawCharacteOpCode:
		if n != 12 {
			return nil, errors.WithStack(errInvalidCmdLen{"draw character", 12, packet})
		}

		return DrawCharCmd{packet[1], decodePosition(packet[2:]), decodeColor(packet[6:]), decodeColor(packet[9:])}, nil

	// 254 (0xFE) - Draw rectangle command:
	//    12 bytes. int16 x position, int16 y position, int16 width, int16 height, uint8 r, uint8 g, uint8 b
	case drawRectOpCode:
		if n != 12 {
			return nil, errors.WithStack(errInvalidCmdLen{"draw rect", 12, packet})
		}

		return DrawRectCmd{decodePosition(packet[1:]), decodeSize(packet[5:]), decodeColor(packet[9:])}, nil

	// 252 (0xFC) - Draw oscilloscope waveform command:
	//    zero bytes if off - uint8 r, uint8 g, uint8 b, followed by 320 byte value array containing the waveform
	case drawOscilloscopeWaveformOpCode:
		if n < 4 {
			return nil, errInvalidCmdLen{"draw osc wave", 4, packet}
		}

		if waveformLength := n - 4; waveformLength != 0 && waveformLength != int(screenWidth) {
			return nil, errors.WithStack(errInvalidCmdLen{"draw osc wave data", int(screenWidth), packet})
		}

		return DrawOscWaveformCmd{decodeColor(packet[1:]), packet[4:]}, nil

	// 251 (0xFB) - Joypad key pressed state (hardware M8 only)
	//    - sends the keypress state as a single byte in hardware pin order: LEFT|UP|DOWN|SELECT|START|RIGHT|OPT|EDIT
	case joypadKeyPressedStateOpCode:
		if n != 3 {
			return nil, errors.WithStack(errInvalidCmdLen{"joypad key pressed", 3, packet})
		}

		return JoypadKeyPressedCmd{packet[1]}, nil

	default:
		return nil, errUnknownCmd{opcode}
	}
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

	if c.pos.x == 0 && c.pos.y == 0 && c.size.width == int16(screenWidth) && c.size.height == int16(screenHeight) {
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
		W: screenWidth,
		H: screenHeight / 10,
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
	ctrlCtx.logger.Printf("joypad key pressed: %v", c)

	// TODO: impl
	return nil
}
