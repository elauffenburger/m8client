package main

import (
	"log"
	"os"

	"github.com/pkg/errors"
)

var defaultDeviceName = "/dev/cu.usbmodem136136901"

var (
	screenWidth  = 640
	screenHeight = 480
)

func main() {
	dev, err := os.OpenFile(defaultDeviceName, os.O_RDWR, 0666)
	if err != nil {
		panic(errors.Wrap(err, "error opening device"))
	}

	controller := controller{log.New(os.Stderr, "m8client", log.Flags()), dev, 0}
	if err := controller.enableAndResetDisplay(); err != nil {
		panic(err)
	}

	for {
		if err := controller.sendInput(); err != nil {
			panic(err)
		}

		cmds, err := controller.nextCmds()
		if err != nil {
			panic(err)
		}

		for _, cmd := range cmds {
			if err := controller.executeCmd(cmd); err != nil {
				panic(err)
			}
		}
	}
}
