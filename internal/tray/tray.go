//go:build windows

package tray

import (
	"log"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/marcelobarbieri/caffeinate/internal/icon"
	"github.com/marcelobarbieri/caffeinate/internal/jiggler"
)

var (
	user32  = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	shell32 = windows.NewLazySystemDLL("shell32.dll")

	procRegisterClassExW       = user32.NewProc("RegisterClassExW")
	procCreateWindowExW        = user32.NewProc("CreateWindowExW")
	procDefWindowProcW         = user32.NewProc("DefWindowProcW")
	procDestroyWindow          = user32.NewProc("DestroyWindow")
	procGetMessageW            = user32.NewProc("GetMessageW")
	procTranslateMessage       = user32.NewProc("TranslateMessage")
	procDispatchMessageW       = user32.NewProc("DispatchMessageW")
	procCreatePopupMenu        = user32.NewProc("CreatePopupMenu")
	procDestroyMenu            = user32.NewProc("DestroyMenu")
	procAppendMenuW            = user32.NewProc("AppendMenuW")
	procTrackPopupMenu         = user32.NewProc("TrackPopupMenu")
	procCheckMenuItem          = user32.NewProc("CheckMenuItem")
	procEnableMenuItem         = user32.NewProc("EnableMenuItem")
	procPostQuitMessage        = user32.NewProc("PostQuitMessage")
	procLoadCursorW            = user32.NewProc("LoadCursorW")
	procGetCursorPos           = user32.NewProc("GetCursorPos")
	procSetForegroundWindow    = user32.NewProc("SetForegroundWindow")
	procDestroyIcon            = user32.NewProc("DestroyIcon")
	procCreateIconFromResource = user32.NewProc("CreateIconFromResource")
	procGetModuleHandleW       = kernel32.NewProc("GetModuleHandleW")
	procShellNotifyIconW       = shell32.NewProc("Shell_NotifyIconW")
	procMessageBoxW            = user32.NewProc("MessageBoxW")
)

const (
	windowClass = "CaffeinateTray"

	WM_APP          = 0x8000
	WM_DESTROY      = 0x0002
	WM_COMMAND      = 0x0111
	WM_LBUTTONDOWN  = 0x0201
	WM_RBUTTONDOWN  = 0x0204

	trayCallback = WM_APP + 1

	NIM_ADD    = 0
	NIM_MODIFY = 1
	NIM_DELETE = 2

	NIF_MESSAGE = 0x0001
	NIF_ICON    = 0x0002
	NIF_TIP     = 0x0004

	MF_STRING    = 0x0000
	MF_SEPARATOR = 0x0800
	MF_CHECKED   = 0x0008
	MF_BYCOMMAND = 0x0000
	MF_GRAYED    = 0x0001

	TPM_LEFTALIGN   = 0x0000
	TPM_RIGHTBUTTON = 0x0002

	IDM_ENABLE = 1
	IDM_GHOST  = 2
	IDM_EXIT   = 3
	IDM_ABOUT  = 4
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     uintptr
	hIcon         uintptr
	hCursor       uintptr
	hbrBackground uintptr
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       uintptr
}

