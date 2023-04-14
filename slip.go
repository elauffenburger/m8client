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

func readSLIP(file *os.File) ([]byte, error) {
	buf := make([]byte, 4*1024)

	_, err := file.Read(buf)
	if err != nil {
		return nil, errors.Wrap(err, "error reading SLIP data from file")
	}

	return buf, nil
}

func decodeSLIP(data []byte) ([]slipPacket, error) {
	var (
		packet  slipPacket
		packets []slipPacket
		escaped = false
	)

	for _, ch := range data {
		switch ch {
		case slipEnd:
			if len(packet) > 0 {
				packets = append(packets, packet)
				packet = nil
			}

			continue

		case slipEsc:
			escaped = true
			continue

		case slipEscEnd:
			if escaped {
				escaped = false
				packet = append(packet, slipEnd)
			}

			continue

		case slipEscEsc:
			if escaped {
				escaped = false
				packet = append(packet, slipEsc)
			}

			continue

		default:
			packet = append(packet, ch)
		}
	}

	if len(packet) != 0 {
		return nil, errSlipDecodingFailed{packet}
	}

	return packets, nil
}
