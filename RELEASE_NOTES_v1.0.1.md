# v1.0.1

> First public release — a pure Win32 system tray tool that keeps your Windows machine awake.

## What's new

- **Reliable sleep prevention**: Uses `SetThreadExecutionState(ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED)` to prevent the system from sleeping or turning off the display — not just mouse jiggling.
- **Native Win32 tray**: Replaced `getlantern/systray` with a direct Win32 implementation (`Shell_NotifyIconW`, `CreatePopupMenu`, `TrackPopupMenu`). No IPC, no named pipes — the tray icon appears instantly.
- **Much smaller dependency graph**: `go.mod` went from 9 dependencies down to 1 (`golang.org/x/sys`).
- **~2 MB standalone binary** — Go runtime included, no external dependencies, no installer.
- **Ghost Sip mode**: Invisible micro-movement (1px + snap back) for when you want to be discreet.
- **Randomised jiggle interval**: 25–35 seconds to avoid mechanical detection patterns.
- **No elevated privileges required**: Works with standard user rights.

## Fixes

- Sleep/screensaver now actually prevented (was only moving the mouse before, which didn't reset the system power idle timer).
- Startup delay from `getlantern/systray` (1–3 seconds) eliminated — the tray is ready immediately.

## How to use

1. Download `caffeinate.exe` from the Assets section.
2. Run it — the coffee cup icon appears in the system tray.
3. Click the icon and check **☕ Enable Jiggle**.
4. Optionally check **👻 Ghost Sip** for invisible micro-movements.
5. Drop a shortcut in `shell:startup` if you want it to run automatically.
