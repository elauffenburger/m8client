package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

type controllerContext struct {
	logger *log.Logger
}

type controller struct {
	logger *log.Logger

	device *os.File
	input  uint8
}

func (c controller) enableAndResetDisplay() error {
	if _, err := c.device.Write([]byte{'E', 'R'}); err != nil {
		return errors.Wrap(err, "error resetting display")
	}

	return nil
}

func (c controller) nextCmds() ([]cmd, error) {
	buf, err := readSLIP(c.device)
	if err != nil {
		return nil, err
	}

	packets, err := decodeSLIP(buf)
	if err != nil {
		return nil, err
	}

	var cmds []cmd
	for _, packet := range packets {
		cmd, err := decodeCommand(packet)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding packet %v as command", packet)
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func (c *controller) executeCmd(cmd cmd) error {
	return cmd.execute(&controllerContext{c.logger})
}

func (c *controller) getInput() (input input, changed bool, err error) {
	// TODO: impl
	return 0, false, nil
}

func (c *controller) sendInput() error {
	input, changed, err := c.getInput()
	if err != nil {
		return errors.Wrap(err, "error updating input")
	}

	if !changed {
		return nil
	}

	if _, err := c.device.Write([]byte{'C', byte(input)}); err != nil {
		return errors.Wrap(err, "error sending input")
	}

	return nil
}
