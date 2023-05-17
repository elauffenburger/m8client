package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

type controllerContext struct {
	logger   *log.Logger
	renderer *renderer
}

type controller struct {
	logger *log.Logger

	renderer *renderer
	slip     *slipReader
	device   *os.File

	lastInput   normalInputCmd
	inputReader inputReader
}

func (c controller) enableAndResetDisplay() error {
	if _, err := c.device.Write([]byte{'E', 'R'}); err != nil {
		return errors.Wrap(err, "error resetting display")
	}

	return nil
}

func (c *controller) nextCmds() ([]cmd, error) {
	buf, err := c.slip.read(c.device)
	if err != nil {
		return nil, err
	}

	packets, err := c.slip.decode(buf)
	if err != nil {
		return nil, err
	}

	var cmds []cmd
	for _, packet := range packets {
		cmd, err := c.slip.decodeCommand(packet)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding packet %v as command", packet)
		}

		cmds = append(cmds, cmd)
	}

	return cmds, nil
}

func (c *controller) executeCmd(cmd cmd) error {
	if err := cmd.execute(&controllerContext{c.logger, c.renderer}); err != nil {
		return errors.Wrap(err, "error executing command")
	}

	// Just assume the renderer is dirty now.
	c.renderer.dirty = true

	return nil
}

type errQuitRequested struct{}

func (errQuitRequested) Error() string {
	return "quit requested"
}

func (c *controller) sendInput() error {
	input, err := c.inputReader.getInput()
	if err != nil {
		return errors.Wrap(err, "error updating input")
	}

	switch val := input.(type) {
	case normalInputCmd:
		// If nothing's changed, bail.
		if c.lastInput == val {
			return nil
		}

		// Update last input.
		c.lastInput = val

		// Send input.
		if _, err := c.device.Write([]byte{'C', byte(val)}); err != nil {
			return errors.Wrap(err, "error sending input")
		}

		return nil

	case fullscreenInputCmd:
		c.renderer.toggleFullscreen()
		return nil

	case exitInputCmd:
		// todo: is this right? should we do something better?
		return errQuitRequested{}

	default:
		return errors.Errorf("unknown command %v", val)
	}
}

func (c *controller) render() error {
	return c.renderer.render()
}
