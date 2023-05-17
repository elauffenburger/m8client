package input

type Cmd interface {
	isInput()
}

type CmdKey uint8

func (CmdKey) isInput() {}

const (
	keyLeft   CmdKey = 1 << 7
	keyUp     CmdKey = 1 << 6
	keyDown   CmdKey = 1 << 5
	keySelect CmdKey = 1 << 4
	keyStart  CmdKey = 1 << 3
	keyRight  CmdKey = 1 << 2
	keyOption CmdKey = 1 << 1
	keyEdit   CmdKey = 1
)

type CmdRequestFullScreen struct{}

func (CmdRequestFullScreen) isInput() {}

type CmdRequestExit struct{}

func (CmdRequestExit) isInput() {}
