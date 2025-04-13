package script

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
)

const ScreenshotsFolder = "./docs/screenshots"

var ScreenshotsScript = &Script{
	steps: []Step{
		ClearScreenshotsFolder(),
		Wait(2 * time.Second),
		DescribeScreenshot("about dialog", 1*time.Second),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(app.About)
			return 0
		},
		TriggerScreenshot("about"),
		Wait(2 * time.Second),
		DescribeScreenshot("file menu, highlight QUIT", 10*time.Second),
		TriggerScreenshot("menu_file_quit"),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(func() {
				app.Entry.EntrySelected(core.BandmapEntry{
					Call: callsign.MustParse("DL3NEY"),
				})
			})
			return 0
		},
		DescribeScreenshot("main window with data entry", 0),
		TriggerScreenshot("main_window_data"),
		Describe("all screenshots taken, closing the application", 0),
		func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
			ui(app.Quit)
			return 0
		},
	},
}

func DescribeScreenshot(description string, delay time.Duration) Step {
	return Describe("[SCREENSHOT]\n\n"+description, delay)
}

func ClearScreenshotsFolder() Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		d, err := os.Open(ScreenshotsFolder)
		if err != nil {
			log.Printf("Cannot open screenshots folder: %v", err)
			return 0
		}
		defer d.Close()

		names, err := d.Readdirnames(-1)
		if err != nil {
			log.Printf("Cannot read filenames in %s: %v", ScreenshotsFolder, err)
			return 0
		}
		for _, name := range names {
			filename := filepath.Join(ScreenshotsFolder, name)
			err = os.RemoveAll(filename)
			if err != nil {
				log.Printf("Cannot delete %s: %v", filename, err)
			}
		}
		return 0
	}
}

func TriggerScreenshot(filename string) Step {
	return TriggerScreenshotWithDelay(filename, 0)
}

func TriggerScreenshotWithDelay(filename string, delay time.Duration) Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		// TODO: evaluate ctx.Done() and stop the flameshot process
		cmd := exec.Command("flameshot", "gui")
		cmd.Args = append(cmd.Args, "--path", filepath.Join(ScreenshotsFolder, filename+".png"))
		if delay > 0 {
			cmd.Args = append(cmd.Args, "--delay", fmt.Sprintf("%d", delay.Milliseconds()))
		}

		err := cmd.Run()
		if err != nil {
			log.Printf("Screenshot failed: %v", err)
		} else {
			log.Println("Screenshot successful")
		}
		return 0
	}
}
