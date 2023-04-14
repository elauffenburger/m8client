package main

import (
	"encoding/binary"
	"errors"
	"fmt"
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

// DrawRectangleCommand
type DrawRectangleCommand struct {
	pos   position
	size  size
	color color
}

func (c DrawRectangleCommand) execute(ctrlCtx *controllerContext) error {
	ctrlCtx.logger.Printf("draw rect: %v\n", c)

	// TODO: impl
	return nil
}

// DrawCharacterCommand
type DrawCharacterCommand struct {
	c          byte
	pos        position
	foreground color
	background color
}

func (c DrawCharacterCommand) execute(ctrlCtx *controllerContext) error {
	ctrlCtx.logger.Printf("draw ch: %v\n", c)

	// TODO: impl
	return nil
}

// DrawOscilloscopeWaveformCommand
type DrawOscilloscopeWaveformCommand struct {
	color    color
	waveform []byte
}

func (c DrawOscilloscopeWaveformCommand) execute(ctrlCtx *controllerContext) error {
	ctrlCtx.logger.Printf("draw osc wav: %v\n", c)

	// TODO: impl
	return nil
}

// JoypadKeyPressedStateCommand
type JoypadKeyPressedStateCommand struct {
	key byte
}

func (c JoypadKeyPressedStateCommand) execute(ctrlCtx *controllerContext) error {
	ctrlCtx.logger.Printf("joypad key pressed: %v", c)

	// TODO: impl
	return nil
}

type errUnknownCmd struct {
	command byte
}

func (e errUnknownCmd) Error() string {
	return fmt.Sprintf("unknown command byte: 0x%x", e.command)
}

type errInvalidCmdLen struct {
	cmdName          string
	expected, actual int
}

func (e errInvalidCmdLen) Error() string {
	return fmt.Sprintf("invalid %s packet; expected %d bytes; actual %d bytes", e.cmdName, e.expected, e.actual)
}

// decodeCommand decodes the given M8 SLIP command packet
func decodeCommand(data []byte) (cmd, error) {
	n := len(data)
	if n == 0 {
		return nil, errors.New("empty packet")
	}

	opcode := data[0]
	switch opcode {

	// 253 (0xFD) - Draw character command:
	//    12 bytes. char c, int16 x position, int16 y position, uint8 r, uint8 g, uint8 b, uint8 r_background, uint8 g_background, uint8 b_background
	case drawCharacteOpCode:
		if n != 12 {
			return nil, errInvalidCmdLen{"draw character", 12, n}
		}

		return DrawCharacterCommand{data[1], decodePosition(data[2:]), decodeColor(data[6:]), decodeColor(data[9:])}, nil

	// 253 (0xFD) - Draw character command:
	//    12 bytes. char c, int16 x position, int16 y position, uint8 r, uint8 g, uint8 b, uint8 r_background, uint8 g_background, uint8 b_background
	case drawRectOpCode:
		if n != 12 {
			return nil, errInvalidCmdLen{"draw rect", 12, n}
		}

		return DrawRectangleCommand{decodePosition(data[1:]), decodeSize(data[5:]), decodeColor(data[9:])}, nil

	// 252 (0xFC) - Draw oscilloscope waveform command:
	//    zero bytes if off - uint8 r, uint8 g, uint8 b, followed by 320 byte value array containing the waveform
	case drawOscilloscopeWaveformOpCode:
		if n < 4 {
			return nil, errInvalidCmdLen{"draw osc wave", 4, n}
		}

		if waveformLength := n - 4; waveformLength != 0 && waveformLength != screenWidth {
			return nil, errInvalidCmdLen{"draw osc wave data", screenWidth, n}
		}

		return DrawOscilloscopeWaveformCommand{decodeColor(data[1:]), data[4:]}, nil

	// 251 (0xFB) - Joypad key pressed state (hardware M8 only)
	//    - sends the keypress state as a single byte in hardware pin order: LEFT|UP|DOWN|SELECT|START|RIGHT|OPT|EDIT
	case joypadKeyPressedStateOpCode:
		if n != 3 {
			return nil, errInvalidCmdLen{"joypad key pressed", 3, n}
		}

		return JoypadKeyPressedStateCommand{data[1]}, nil

	default:
		return nil, errUnknownCmd{opcode}
	}
}
