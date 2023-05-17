package main

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/pkg/errors"
)

const slipEnd = 0xC0
const slipEsc = 0xDB
const slipEscEnd = 0xDC
const slipEscEsc = 0xDD

type slipPacket []byte

type slipReader struct {
	escaped bool
	prev    []byte
}

func (r *slipReader) Read(dev *os.File) ([]byte, error) {
	buf := make([]byte, 4*1024)

	n, err := dev.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "error reading SLIP data from device")
	}

	return buf[:n], nil
}

func (r *slipReader) Decode(data []byte) ([]slipPacket, error) {
	var (
		packets []slipPacket
		packet  slipPacket = r.prev
	)

	for i := 0; i < len(data); i++ {
		ch := data[i]

		switch ch {
		case slipEnd:
			if len(packet) > 0 {
				packets = append(packets, packet)
			}

			packet = nil

			continue

		case slipEsc:
			r.escaped = true

			i++
			switch ch := data[i]; ch {
			case slipEscEnd:
				packet = append(packet, slipEnd)

			case slipEscEsc:
				packet = append(packet, slipEsc)

			default:
				return nil, errors.Errorf("unexpected escaped ch %x", ch)
			}

			continue

		default:
			packet = append(packet, ch)
		}
	}

	r.prev = packet

	return packets, nil
}

// DecodeCommand decodes the given M8 SLIP command packet
func (r slipReader) DecodeCommand(packet []byte) (cmd, error) {
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

		return DrawCharCmd{packet[1], r.decodePosition(packet[2:]), r.decodeColor(packet[6:]), r.decodeColor(packet[9:])}, nil

	// 254 (0xFE) - Draw rectangle command:
	//    12 bytes. int16 x position, int16 y position, int16 width, int16 height, uint8 r, uint8 g, uint8 b
	case drawRectOpCode:
		if n != 12 {
			return nil, errors.WithStack(errInvalidCmdLen{"draw rect", 12, packet})
		}

		return DrawRectCmd{r.decodePosition(packet[1:]), r.decodeSize(packet[5:]), r.decodeColor(packet[9:])}, nil

	// 252 (0xFC) - Draw oscilloscope waveform command:
	//    zero bytes if off - uint8 r, uint8 g, uint8 b, followed by 320 byte value array containing the waveform
	case drawOscilloscopeWaveformOpCode:
		if n < 4 {
			return nil, errInvalidCmdLen{"draw osc wave", 4, packet}
		}

		if waveformLength := n - 4; waveformLength != 0 && waveformLength != int(m8ScreenWidth) {
			return nil, errors.WithStack(errInvalidCmdLen{"draw osc wave data", int(m8ScreenWidth), packet})
		}

		return DrawOscWaveformCmd{r.decodeColor(packet[1:]), packet[4:]}, nil

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

func (r slipReader) decodeInt16(data []byte) int16 {
	return int16(binary.LittleEndian.Uint16(data))
}

func (r slipReader) decodePosition(data []byte) position {
	return position{r.decodeInt16(data[0:2]), r.decodeInt16(data[2:4])}
}

func (r slipReader) decodeSize(data []byte) size {
	return size{r.decodeInt16(data[0:2]), r.decodeInt16(data[2:4])}
}

func (r slipReader) decodeColor(data []byte) color {
	return color{data[0], data[1], data[2]}
}

type safeSlipReader struct {
	logger *log.Logger
	reader slipRdr
}

func (r *safeSlipReader) Read(dev *os.File) ([]byte, error) {
	buf, err := r.reader.Read(dev)
	if err != nil {
		return nil, err
	}

	r.logger.Printf("read bytes: %v\n", buf)

	return buf, nil
}

func (r *safeSlipReader) Decode(data []byte) ([]slipPacket, error) {
	packets, err := r.reader.Decode(data)
	if err != nil {
		return nil, err
	}

	r.logger.Printf("decoded packets: %v\n", packets)

	return packets, err
}

// DecodeCommand decodes the given M8 SLIP command packet
func (r *safeSlipReader) DecodeCommand(packet []byte) (cmd, error) {
	cmd, err := r.reader.DecodeCommand(packet)
	if err != nil {
		r.logger.Printf("error decoding packet: %s; ignoring.\npacket: %v\n", err, packet)

		return &NoOpCmd{}, nil
	}

	r.logger.Printf("decoded command: %v\n", cmd)
	return cmd, nil
}
