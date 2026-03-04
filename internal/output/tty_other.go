//go:build !(darwin || linux)

package output

import "errors"

func isTerminal(fd uintptr) bool { return false }

func terminalSize(fd uintptr) (rows, cols int, ok bool) {
	return 0, 0, false
}

func enableRawMode(fd int) (func(), error) {
	return nil, errors.New("raw mode not supported on this platform")
}
