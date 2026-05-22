//go:build windows

package main

import (
	"github.com/marcelobarbieri/caffeinate/internal/tray"
)

func main() {
	tray.Run()
}
