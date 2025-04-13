package script

import (
	"context"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
)

var ScreenshotsScript = &Script{
	steps: []Step{
		Wait(4 * time.Second),
		Describe("about dialog", 1*time.Second),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(app.About)
			return 0
		},
		TriggerScreenshot("about"),
		Wait(2 * time.Second),
		Describe("file menu, highlight QUIT", 10*time.Second),
		TriggerScreenshot("menu_file_quit"),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(func() {
				app.Entry.EntrySelected(core.BandmapEntry{
					Call: callsign.MustParse("DL3NEY"),
				})
			})
			return 0
		},
		Describe("main window with data entry", 0),
		TriggerScreenshot("main_window_data"),
		Describe("all screenshots taken, closing the application", 0),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(app.Quit)
			return 0
		},
	},
}
