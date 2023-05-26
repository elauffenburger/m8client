package input

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type KeyboardInputReader struct {
	input uint8
}

func (r *KeyboardInputReader) PollRate() time.Duration {
	return 0
}

func (r *KeyboardInputReader) GetInput() (Cmd, error) {
	ev := sdl.PollEvent()
	switch ev := ev.(type) {
	case *sdl.KeyboardEvent:
		if ev.Type == sdl.KEYUP {
			switch ev.Keysym.Sym {
			case sdl.K_RETURN:
				if ev.Keysym.Mod&sdl.KMOD_ALT > 0 {
					return CmdRequestFullScreen{}, nil
				}

			case sdl.K_q:
				return CmdRequestExit{}, nil
			}
		}

		var key CmdKey

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
			return CmdKey(0), nil
		}

		if ev.State == sdl.PRESSED {
			r.input |= uint8(key)
		} else {
			r.input &= 255 ^ uint8(key)
		}

		return CmdKey(r.input), nil
	}

	// TODO: impl
	return CmdKey(0), nil
}
