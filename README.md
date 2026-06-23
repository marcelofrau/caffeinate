# ☕ Caffeinate

[![License: GPLv3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Windows](https://img.shields.io/badge/Platform-Windows-0078D4?logo=windows)](https://www.microsoft.com/windows)
[![Release](https://img.shields.io/badge/Release-v1.0.4-brightgreen)](https://github.com/marcelobarbieri/caffeinate/releases)
[![Standalone Binary](https://img.shields.io/badge/Binary-Standalone-green)]()

> ⚠️ **Windows only.** This app uses native Win32 APIs (`user32.dll`, `kernel32.dll`, `shell32.dll`) and is not cross-platform.

**Caffeinate** keeps your Windows machine from going idle — no screensaver, no sleep, no "Away" status on Teams or Slack.  
It lives quietly in the system tray as a coffee cup icon and gets out of your way.

## 🎯 Features

| Feature | Description |
|---|---|
| **Enable Jiggle** | Periodically moves the mouse and sends a synthetic modifier keypress to reset the system idle timer |
| **Ghost Sip** | Micro-movement mode: cursor moves 1px and snaps back instantly — visually invisible |
| **Random interval** | Jiggle fires every 25–35 seconds (randomised) to avoid mechanical detection patterns |
| **No console window** | Pure Win32 tray app, zero UI chrome |
| **Single static binary** | No installer, no runtime dependencies |
| **Lightweight** | Minimal resource usage, designed for always-on operation |
| **Windows only** | Uses native Win32 APIs — no cross-platform support planned |

## 🎨 Tray Menu

```
[✓] ☕ Enable Jiggle
[ ] 👻 Ghost Sip                    ← only active when Jiggle is enabled
────────────────────────────────
    About Caffeinate...
────────────────────────────────
    Exit
```

Both left-click and right-click on the tray icon open the menu.

## 🔧 Building

### Requirements

- Go 1.21+
- Windows (cross-compile to Windows from other OSes is fine)
- **rsrc** (for no-console window + exe icon) — `go install github.com/akavel/rsrc@latest`

### Build

```powershell
go install github.com/akavel/rsrc@latest
.\build.ps1
```

If `rsrc` is not installed the build still works, but a console window may briefly flash on startup.

This produces `dist/caffeinate.exe` — a standalone binary with no dependencies.

## 🚀 Run on startup (optional)

1. Press `Win+R`, type `shell:startup`, hit Enter.  
2. Drop a shortcut to `caffeinate.exe` in the folder that opens.

## 📖 How it works

Windows has two separate timers that communication apps (Teams, Slack, Zoom) monitor:

| Timer | Controlled by | Reset by |
|-------|--------------|----------|
| **Power timer** | `SetThreadExecutionState(ES_SYSTEM_REQUIRED \| ES_DISPLAY_REQUIRED)` | Prevents sleep and display-off |
| **Idle timer** | `GetLastInputInfo()` | Physical input OR `SendInput` with `INPUT_KEYBOARD` / mouse button events — **NOT** `MOUSEEVENTF_MOVE` alone |

Caffeinate handles **both**:

- `SetThreadExecutionState(ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED | ES_CONTINUOUS)` keeps the power timer satisfied — no sleep, no display off.
- Each jiggle cycle also sends a synthetic modifier key press (`SendInput` with `INPUT_KEYBOARD`, alternating Left Ctrl / Left Shift), which properly resets the idle timer so communication apps never see the user as Away.
- **Normal mode**: ±5px diagonal nudge, 200 ms apart.
- **Ghost Sip mode**: zero-delta mouse move (`dx=0, dy=0`). Windows registers the input event and resets the idle timer; cursor never moves at all. Same technique used by ArkaneSystems MouseJiggler.
- Interval between jiggle cycles is randomised (25–35 s) to avoid clock-perfect patterns that monitoring software can flag.
- Pure Win32 tray implementation — no external libraries, no IPC, fast startup.
- No registry writes, no background services, no elevated privileges required.

## 🎁 Assets

The application icon was designed by [Icons8](https://icons8.com) and is licensed under their standard terms.

## 📄 License

GPL v3 — See [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.
