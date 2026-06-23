//go:build windows

package jiggler

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                      = windows.NewLazySystemDLL("user32.dll")
	procSendInput               = user32.NewProc("SendInput")
	kernel32                    = windows.NewLazySystemDLL("kernel32.dll")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002

	inputMouse      = 0
	mouseeventfMove = 0x0001
)

// MOUSEINPUT mirrors the Win32 MOUSEINPUT structure.
type mouseInput struct {
	dx          int32
	dy          int32
	mouseData   uint32
	dwFlags     uint32
	time        uint32
	dwExtraInfo uintptr
}

// INPUT mirrors the Win32 INPUT structure (type=mouse).
// Win32 INPUT on amd64: DWORD type (4) + 4-byte pad + 32-byte union = 40 bytes total.
// MOUSEINPUT is 32 bytes, exactly filling the union slot — no trailing pad needed.
type input struct {
	inputType uint32
	_         [4]byte // pad: align union to 8-byte boundary
	mi        mouseInput
}

// Compile-time size assertion: Win32 INPUT must be exactly 40 bytes on amd64.
var _ [40]byte = [unsafe.Sizeof(input{})]byte{}

// Jiggler manages the mouse movement goroutine.
type Jiggler struct {
	mu      sync.Mutex
	enabled bool
	zen     bool
	stop    chan struct{}
}

// New creates a new Jiggler (initially disabled).
func New() *Jiggler {
	return &Jiggler{}
}

// SetEnabled starts or stops jiggling.
func (j *Jiggler) SetEnabled(enabled bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.enabled == enabled {
		return
	}
	j.enabled = enabled

	if enabled {
		setExecState(esContinuous | esSystemRequired | esDisplayRequired)
		j.stop = make(chan struct{})
		go j.loop(j.stop)
	} else {
		setExecState(esContinuous) // release requirements
		close(j.stop)
		j.stop = nil
	}
}

// SetZen toggles zen mode (zero-delta move — cursor stays put, idle timer resets).
func (j *Jiggler) SetZen(zen bool) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.zen = zen
}

// IsEnabled returns current enabled state.
func (j *Jiggler) IsEnabled() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.enabled
}

// IsZen returns current zen state.
func (j *Jiggler) IsZen() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.zen
}

func (j *Jiggler) loop(stop <-chan struct{}) {
	for {
		// Randomise interval between 25–35 seconds to avoid mechanical patterns.
		interval := time.Duration(25+rand.Intn(10)) * time.Second

		select {
		case <-stop:
			return
		case <-time.After(interval):
		}

		j.mu.Lock()
		zen := j.zen
		j.mu.Unlock()

		if zen {
			doZen()
		} else {
			doNormal()
		}
	}
}

// doZen fires a zero-delta mouse move event.
// The cursor stays put visually; Windows still registers the input and resets
// GetLastInputInfo — the same technique used by ArkaneSystems MouseJiggler.
func doZen() {
	sendMouseMove(0, 0)
}

// doNormal performs a small visible nudge (5px diagonal) and returns.
func doNormal() {
	sendMouseMove(5, 5)
	time.Sleep(200 * time.Millisecond)
	sendMouseMove(-5, -5)
}

func setExecState(state uint32) {
	procSetThreadExecutionState.Call(uintptr(state))
}

func sendMouseMove(dx, dy int32) {
	inp := input{
		inputType: inputMouse,
		mi: mouseInput{
			dx:      dx,
			dy:      dy,
			dwFlags: mouseeventfMove,
		},
	}
	procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		unsafe.Sizeof(inp),
	)
}
