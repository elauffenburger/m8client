package main

import (
	"log"
	"m8client/input"
	"os"

	"github.com/pkg/errors"
)

type controllerContext struct {
	logger   *log.Logger
	renderer *renderer
}

type slipRdr interface {
	Read(*os.File) ([]byte, error)
	Decode([]byte) ([]slipPacket, error)
	DecodeCommand([]byte) (cmd, error)
}

type controller struct {
	logger *log.Logger

	renderer *renderer
	slip     slipRdr
	device   *os.File

	lastInput   input.CmdKey
	inputReader inputReader
}

func (c controller) enableAndResetDisplay() error {
	if _, err := c.device.Write([]byte{'E', 'R'}); err != nil {
		return errors.Wrap(err, "error resetting display")
	}

	return nil
}

func (c *controller) nextCmds() ([]cmd, error) {
	buf, err := c.slip.Read(c.device)
	if err != nil {
		return nil, err
	}

	packets, err := c.slip.Decode(buf)
	if err != nil {
		return nil, err
	}

	var cmds []cmd
	for _, packet := range packets {
		cmd, err := c.slip.DecodeCommand(packet)
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
	inpt, err := c.inputReader.GetInput()
	if err != nil {
		return errors.Wrap(err, "error updating input")
	}

	switch val := inpt.(type) {
	case input.CmdKey:
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

	case input.CmdRequestFullScreen:
		c.renderer.toggleFullscreen()
		return nil

	case input.CmdRequestExit:
		// todo: is this right? should we do something better?
		return errQuitRequested{}

	default:
		return errors.Errorf("unknown command %v", val)
	}
}

func (c *controller) render() error {
	return c.renderer.render()
}
