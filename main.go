package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

var defaultDeviceName = "/dev/cu.usbmodem136136901"

var (
	screenWidth  int32 = 320
	screenHeight int32 = 240
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic: %+v", err)
		}
	}()

	dev, err := os.OpenFile(defaultDeviceName, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0666)
	if err != nil {
		panic(errors.Wrap(err, "error opening device"))
	}

	logger := log.New(os.Stderr, "m8client", log.Flags())

	controller := controller{logger, &renderer{}, &slipReader{}, dev, 0}
	if err := controller.enableAndResetDisplay(); err != nil {
		panic(err)
	}

	for {
		logger.Println("waiting...")

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
