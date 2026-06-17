# v1.0.3

## What's new

- **Fix Teams "Away" detection**: Mouse jiggling alone (`MOUSEEVENTF_MOVE` via `SendInput`) does not reset the Windows idle timer (`GetLastInputInfo`), so communication apps like Teams, Slack, and Zoom still marked the user as away. Now each jiggle cycle also sends a synthetic modifier key press (alternating between Ctrl and Shift) which properly resets the idle timer with zero side effects.
- **Alternating modifier keys**: Each jiggle cycle cycles between Left Ctrl and Left Shift to avoid mechanical patterns.
- **No side effects**: Modifier keys by themselves do not type characters, open menus, or interfere with active applications.

## Background

Windows has two separate timers:

| Timer | Controlled by | Reset by |
|-------|--------------|----------|
| Power timer | `SetThreadExecutionState(ES_SYSTEM_REQUIRED \| ES_DISPLAY_REQUIRED)` | Already handled by caffeinate (prevents sleep/display off) |
| Idle timer | `GetLastInputInfo()` | Physical input or `SendInput` with `INPUT_KEYBOARD` / mouse button events — **NOT** `MOUSEEVENTF_MOVE` |

The mouse jiggling prevented the system from sleeping but Teams still saw the idle timer expiring. Adding a keyboard modifier press (`SendInput` with `INPUT_KEYBOARD`) ensures the idle timer is properly reset.
