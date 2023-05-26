package main

import (
	"fmt"
	"log"
	"m8client/input"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"
	gpio "github.com/stianeikeland/go-rpio/v4"
	"go.bug.st/serial"
)

var defaultDeviceName = "/dev/cu.usbmodem136136901"

const (
	// m8ScreenWidth is the actual width of the m8's screen.
	m8ScreenWidth int32 = 320

	// m8ScreenHeight is the actual height of the m8's screen.
	m8ScreenHeight int32 = 240
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %+v\n%s", err, debug.Stack())
		}
	}()

	logger := log.New(os.Stderr, "m8client", log.Flags())

	devName := defaultDeviceName
	if val, ok := os.LookupEnv("M8_DEV"); ok {
		devName = val
	}

	dev, err := serial.Open(devName, &serial.Mode{
		BaudRate: 9000,
		Parity:   serial.NoParity,
	})
	if err != nil {
		panic(errors.Wrap(err, "error opening device"))
	}

	renderer, err := newRenderer(1280, 720)
	if err != nil {
		panic(errors.Wrap(err, "error creating renderer"))
	}

	inputReader, err := newInputReader()
	if err != nil {
		panic(errors.Wrap(err, "error creating input reader"))
	}

	slipReader, err := newSlipReader(logger)
	if err != nil {
		panic(errors.Wrap(err, "error creating slip reader"))
	}

	controller := controller{
		logger,
		renderer,
		slipReader,
		dev,
		0,
		inputReader,
	}
	if err := controller.enableAndResetDisplay(); err != nil {
		panic(err)
	}

	go func() {
		for {
			logger.Println("waiting...")

			cmds, err := controller.nextCmds()
			if err != nil {
				panic(err)
			}

			for _, cmd := range cmds {
				if err := controller.executeCmd(cmd); err != nil {
					panic(err)
				}
			}

			if err := controller.render(); err != nil {
				panic(errors.Wrap(err, "error rendering"))
			}
		}
	}()

	for {
		if err := controller.sendInput(); err != nil {
			panic(err)
		}
	}
}

func newInputReader() (inputReader, error) {
	// Check if we're using GPIO.
	if gpioConfig, ok := os.LookupEnv("M8_USE_GPIO"); ok {
		gpio.Open()
		return input.NewGPIOInputReaderFromStrConfig(gpioConfig)
	}

	// Otherwise, default to keyboard.
	return &input.KeyboardInputReader{}, nil
}

func newSlipReader(logger *log.Logger) (slipRdr, error) {
	if val, ok := os.LookupEnv("M8_SLIP_SAFE"); ok && val == "1" {
		return &safeSlipReader{
			logger: logger,
			reader: &slipReader{},
		}, nil
	}

	return &slipReader{}, nil
}
