//go:build darwin || linux

package output

import (
	"syscall"
	"unsafe"
)

func isTerminal(fd uintptr) bool {
	_, _, ok := terminalSize(fd)
	return ok
}

func terminalSize(fd uintptr) (rows, cols int, ok bool) {
	var wsz [4]uint16
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&wsz[0])))
	if err != 0 {
		return 0, 0, false
	}
	rows = int(wsz[0])
	cols = int(wsz[1])
	if rows <= 0 || cols <= 0 {
		return 0, 0, false
	}
	return rows, cols, true
}

func enableRawMode(fd int) (restore func(), err error) {
	var orig syscall.Termios
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&orig)), 0, 0, 0); e != 0 {
		return nil, e
	}
	raw := orig
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if _, _, e := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); e != 0 {
		return nil, e
	}
	return func() {
		syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(&orig)), 0, 0, 0)
	}, nil
}
