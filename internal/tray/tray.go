//go:build windows

package tray

import (
	"github.com/getlantern/systray"
	"github.com/marcelobarbieri/caffeinate/internal/icon"
	"github.com/marcelobarbieri/caffeinate/internal/jiggler"
)

const (
	appName    = "Caffeinate"
	appTooltip = "Caffeinate — Keep your machine awake"
)

// Run starts the systray event loop. Blocks until the tray icon is removed.
func Run() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.PNG())
	systray.SetTitle(appName)
	systray.SetTooltip(appTooltip)

	j := jiggler.New()

	// ── Menu items ──────────────────────────────────────────────────
	mEnable := systray.AddMenuItemCheckbox("Enable Jiggle", "Start keeping the system awake", false)
	mZen := systray.AddMenuItemCheckbox("☯  Ghost Sip  (Zen Jiggle)", "Invisible micro-movement — cursor stays put", false)
	mZen.Disable() // only usable when jiggle is enabled

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Exit Caffeinate")

	// ── Event loop ──────────────────────────────────────────────────
	go func() {
		for {
			select {

			case <-mEnable.ClickedCh:
				if mEnable.Checked() {
					mEnable.Uncheck()
					j.SetEnabled(false)
					mZen.Disable()
					systray.SetTooltip(appTooltip)
				} else {
					mEnable.Check()
					j.SetEnabled(true)
					mZen.Enable()
					systray.SetTooltip(appName + " — Active ☕")
				}

			case <-mZen.ClickedCh:
				if mZen.Checked() {
					mZen.Uncheck()
					j.SetZen(false)
				} else {
					mZen.Check()
					j.SetZen(true)
				}

			case <-mQuit.ClickedCh:
				j.SetEnabled(false)
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {}
