package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

const slipEnd = 0xC0
const slipEsc = 0xDB
const slipEscEnd = 0xDC
const slipEscEsc = 0xDD

type slipPacket []byte

type errSlipDecodingFailed struct {
	remaining []byte
}

func (e errSlipDecodingFailed) Error() string {
	return fmt.Sprintf("SLIP decoding failed; remaining data: %v", e.remaining)
}

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

	for _, ch := range data {
		switch ch {
		case slipEnd:
			if len(packet) > 0 {
				packets = append(packets, packet)
			}

			packet = nil

			continue

		case slipEsc:
			r.escaped = true
			continue

		case slipEscEnd:
			if r.escaped {
				r.escaped = false
				packet = append(packet, slipEnd)
			}

			continue

		case slipEscEsc:
			if r.escaped {
				r.escaped = false
				packet = append(packet, slipEsc)
			}

			continue

		default:
			packet = append(packet, ch)
		}
	}

	r.prev = packet

	return packets, nil
}
