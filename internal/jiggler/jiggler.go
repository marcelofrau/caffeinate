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
	user32        = windows.NewLazySystemDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")

	kernel32                   = windows.NewLazySystemDLL("kernel32.dll")
	procSetThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
)

// KEYBDINPUT mirrors the Win32 KEYBDINPUT structure.
type keybdInput struct {
	wVk      uint16
	wScan    uint16
	dwFlags  uint32
	time     uint32
	dwExtraInfo uintptr
}

// INPUT wrapper for keyboard events (same total size as mouse input).
type inputKeyboard struct {
	inputType uint32
	_         [4]byte
	ki        keybdInput
	_         [16]byte
}

const (
	inputMouse      = 0
	mouseeventfMove = 0x0001

	inputTypeKeyboard = 1
	keyeventfKeyUp   = 0x0002

	vkLCtrl = 0xA2
	vkLShift = 0xA0
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
type input struct {
	inputType uint32
	mi        mouseInput
	_         [8]byte // padding to match union size on amd64
}

var jiggleKeys = [...]uint16{vkLCtrl, vkLShift}

// Jiggler manages the mouse movement goroutine.
type Jiggler struct {
	mu       sync.Mutex
	enabled  bool
	zen      bool
	stop     chan struct{}
	keyIndex int
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

// SetZen toggles zen mode (micro-movement, invisible to user).
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
			j.doZen()
		} else {
			j.doNormal()
		}
	}
}

// doZen performs a 1px move and immediately reverses — cursor stays put visually.
func (j *Jiggler) doZen() {
	sendRelative(1, 0)
	time.Sleep(50 * time.Millisecond)
	sendRelative(-1, 0)
	j.sendModifier()
}

// doNormal performs a small visible nudge (5px diagonal) and returns.
func (j *Jiggler) doNormal() {
	sendRelative(5, 5)
	time.Sleep(200 * time.Millisecond)
	sendRelative(-5, -5)
	j.sendModifier()
}

func (j *Jiggler) sendModifier() {
	vk := jiggleKeys[j.keyIndex%len(jiggleKeys)]
	j.keyIndex++
	sendKey(vk)
}

func setExecState(state uint32) {
	procSetThreadExecutionState.Call(uintptr(state))
}

func sendRelative(dx, dy int32) {
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

func sendKey(vk uint16) {
	inp := inputKeyboard{
		inputType: inputTypeKeyboard,
		ki: keybdInput{
			wVk: vk,
		},
	}
	procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		unsafe.Sizeof(inp),
	)
	inp.ki.dwFlags = keyeventfKeyUp
	procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		unsafe.Sizeof(inp),
	)
}