type msg struct {
	hwnd    uintptr
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

type point struct {
	x, y int32
}

type notifyIconData struct {
	cbSize           uint32
	hWnd             uintptr
	uID              uint32
	uFlags           uint32
	uCallbackMessage uint32
	hIcon            uintptr
	szTip            [128]uint16
	dwState          uint32
	dwStateMask      uint32
	szInfo           [256]uint16
	uVersion         uint32
	szInfoTitle      [64]uint16
	dwInfoFlags      uint32
	guidItem         windows.GUID
	hBalloonIcon     uintptr
}

type icoHeader struct {
	reserved uint16
	typ      uint16
	count    uint16
}

type icoEntry struct {
	width    uint8
	height   uint8
	colors   uint8
	reserved uint8
	planes   uint16
	bpp      uint16
	size     uint32
	offset   uint32
}

var (
	trayData   notifyIconData
	appJiggle  *jiggler.Jiggler
	jigglOn    bool
	ghostOn    bool
	wndProcCB  uintptr
)

func Run() {
	runtime.LockOSThread()

	appJiggle = jiggler.New()

	hInst, _, _ := procGetModuleHandleW.Call(0)

	clsName, _ := windows.UTF16PtrFromString(windowClass)
	wndProcCB = syscall.NewCallback(wndProc)

	wc := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   wndProcCB,
		hInstance:     hInst,
		lpszClassName: clsName,
	}
	ret, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	if ret == 0 {
		log.Fatalf("RegisterClassEx failed")
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(clsName)),
		0, 0,
		0, 0, 0, 0,
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		log.Fatalf("CreateWindowEx failed")
	}

	hIcon := loadIcon()

	tip, _ := windows.UTF16FromString("Caffeinate — Keep your machine awake")
	var tipArr [128]uint16
	copy(tipArr[:], tip)

	trayData = notifyIconData{
		cbSize:           uint32(unsafe.Sizeof(notifyIconData{})),
		hWnd:             hwnd,
		uID:              1,
		uFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		uCallbackMessage: trayCallback,
		hIcon:            hIcon,
		szTip:            tipArr,
	}

	ret, _, _ = procShellNotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&trayData)))
	if ret == 0 {
		log.Fatalf("Shell_NotifyIcon(NIM_ADD) failed")
	}

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	appJiggle.SetEnabled(false)
	procShellNotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&trayData)))
	if hIcon != 0 {
		procDestroyIcon.Call(hIcon)
	}
	procDestroyWindow.Call(hwnd)
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0

	case trayCallback:
		switch lParam {
		case WM_LBUTTONDOWN, WM_RBUTTONDOWN:
			showMenu(hwnd)
		}
		return 0

	case WM_COMMAND:
		switch id := uint32(wParam) & 0xFFFF; id {
		case IDM_ENABLE:
			jigglOn = !jigglOn
			appJiggle.SetEnabled(jigglOn)
			if jigglOn {
				setTooltip("Caffeinate — Active")
			} else {
				setTooltip("Caffeinate — Keep your machine awake")
			}

		case IDM_GHOST:
			ghostOn = !ghostOn
			appJiggle.SetZen(ghostOn)

		case IDM_ABOUT:
			title, _ := windows.UTF16PtrFromString("About Caffeinate")
			msg, _ := windows.UTF16PtrFromString(
				"Caffeinate v1.0.2\n\nA lightweight Windows utility that keeps your machine awake.\n\n" +
					"Prevents sleep, screensaver, and 'Away' status in communication apps.\n\n" +
					"https://github.com/marcelobarbieri/caffeinate",
			)
			procMessageBoxW.Call(hwnd, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0)

		case IDM_EXIT:
			procDestroyWindow.Call(hwnd)
		}
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func showMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	item := func(s string) *uint16 { p, _ := windows.UTF16PtrFromString(s); return p }

	procAppendMenuW.Call(hMenu, MF_STRING, IDM_ENABLE, uintptr(unsafe.Pointer(item("Enable Jiggle"))))
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_GHOST, uintptr(unsafe.Pointer(item("Ghost Sip"))))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_ABOUT, uintptr(unsafe.Pointer(item("About Caffeinate..."))))
	procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenuW.Call(hMenu, MF_STRING, IDM_EXIT, uintptr(unsafe.Pointer(item("Exit"))))

	if jigglOn {
		procCheckMenuItem.Call(hMenu, IDM_ENABLE, MF_BYCOMMAND|MF_CHECKED)
	}
	if ghostOn {
		procCheckMenuItem.Call(hMenu, IDM_GHOST, MF_BYCOMMAND|MF_CHECKED)
	}
	if !jigglOn {
		procEnableMenuItem.Call(hMenu, IDM_GHOST, MF_BYCOMMAND|MF_GRAYED)
	}

	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetForegroundWindow.Call(hwnd)
	procTrackPopupMenu.Call(hMenu, TPM_LEFTALIGN|TPM_RIGHTBUTTON, uintptr(pt.x), uintptr(pt.y), 0, hwnd, 0)
}

func setTooltip(s string) {
	tip, _ := windows.UTF16FromString(s)
	copy(trayData.szTip[:], tip)
	procShellNotifyIconW.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&trayData)))
}

func loadIcon() uintptr {
	data := icon.PNG()
	if len(data) < 6 {
		return 0
	}

	hdr := (*icoHeader)(unsafe.Pointer(&data[0]))
	if hdr.reserved != 0 || hdr.typ != 1 || hdr.count == 0 {
		return 0
	}

	var best int
	var bestDist uint32 = 9999
	for i := uint16(0); i < hdr.count; i++ {
		e := (*icoEntry)(unsafe.Pointer(&data[6+uintptr(i)*16]))
		w, h := uint32(e.width), uint32(e.height)
		if e.width == 0 {
			w = 256
		}
		if e.height == 0 {
			h = 256
		}
		d := (w-32)*(w-32) + (h-32)*(h-32)
		if d < bestDist {
			bestDist = d
			best = int(i)
		}
	}

	e := (*icoEntry)(unsafe.Pointer(&data[6+uintptr(best)*16]))
	hIcon, _, _ := procCreateIconFromResource.Call(
		uintptr(unsafe.Pointer(&data[e.offset])),
		uintptr(e.size),
		1,
		0x00030000,
	)
	runtime.KeepAlive(data)
	return hIcon
}
