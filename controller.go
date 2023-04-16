package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/veandco/go-sdl2/sdl"
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
	input    uint8
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

func (c *controller) getInput() (inpt input, changed bool, err error) {
	ev := sdl.PollEvent()
	switch ev := ev.(type) {
	case *sdl.KeyboardEvent:
		if ev.Type == sdl.KEYUP {
			switch ev.Keysym.Sym {
			case sdl.K_RETURN:
				if ev.Keysym.Mod&sdl.KMOD_ALT > 0 {
					c.renderer.toggleFullscreen()
					return 0, false, nil
				}

			case sdl.K_q:
				return 0, false, errQuitRequested{}
			}
		}

		var key input

		switch ev.Keysym.Sym {
		case sdl.K_RIGHT, sdl.K_KP_6:
			key = keyRight
		case sdl.K_LEFT, sdl.K_KP_4:
			key = keyLeft
		case sdl.K_UP, sdl.K_KP_8:
			key = keyUp
		case sdl.K_DOWN, sdl.K_KP_2:
			key = keyDown
		case sdl.K_x, sdl.K_m, sdl.K_LCTRL, sdl.K_RCTRL:
			key = keyEdit
		case sdl.K_z, sdl.K_n, sdl.K_LALT, sdl.K_RALT:
			key = keyOpt
		case sdl.K_SPACE:
			key = keyStart
		case sdl.K_LSHIFT, sdl.K_RSHIFT:
			key = keySelect
		}

		if key == 0 {
			return 0, false, nil
		}

		oldInput := c.input

		if ev.State == sdl.PRESSED {
			c.input |= uint8(key)
		} else {
			// Go does not have a bitwise negation operator
			c.input &= 255 ^ uint8(key)
		}

		return input(c.input), oldInput != c.input, nil
	}

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

func (c *controller) render() error {
	return c.renderer.render()
}
