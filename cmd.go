package main

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
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

		return DrawCharacterCommand{packet[1], decodePosition(packet[2:]), decodeColor(packet[6:]), decodeColor(packet[9:])}, nil

	// 254 (0xFE) - Draw rectangle command:
	//    12 bytes. int16 x position, int16 y position, int16 width, int16 height, uint8 r, uint8 g, uint8 b
	case drawRectOpCode:
		if n != 12 {
			return nil, errors.WithStack(errInvalidCmdLen{"draw rect", 12, packet})
		}

		return DrawRectangleCommand{decodePosition(packet[1:]), decodeSize(packet[5:]), decodeColor(packet[9:])}, nil

	// 252 (0xFC) - Draw oscilloscope waveform command:
	//    zero bytes if off - uint8 r, uint8 g, uint8 b, followed by 320 byte value array containing the waveform
	case drawOscilloscopeWaveformOpCode:
		if n < 4 {
			return nil, errInvalidCmdLen{"draw osc wave", 4, packet}
		}

		if waveformLength := n - 4; waveformLength != 0 && waveformLength != screenWidth {
			return nil, errors.WithStack(errInvalidCmdLen{"draw osc wave data", screenWidth, packet})
		}

		return DrawOscilloscopeWaveformCommand{decodeColor(packet[1:]), packet[4:]}, nil

	// 251 (0xFB) - Joypad key pressed state (hardware M8 only)
	//    - sends the keypress state as a single byte in hardware pin order: LEFT|UP|DOWN|SELECT|START|RIGHT|OPT|EDIT
	case joypadKeyPressedStateOpCode:
		if n != 3 {
			return nil, errors.WithStack(errInvalidCmdLen{"joypad key pressed", 3, packet})
		}

		return JoypadKeyPressedStateCommand{packet[1]}, nil

	default:
		return nil, errUnknownCmd{opcode}
	}
}
