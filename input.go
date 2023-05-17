package main

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	gpio "github.com/stianeikeland/go-rpio/v4"
	"github.com/veandco/go-sdl2/sdl"
)

type inputCmd interface {
	isInput()
}

type normalInputCmd uint8

func (normalInputCmd) isInput() {}

const (
	keyLeft   normalInputCmd = 1 << 7
	keyUp     normalInputCmd = 1 << 6
	keyDown   normalInputCmd = 1 << 5
	keySelect normalInputCmd = 1 << 4
	keyStart  normalInputCmd = 1 << 3
	keyRight  normalInputCmd = 1 << 2
	keyOption normalInputCmd = 1 << 1
	keyEdit   normalInputCmd = 1
)

type fullscreenInputCmd struct{}

func (fullscreenInputCmd) isInput() {}

type exitInputCmd struct{}

func (exitInputCmd) isInput() {}

type inputReader interface {
	getInput() (inputCmd, error)
}

type keyboardInputReader struct {
	input uint8
}

func (r *keyboardInputReader) getInput() (inputCmd, error) {
	ev := sdl.PollEvent()
	switch ev := ev.(type) {
	case *sdl.KeyboardEvent:
		if ev.Type == sdl.KEYUP {
			switch ev.Keysym.Sym {
			case sdl.K_RETURN:
				if ev.Keysym.Mod&sdl.KMOD_ALT > 0 {
					return fullscreenInputCmd{}, nil
				}

			case sdl.K_q:
				return exitInputCmd{}, nil
			}
		}

		var key normalInputCmd

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
			key = keyOption
		case sdl.K_SPACE:
			key = keyStart
		case sdl.K_LSHIFT, sdl.K_RSHIFT:
			key = keySelect
		}

		if key == 0 {
			return normalInputCmd(0), nil
		}

		if ev.State == sdl.PRESSED {
			r.input |= uint8(key)
		} else {
			r.input &= 255 ^ uint8(key)
		}

		return normalInputCmd(r.input), nil
	}

	// TODO: impl
	return normalInputCmd(0), nil
}

type gpioInputReader struct {
	pins  gpioInputReaderPins
	input uint8
}

type gpioInputReaderPins struct {
	left  int
	up    int
	right int
	down  int

	sel   int
	start int

	option int
	edit   int
}

type gpioInputReaderPinName string

func (name gpioInputReaderPinName) toM8Key() (normalInputCmd, error) {
	switch name {
	case gpioInputReaderPinLeft:
		return keyLeft, nil
	case gpioInputReaderPinUp:
		return keyUp, nil
	case gpioInputReaderPinRight:
		return keyRight, nil
	case gpioInputReaderPinDown:
		return keyDown, nil
	case gpioInputReaderPinSelect:
		return keySelect, nil
	case gpioInputReaderPinStart:
		return keyStart, nil
	case gpioInputReaderPinOption:
		return keyOption, nil
	case gpioInputReaderPinEdit:
		return keyEdit, nil

	default:
		return 0, errors.Errorf("unknown pin %s", name)
	}
}

const (
	gpioInputReaderPinLeft  gpioInputReaderPinName = "left"
	gpioInputReaderPinUp    gpioInputReaderPinName = "up"
	gpioInputReaderPinRight gpioInputReaderPinName = "right"
	gpioInputReaderPinDown  gpioInputReaderPinName = "down"

	gpioInputReaderPinSelect gpioInputReaderPinName = "select"
	gpioInputReaderPinStart  gpioInputReaderPinName = "start"

	gpioInputReaderPinOption gpioInputReaderPinName = "option"
	gpioInputReaderPinEdit   gpioInputReaderPinName = "edit"
)

func newGPIOInputReaderFromStrConfig(config string) (*gpioInputReader, error) {
	var (
		rdr    gpioInputReader
		pinMap = rdr.pins.pinMap()
	)

	for _, pinCfg := range strings.Split(config, ";") {
		pinCfgParts := strings.Split(pinCfg, "=")
		if len(pinCfgParts) != 2 {
			// Try to get the pin (if any).
			var pin string
			if len(pinCfgParts) > 0 {
				pin = pinCfgParts[0]
			}

			return nil, errors.Errorf("bad pin config for GPIO\nconfig:'%s'\nbad pin: %s", config, pin)
		}

		pinName, pinValueStr := gpioInputReaderPinName(strings.ToLower(pinCfgParts[0])), pinCfgParts[1]
		pinValue, err := strconv.Atoi(pinValueStr)
		if err != nil {
			return nil, errors.Wrap(err, "could not parse pin %s's value")
		}

		pin, ok := pinMap[pinName]
		if !ok {
			return nil, errors.Errorf("unknown pin %s", pinName)
		}

		*pin = pinValue
	}

	// Make sure we configured all the pins.
	for pin, val := range pinMap {
		if *val == 0 {
			return nil, errors.Errorf("missing pin config for pin '%s'", pin)
		}
	}

	return &rdr, nil
}

func (r *gpioInputReader) getInput() (inputCmd, error) {
	for name, pin := range r.pins.pinMap() {
		m8Input, err := name.toM8Key()
		if err != nil {
			return nil, errors.Wrap(err, "error converting pin to m8 input")
		}

		if gpio.ReadPin(gpio.Pin(*pin)) == gpio.High {
			r.input |= uint8(m8Input)
		} else {
			r.input &= 255 ^ uint8(m8Input)
		}
	}

	return normalInputCmd(r.input), nil
}

func (p *gpioInputReaderPins) pinMap() map[gpioInputReaderPinName]*int {
	return map[gpioInputReaderPinName]*int{
		gpioInputReaderPinLeft:  &p.left,
		gpioInputReaderPinUp:    &p.up,
		gpioInputReaderPinRight: &p.right,
		gpioInputReaderPinDown:  &p.down,

		gpioInputReaderPinSelect: &p.sel,
		gpioInputReaderPinStart:  &p.start,

		gpioInputReaderPinOption: &p.option,
		gpioInputReaderPinEdit:   &p.edit,
	}
}
