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
	sections: []*Section{
		{
			steps: []Step{
				Wait(2 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("about dialog", 1*time.Second),
			steps: []Step{
				func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
					ui(app.About)
					return 0
				},
				TriggerScreenshot("about"),
				Wait(5 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("file menu, hightlight OPEN CONFIGURATION FILE", 10*time.Second),
			steps: []Step{
				TriggerScreenshot("menu_file_open_configuration"),
				Wait(5 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("file menu, hightlight NEW", 10*time.Second),
			steps: []Step{
				TriggerScreenshot("menu_file_new"),
				Wait(5 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("new CWT, enter name CWT 2025 Test", 1*time.Second),
			steps: []Step{
				func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
					ui(app.New)
					ui(func() {
						app.NewContestController.SelectContestIdentifier("CW-OPS")
						app.NewContestController.EnterContestName("CWT 2025 Test")
						app.NewContestController.RefreshView()
					})
					return 0
				},
				TriggerScreenshot("new_cwt"),
				Describe("close the dialog with 'NEW', save the contest with the proposed filename", 10*time.Second),
				Describe("set the current hour as start time, select a current call history file", 20*time.Second),
				Describe("contest settings dialog, complete", 1*time.Second),
				TriggerScreenshot("contest_settings_complete"),
				Describe("contest settings dialog, section 'My Exchange' with name Flo and dxcc_prefix DL", 10*time.Second),
				TriggerScreenshot("contest_settings_myexchange_cwt"),
				Describe("close the contest settings dialog, screenshot of empty main window", 5*time.Second),
				TriggerScreenshot("main_window_empty"),
			},
		},
		{
			enter: AskForScreenshot("main window with QSO data", 0),
			steps: []Step{
				func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
					ui(func() {
						app.Entry.EntrySelected(core.BandmapEntry{
							Call: callsign.MustParse("DL3NEY"),
						})
						// TODO: app.Entry.RefreshView()
					})
					return 0
				},
				Describe("only the entry area", 1*time.Second),
				TriggerScreenshot("main_window_entry"),
				Describe("only the vfo area", 1*time.Second),
				TriggerScreenshot("main_window_vfo"),
				Describe("only the status bar", 1*time.Second),
				TriggerScreenshot("main_window_status_bar"),
			},
		},
		{
			steps: []Step{
				Describe("all screenshots taken, closing the application", 0),
				func(_ context.Context, app *app.Controller, ui func(func())) time.Duration {
					ui(app.Quit)
					return 0
				},
			},
		},
	},
}

func AskForScreenshot(description string, delay time.Duration) Condition {
	return Ask("[SCREENSHOT]\n\n"+description, delay)
}

func DescribeScreenshot(description string, delay time.Duration) Step {
	return Describe("[SCREENSHOT]\n\n"+description, delay)
}

func ClearScreenshotsFolder() Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		log.Printf("[clearing screenshots folder]")
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

func DeleteScreenshot(name string) Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		filename := filepath.Join(ScreenshotsFolder, name)
		err := os.RemoveAll(filename)
		if err != nil {
			log.Printf("Cannot delete %s: %v", filename, err)
		}
		return 0
	}
}

func TriggerScreenshot(filename string) Step {
	return TriggerScreenshotWithDelay(filename, 0)
}

func TriggerScreenshotWithDelay(name string, delay time.Duration) Step {
	return func(_ context.Context, _ *app.Controller, _ func(func())) time.Duration {
		filename := filepath.Join(ScreenshotsFolder, name+".png")
		backupFilename := filepath.Join(ScreenshotsFolder, name+".bak.png")
		_ = backupFilename
		// TODO: do not delete the filename, but rename it to backupFilename; if the later already exists, delete it first
		err := os.RemoveAll(filename)
		if err != nil {
			log.Printf("Cannot delete %s: %v", filename, err)
		}

		// TODO: evaluate ctx.Done() and stop the flameshot process
		cmd := exec.Command("flameshot", "gui")
		cmd.Args = append(cmd.Args, "--path", filename)
		if delay > 0 {
			cmd.Args = append(cmd.Args, "--delay", fmt.Sprintf("%d", delay.Milliseconds()))
		}

		err = cmd.Run()
		if err != nil {
			log.Printf("Screenshot %s failed: %v", name, err)
		} else {
			log.Printf("Screenshot %s successful", name)
		}

		// TODO: if the filename does not exist, flameshot was canceled; rename backupFilename to filename in order to restore the screenshot

		return 0
	}
}
