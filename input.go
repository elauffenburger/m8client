package main

import "m8client/input"

type inputReader interface {
	GetInput() (input.Cmd, error)
}
