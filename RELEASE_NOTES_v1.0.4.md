# v1.0.4

## Bug fix: `SendInput` was silently failing (Teams/Slack still went Away)

**Root cause**: The Win32 `INPUT` struct on amd64 must be exactly **40 bytes**. Both `input` (mouse) and `inputKeyboard` were padded to **48 bytes**, causing every `SendInput` call to return `0` with `ERROR_INVALID_PARAMETER`. Mouse cursor never moved, keyboard events were never sent.

`SetThreadExecutionState` still worked (prevented Windows sleep), but the idle timer (`GetLastInputInfo`) was never reset — which is what Teams, Slack, and Zoom use to detect Away status.

**Fix**: Corrected struct layout to match the Win32 `INPUT` union exactly:

```
INPUT (amd64) = DWORD type (4) + 4-byte pad + 32-byte union = 40 bytes
MOUSEINPUT    = 32 bytes  →  no trailing pad needed
```

Added compile-time assertion to catch regressions:
```go
var _ [40]byte = [unsafe.Sizeof(input{})]byte{}
```

## Ghost Sip: simplified to zero-delta move

Ghost Sip previously sent `dx=+1` then `dx=-1` with a 50ms sleep between, plus a synthetic modifier keypress (LCtrl/LShift). Replaced with a single `SendInput(dx=0, dy=0, MOUSEEVENTF_MOVE)` — the same technique used by [ArkaneSystems MouseJiggler](https://github.com/arkane-systems/mousejiggler). Windows registers the input event and resets `GetLastInputInfo`; the cursor never moves at all. Keyboard input removed entirely.

## Other changes

- About dialog version string updated to v1.0.4.

