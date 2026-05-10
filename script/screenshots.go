package script

import (
	"context"
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui"
)

const ScreenshotsFolder = "./docs/screenshots"

//go:embed screenshots_qsos.csv
var qsoDataCSV string

//go:embed screenshots_config.conf
var screenshotsWindowState string

var ScreenshotsScript = &Script{
	windowState: screenshotsWindowState,
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
			enter: AskForScreenshot("file menu", 1*time.Second),
			steps: []Step{
				OpenMenu("fileMenu"),
				Wait(1 * time.Second),
				HighlightMenuAction("fileMenu", "New..."),
				Wait(1 * time.Second),
				TriggerScreenshot("menu_file_new", WithCapturePopup(), WithDelay(150*time.Millisecond)),
				HighlightMenuAction("fileMenu", "Configuration File..."),
				Wait(1 * time.Second),
				TriggerScreenshot("menu_file_open_configuration", WithCapturePopup(), WithDelay(150*time.Millisecond)),
				CloseMenu("fileMenu"),
				Wait(1 * time.Second),
			},
		},
		{
			enter: AskForScreenshot("new CWT contest session", 1*time.Second),
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
				Describe("close the dialog with 'NEW', save the contest with the proposed filename", 10*time.Second),
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
				Describe("select a current call history file", 20*time.Second),
				TriggerScreenshot("contest_settings_complete", WithCaptureWidget("settingsDialog")),
				TriggerScreenshot("contest_settings_myexchange_cwt", WithCaptureWidget("myExchangeGroup")),
				Describe("close the contest settings dialog, screenshot of empty main window", 10*time.Second),
				TriggerScreenshot("main_window_empty"),
			},
		},
		{
			enter: AskForScreenshot("main window CW macros", 0),
			steps: []Step{
				TriggerScreenshot("main_window_macros", WithCaptureUnion("", "keyerButtons", "esmWorkmode"),
					WithMarkerAtWidget(1, "workmode", -20, 20),
					WithMarkerAtWidget(2, "esm", 110, 20),
					WithMarkerAtWidget(3, "keyerButtons", 300, 20),
					WithMarkerAtWidget(4, "keyerSettingsButton", 20, 10),
					WithMarkerAtWidget(5, "keyerSpeed", -80, 10),
				),
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(func() {
						r.App.Keyer.OpenKeyerSettings()
					})
					return 0
				},
				Describe("the macros dialog complete, select a preset", 10*time.Second),
				TriggerScreenshot("macros_dialog", WithCaptureWidget("keyerSettingsDialog")),
				Describe("close the macros dialog", 10*time.Second),
			},
		},
		{
			enter: AskForScreenshot("main window with QSO data", 0),
			steps: []Step{
				FillQSOList(0, 14),
				TriggerScreenshot("score_table_filled", WithCaptureWidget("scoreTable")),
				TriggerScreenshot("score_graph_filled", WithCaptureWidget("scoreGraph")),
				Wait(7 * time.Second),
				TriggerScreenshot("rate_filled", WithCaptureWidget("rate")),
				FillQSOList(14, -1),
				TriggerScreenshot("main_window_filled"),
			},
		},
		{
			enter: AskForScreenshot("main window QSO data entry", 0),
			steps: []Step{
				func(_ context.Context, r *Runtime) time.Duration {
					r.UI(func() {
						r.App.Entry.Clear()
						r.App.Entry.Enter("AA3B")
						r.App.Entry.RefreshView()
					})
					return 0
				},
				TriggerScreenshot("main_window_entry",
					WithCaptureWidget("entryWidget"),
					WithMarkerAtWidget(1, "predictedBestMatch", 65, 10),
					WithMarkerAtWidget(2, "predictedExchange", 50, 10),
					WithMarkerAtWidget(3, "predictedValue", -30, 10),
					WithMarkerAtWidget(4, "dxccLabel", 20, 10),
				),
				TriggerScreenshot("main_window_supercheck", WithCaptureWidget("supercheckLabel")),
				TriggerScreenshot("main_window_vfo", WithCaptureUnion("", "frequencyLabel", "bandCombo", "modeCombo", "xit")),
				TriggerScreenshot("main_window_status_bar", WithCaptureWidget("statusBar")),
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

type screenshotConfig struct {
	mode        ui.CaptureMode
	widgetName  string
	widgetNames []string
	parentName  string
	rect        ui.Rect
	padding     int
	markers     []ui.Marker
	delay       time.Duration
}

type ScreenshotOption func(*screenshotConfig)

func WithDelay(d time.Duration) ScreenshotOption {
	return func(c *screenshotConfig) { c.delay = d }
}

func WithMainWindow() ScreenshotOption {
	return func(c *screenshotConfig) { c.mode = ui.CaptureMainWindow }
}

func WithActiveWindow() ScreenshotOption {
	return func(c *screenshotConfig) { c.mode = ui.CaptureActiveWindow }
}

func WithCaptureWidget(objectName string) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.mode = ui.CaptureWidget
		c.widgetName = objectName
	}
}

func WithCaptureRect(x, y, w, h int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.mode = ui.CaptureRect
		c.rect = ui.Rect{X: x, Y: y, W: w, H: h}
	}
}

func WithCaptureRectIn(parentObjectName string, x, y, w, h int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.mode = ui.CaptureRect
		c.parentName = parentObjectName
		c.rect = ui.Rect{X: x, Y: y, W: w, H: h}
	}
}

