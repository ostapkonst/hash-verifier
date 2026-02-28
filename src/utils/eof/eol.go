package eof

import (
	"runtime"
)

var PlatformEOF = getEOF()

func getEOF() string {
	switch runtime.GOOS {
	case "windows":
		return "\r\n"
	case "darwin":
		return "\n"
	default:
		return "\n"
	}
}
