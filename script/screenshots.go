package script

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ftl/hellocontest/core"
)

const ScreenshotsFolder = "./docs/screenshots"

var ScreenshotsScript = &Script{
	sections: []*Section{
		{
			steps: []Step{
				SetTimebase("2023-06-28T19:00:00Z"),
				Wait(2 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("about dialog", 1*time.Second),
			steps: []Step{
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(r.App.About)
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
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(r.App.New)
					r.UI(func() {
						r.App.NewContestController.SelectContestIdentifier("CW-OPS")
						r.App.NewContestController.EnterContestName("CWT Screenshot Demo")
						r.App.NewContestController.RefreshView()
					})
					return 0
				},
				TriggerScreenshot("new_cwt"),
				Describe("close the dialog with 'NEW', save the contest with the proposed filename\nthe settings dialog will show up, just wait for the next set of instructions", 10*time.Second),
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(func() {
						r.App.Settings.EnterStationCallsign("DL0ABC")
						r.App.Settings.EnterStationOperator("DL1ABC")
						r.App.Settings.EnterStationLocator("AA00xx")
						r.App.Settings.SetContestStartTimeNow()
						r.App.Settings.EnterContestExchangeValue(core.EntryField("myExchange_1"), "Walter")
						r.App.Settings.EnterContestExchangeValue(core.EntryField("myExchange_2"), "DL")
						r.App.Settings.RefreshView()
					})
					return 0
				},
				Describe("set the current hour as start time, select a current call history file", 20*time.Second),
				Describe("contest settings dialog, complete", 1*time.Second),
				TriggerScreenshot("contest_settings_complete"),
				Describe("contest settings dialog, section 'My Exchange' with name Flo and dxcc_prefix DL", 10*time.Second),
				TriggerScreenshot("contest_settings_myexchange_cwt"),
				Describe("close the contest settings dialog, screenshot of empty main window", 10*time.Second),
				TriggerScreenshot("main_window_empty"),
			},
		},
		{
			enter: AskForScreenshot("main window QSO data entry", 0),
			steps: []Step{
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(func() {
						r.App.Entry.Clear()
						r.App.Entry.Enter("DL3NEY")
						r.App.Entry.RefreshView()
					})
					return 0
				},
				Describe("only the entry area, mark (1) best matching callsign, (2) predicted exchange, (3) qso value, (4) callsign infos", 1*time.Second),
				TriggerScreenshot("main_window_entry"),
				Describe("only the supercheck area", 1*time.Second),
				TriggerScreenshot("main_window_supercheck"),
				Describe("only the vfo area", 1*time.Second),
				TriggerScreenshot("main_window_vfo"),
				Describe("only the status bar", 1*time.Second),
				TriggerScreenshot("main_window_status_bar"),
			},
		},
		{
			steps: []Step{
				Describe("all screenshots taken, closing the application", 0),
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(r.App.Quit)
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
	return func(_ context.Context, _ *Runtime) time.Duration {
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
	return func(_ context.Context, _ *Runtime) time.Duration {
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
	return func(_ context.Context, _ *Runtime) time.Duration {
		filename, _ := screenshotFilenames(name)

		err := backupScreenshot(name)
		if err != nil {
			log.Printf("Cannot backup screenshot %s: %v", filename, err)
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

		if !fileExists(filename) {
			log.Printf("restoring screenhot %s", name)
			err = restoreScreenshot(name)
		} else {
			log.Printf("removing screenshot backup %s", name)
			err = removeBackup(name)
		}
		if err != nil {
			log.Printf("Screenshot %s backup handling failed: %v", name, err)
		}

		return 0
	}
}

func screenshotFilenames(name string) (string, string) {
	return filepath.Join(ScreenshotsFolder, name+".png"), filepath.Join(ScreenshotsFolder, name+".bak.png")

}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func backupScreenshot(name string) error {
	filename, backupFilename := screenshotFilenames(name)
	if !fileExists(filename) {
		return nil
	}

	if fileExists(backupFilename) {
		err := os.Remove(backupFilename)
		if err != nil {
			return err
		}
	}

	return os.Rename(filename, backupFilename)
}

func restoreScreenshot(name string) error {
	filename, backupFilename := screenshotFilenames(name)
	if fileExists(filename) {
		err := os.Remove(filename)
		if err != nil {
			return err
		}
	}

	if fileExists(backupFilename) {
		return os.Rename(backupFilename, filename)
	}

	return nil
}

func removeBackup(name string) error {
	_, backupFilename := screenshotFilenames(name)
	if fileExists(backupFilename) {
		return os.Remove(backupFilename)
	}
	return nil
}
