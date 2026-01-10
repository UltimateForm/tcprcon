package ansi

import "fmt"

func MoveCursorTo(row, col int) {
	fmt.Printf(CursorToPos, row, col)
}
