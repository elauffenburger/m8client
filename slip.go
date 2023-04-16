package main

import (
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

func (r *slipReader) read(dev *os.File) ([]byte, error) {
	buf := make([]byte, 4*1024)

	n, err := dev.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "error reading SLIP data from device")
	}

	return buf[:n], nil
}

func (r *slipReader) decode(data []byte) ([]slipPacket, error) {
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