func WithCaptureUnion(parentObjectName string, widgetObjectNames ...string) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.mode = ui.CaptureWidgetUnion
		c.parentName = parentObjectName
		c.widgetNames = append([]string(nil), widgetObjectNames...)
	}
}

func WithCapturePopup() ScreenshotOption {
	return func(c *screenshotConfig) {
		c.mode = ui.CaptureActivePopup
	}
}

func WithPadding(px int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.padding = px
	}
}

func WithMarker(n, x, y int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.markers = append(c.markers, ui.Marker{Number: n, X: x, Y: y})
	}
}

func WithMarkerAtWidget(n int, widgetObjectName string, dx, dy int) ScreenshotOption {
	return func(c *screenshotConfig) {
		c.markers = append(c.markers, ui.Marker{
			Number:     n,
			WidgetName: widgetObjectName,
			DX:         dx,
			DY:         dy,
		})
	}
}

func TriggerScreenshot(name string, opts ...ScreenshotOption) Step {
	cfg := &screenshotConfig{mode: ui.CaptureAuto}
	for _, o := range opts {
		o(cfg)
	}
	return func(ctx context.Context, r *Runtime) time.Duration {
		if cfg.delay > 0 {
			select {
			case <-time.After(cfg.delay):
			case <-ctx.Done():
				return 0
			}
		}
		filename, _ := screenshotFilenames(name)
		if err := backupScreenshot(name); err != nil {
			log.Printf("Cannot backup screenshot %s: %v", filename, err)
		}
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; skipping %s", name)
			_ = restoreScreenshot(name)
			return 0
		}
		pm, err := r.Screenshotter.Capture(ui.ScreenshotRequest{
			Name:        name,
			Mode:        cfg.mode,
			WidgetName:  cfg.widgetName,
			WidgetNames: cfg.widgetNames,
			ParentName:  cfg.parentName,
			Rect:        cfg.rect,
			Padding:     cfg.padding,
		})
		if err != nil {
			log.Printf("Screenshot %s capture failed: %v", name, err)
			_ = restoreScreenshot(name)
			return 0
		}
		if len(cfg.markers) > 0 {
			if err := r.Screenshotter.Annotate(pm, cfg.markers); err != nil {
				log.Printf("Screenshot %s annotate failed: %v", name, err)
			}
		}
		if err := r.Screenshotter.Save(pm, filename); err != nil {
			log.Printf("Screenshot %s save failed: %v", name, err)
			_ = restoreScreenshot(name)
			return 0
		}
		_ = removeBackup(name)
		log.Printf("Screenshot %s successful", name)
		return 0
	}
}

func TriggerScreenshotWithDelay(name string, delay time.Duration) Step {
	return TriggerScreenshot(name, WithDelay(delay))
}

func OpenMenu(menuObjectName string) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot open menu %s", menuObjectName)
			return 0
		}
		if err := r.Screenshotter.ShowMenu(menuObjectName); err != nil {
			log.Printf("OpenMenu %s failed: %v", menuObjectName, err)
		}
		return 0
	}
}

func HighlightMenuAction(menuObjectName, actionTitle string) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			return 0
		}
		if err := r.Screenshotter.HighlightMenuAction(menuObjectName, actionTitle); err != nil {
			log.Printf("HighlightMenuAction %s/%s failed: %v", menuObjectName, actionTitle, err)
		}
		return 0
	}
}

func CloseMenu(menuObjectName string) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			return 0
		}
		if err := r.Screenshotter.HideMenu(menuObjectName); err != nil {
			log.Printf("CloseMenu %s failed: %v", menuObjectName, err)
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

func FillQSOList(begin, end int) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		qsos := parseQSOCSV()
		if begin < 0 {
			begin = 0
		}
		if end < 0 || end > len(qsos) {
			end = len(qsos)
		}
		for i := begin; i < end; i++ {
			qso := qsos[i]
			enterQSOData(r, qso)
		}
		r.UI(r.App.Entry.Clear)
		r.UI(r.App.Entry.RefreshPrediction)
		return 0
	}
}

type qsoData struct {
	minute    int
	frequency core.Frequency
	workmode  core.Workmode
	values    []string
}

func parseQSOCSV() []qsoData {
	lines := strings.Split(qsoDataCSV, "\n")
	result := make([]qsoData, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}

		qso := qsoData{}
		qso.minute, _ = strconv.Atoi(fields[0])
		kHz, _ := strconv.Atoi(fields[1])
		qso.frequency = core.Frequency(kHz * 1000)
		if fields[2] == "r" {
			qso.workmode = core.Run
		} else {
			qso.workmode = core.SearchPounce
		}
		qso.values = fields[3:]

		result = append(result, qso)
	}
	return result
}

func enterQSOData(r *Runtime, qso qsoData) {
	r.UI(r.App.Entry.Clear)
	r.Clock.SetMinute(qso.minute)
	r.App.VFOs[core.VFO1].SetFrequency(qso.frequency)
	time.Sleep(100 * time.Millisecond)
	r.UI(func() {
		r.App.Workmode.SetWorkmode(qso.workmode)
	})
	r.UI(func() {
		for i, value := range qso.values {
			if i > 0 {
				r.App.Entry.GotoNextField()
			}
			r.App.Entry.Enter(value)
		}
	})
	r.UI(r.App.Entry.RefreshView)
	r.UI(r.App.Entry.Log)
}
