package ansi

import (
	"fmt"
	"strconv"
	"strings"
)

func MoveCursorTo(row, col int) {
	fmt.Printf(CursorToPos, row, col)
}

func Format(text string, flags ...int) string {
	var formatString strings.Builder
	for index, flag := range flags {
		if index > 0 && index < len(flags) {
			formatString.WriteString(";")
		}
		formatString.WriteString(strconv.Itoa(flag))
	}
	return "\033[" + formatString.String() + "m" + text + "\033[0m"
}
