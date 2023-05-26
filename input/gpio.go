package input

import (
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	gpio "github.com/stianeikeland/go-rpio/v4"
)

type GPIOInputReader struct {
	pollRate time.Duration

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

func (name gpioInputReaderPinName) toM8Key() (CmdKey, error) {
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

func NewGPIOInputReaderFromStrConfig(config string) (*GPIOInputReader, error) {
	var (
		rdr = GPIOInputReader{
			pollRate: 50 * time.Millisecond,
		}
		pinMap = rdr.pins.pinMap()
	)

	for _, pinCfg := range strings.Split(config, ";") {
		cfgParts := strings.Split(pinCfg, "=")
		if len(cfgParts) != 2 {
			// Try to get the key (if any).
			var key string
			if len(cfgParts) > 0 {
				key = cfgParts[0]
			}

			return nil, errors.Errorf("bad config key for GPIO\nconfig:'%s'\nbad key: %s", config, key)
		}

		var (
			key   = strings.ToLower(cfgParts[0])
			value = cfgParts[1]
		)

		switch key {
		case "poll_rate_ms":
			pollRateMs, err := strconv.Atoi(value)
			if err != nil {
				return nil, errors.Wrap(err, "could not parse poll_rate_ms")
			}

			rdr.pollRate = time.Duration(pollRateMs) * time.Millisecond

		default:
			pinName, pinValueStr := gpioInputReaderPinName(key), value
			pinValue, err := strconv.Atoi(pinValueStr)
			if err != nil {
				return nil, errors.Wrapf(err, "could not parse pin %s's value", pinName)
			}

			pin, ok := pinMap[pinName]
			if !ok {
				return nil, errors.Errorf("unknown pin %s", pinName)
			}

			*pin = pinValue
		}
	}

	// Make sure we configured all the pins.
	for pin, val := range pinMap {
		if *val == 0 {
			return nil, errors.Errorf("missing pin config for pin '%s'", pin)
		}
	}

	return &rdr, nil
}

func (r GPIOInputReader) PollRate() time.Duration {
	return r.pollRate
}

func (r *GPIOInputReader) GetInput() (Cmd, error) {
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

	return CmdKey(r.input), nil
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
