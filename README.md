# ☕ Caffeinate

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org)
[![Windows](https://img.shields.io/badge/Platform-Windows-0078D4?logo=windows)](https://www.microsoft.com/windows)
[![Standalone Binary](https://img.shields.io/badge/Binary-Standalone-green)]()

**Caffeinate** keeps your Windows machine from going idle — no screensaver, no sleep, no "Away" status on Teams or Slack.  
It lives quietly in the system tray as a coffee cup icon and gets out of your way.

## 🎯 Features

| Feature | Description |
|---|---|
| **Enable Jiggle** | Periodically moves the mouse a few pixels to reset the system idle timer |
| **☯ Ghost Sip (Zen Jiggle)** | Micro-movement mode: cursor moves 1px and snaps back instantly — visually invisible |
| **Random interval** | Jiggle fires every 25–35 seconds (randomised) to avoid mechanical detection patterns |
| **No console window** | Pure tray app, zero UI chrome |
| **Single static binary** | No installer, no runtime dependencies |
| **Lightweight** | Minimal resource usage, designed for always-on operation |

## 🎨 Tray Menu

```
[✓] Enable Jiggle
[ ] ☯  Ghost Sip  (Zen Jiggle)    ← only active when Jiggle is enabled
────────────────────────────────
    Quit
```

Both left-click and right-click on the tray icon open the menu.

## 🔧 Building

### Requirements

- Go 1.21+
- Windows target (cross-compile from Linux/macOS is fine)

### Quick build (PowerShell on Windows)

```powershell
.\build.ps1
```

### Cross-compile from Linux/macOS

```bash
make build
```

This produces `caffeinate.exe` — a standalone binary with no dependencies.

### Suppress console window (recommended)

The `-H windowsgui` linker flag is already set in both `build.ps1` and the `Makefile`.

To also embed the application manifest (proper DPI awareness, no UAC prompt):

```powershell
go install github.com/akavel/rsrc@latest
rsrc -manifest cmd/caffeinate/caffeinate.manifest -o cmd/caffeinate/rsrc.syso
.\build.ps1
```

## 🚀 Run on startup (optional)

1. Press `Win+R`, type `shell:startup`, hit Enter.  
2. Drop a shortcut to `caffeinate.exe` in the folder that opens.

## 📖 How it works

- Uses `SendInput` (Win32 API) with `MOUSEEVENTF_MOVE` and relative deltas.
- **Normal mode**: ±5px diagonal nudge, 200 ms apart.
- **Ghost Sip / Zen mode**: +1px, 50 ms pause, −1px. Cursor visually stays put.
- Interval between jiggle cycles is randomised (25–35 s) to avoid clock-perfect patterns that monitoring software can flag.
- No registry writes, no background services, no elevated privileges required.

## 🎁 Assets

The application icon was designed by [Icons8](https://icons8.com) and is licensed under their standard terms.

## 📄 License

MIT — See [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Feel free to open issues or submit pull requests.
