package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type color string

const (
	colorNone    color = "\x1b[0m"
	colorBrRed   color = "\x1b[91m"
	colorBrGreen color = "\x1b[92m"
	colorBrBlue  color = "\x1b[94m"
	colorBrCyan  color = "\x1b[96m"
)

func status(color color, s, msg string) {
	fmt.Printf("%s%15s%s  %s\n", color, s, colorNone, msg)
}

func statusf(color color, s, msg string, v ...interface{}) {
	status(color, s, fmt.Sprintf(msg, v...))
}

func errorf(f string, v ...interface{}) {
	statusf(colorBrRed, "error", f, v...)
}

func fatalf(f string, v ...interface{}) {
	errorf(f, v...)
	os.Exit(1)
}

func readChar(r io.Reader) (rune, error) {
	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	if err != nil {
		return 0, err
	}

	return char, err
}
