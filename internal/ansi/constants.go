package ansi

const (
	ClearScreen    = "\033[2J"
	CursorHome     = "\033[H"
	CursorToPos    = "\033[%d;%dH" // use with fmt.Sprintf, the two ds are for the row and column coordinates
	EnterAltScreen = "\033[?1049h"
	ExitAltScreen  = "\033[?1049l"
	Red            = 31
	Green          = 32
	Yellow         = 33
	Blue           = 34
	Magenta        = 35
	Cyan           = 36
	Bold           = 1
	Reset          = 0
)
