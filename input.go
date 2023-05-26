package main

import (
	"m8client/input"
	"time"
)

type inputReader interface {
	GetInput() (input.Cmd, error)
	PollRate() time.Duration
}
