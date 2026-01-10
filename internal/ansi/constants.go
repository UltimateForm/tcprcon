package ansi

const (
	ClearScreen = "\033[2J"
	CursorHome  = "\033[H"
	CursorToPos = "\033[%d;%dH" // use with fmt.Sprintf, the two ds are for the row and column coordinates
)
