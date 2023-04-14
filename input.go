package main

type input uint8

const (
	keyLeft   input = 1 << 7
	keyUp     input = 1 << 6
	keyDown   input = 1 << 5
	keySelect input = 1 << 4
	keyStart  input = 1 << 3
	keyRight  input = 1 << 2
	keyOpt    input = 1 << 1
	keyEdit   input = 1
)
