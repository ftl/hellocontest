package script

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/ui"
)

const (
	defaultScreenshotsFolder = "../hellocontest-wiki/screenshots"
	screenshotsFolderEnv     = "HELLOCONTEST_SCREENSHOTS"
	versionFilename          = "VERSION"

	// the size of the main window without any view
	basicWindowWidth  = 720
	basicWindowHeight = 400

	// the size of the main window with some views, and the height of each view
	filledWindowWidth    = 960
	filledWindowHeight   = 960
	qsoTableDockHeight   = 180
	rateDockHeight       = 180
	scoreTableDockHeight = 280
	scoreTableWidth      = 520
	scoreGraphDockHeight = 280
)

func writeVersionFile(version string) error {
	if version == "" {
		return nil
	}
	filename := filepath.Join(screenshotsFolder(), versionFilename)
	if err := os.WriteFile(filename, []byte(version+"\n"), 0644); err != nil {
		return fmt.Errorf("cannot write %s: %w", filename, err)
	}
	return nil
}

func screenshotsFolder() string {
	folder := os.Getenv(screenshotsFolderEnv)
	if folder == "" {
		return defaultScreenshotsFolder
	}
	return folder
}

//go:embed screenshots_config.conf
var screenshotsWindowState string

// the screenshots of both contests: one with a predictable exchange, one with QTCs and a
// serial number. Each run appends its own suffix to the file names.
var (
	CWTScreenshotsScript   = newScreenshotsScript(cwtFixture)
	WAEDCScreenshotsScript = newScreenshotsScript(waedcFixture)
)

// newScreenshotsScript takes the screenshots that the wiki pages use. It runs without any
// interaction: the contest comes from a fixture log file that is written before the
// application starts. Every call builds its own sections, which carry the progress of a run.
func newScreenshotsScript(fixture *contestFixture) *Script {
	return &Script{
		windowState:        screenshotsWindowState,
		prepare:            prepareScreenshotsFixture,
		fixture:            fixture,
		timebase:           fixture.now(),
		demoRadio:          true,
		demoRadioSingleVFO: true,
		sections: []*Section{
			{
				steps: []Step{
					SetWorkmode(core.SearchPounce),
					SetVFOFrequency(core.VFO1, 7028000),
					SetVFOMode(core.VFO1, core.ModeCW),
					Wait(3 * time.Second),
				},
			},
			{
				steps: []Step{
					OpenMenu("fileMenu"),
					// the popup is shown asynchronously, an action that is activated before that
					// does not stay active
					Wait(1 * time.Second),
					HighlightMenuAction("fileMenu", "New..."),
					Wait(1 * time.Second),
					TriggerScreenshot("menu_file_new", WithCapturePopup(), WithDelay(150*time.Millisecond)),
					HighlightMenuAction("fileMenu", "Open Configuration File..."),
					Wait(1 * time.Second),
					TriggerScreenshot("menu_file_open_configuration", WithCapturePopup(), WithDelay(150*time.Millisecond)),
					CloseMenu("fileMenu"),
					Wait(1 * time.Second),
				},
			},
			{
				steps: []Step{
					// the dialog is only shown, never accepted: the contest comes from the
					// fixture log, and accepting it would ask for a file name
					OpenNewContestDialog(),
					Wait(2 * time.Second),
					TriggerScreenshot("new", WithActiveWindow()),
					RejectDialog("newContestDialog"),
					Wait(1 * time.Second),
				},
			},
			{
				steps: []Step{
					OpenSettings(),
					Wait(2 * time.Second),
				},
			},
			{
				enter: ForContest(cwtFixture),
				steps: []Step{
					TriggerScreenshot("contest_settings_complete", WithCaptureWidget("settingsDialog")),
					TriggerScreenshot("contest_settings_myexchange", WithCaptureWidget("myExchangeGroup")),
				},
			},
			{
				// only a contest with a report and a serial number can generate them
				enter: ForContest(waedcFixture),
				steps: []Step{
					TriggerScreenshot("contest_settings_myexchange_generate_all", WithCaptureWidget("myExchangeGroup")),
					SetGenerateReport(false),
					Wait(1 * time.Second),
					TriggerScreenshot("contest_settings_myexchange_generate_serial_only", WithCaptureWidget("myExchangeGroup")),
					SetGenerateReport(true),
					Wait(1 * time.Second),
				},
			},
			{
				steps: []Step{
					HideWidget("settingsDialog"),
					Wait(1 * time.Second),
				},
			},
			{
				steps: []Step{
					// the saved window state decides nothing: the script shows the views that the
					// wiki page describes, and nothing else
					HideDocks(),
					ShowWidget("qsoTableDock"),
					ShowWidget("rateDock"),
					ShowWidget("scoreTableDock"),
					ShowWidget("scoreGraphDock"),
					ResizeMainWindow(filledWindowWidth, filledWindowHeight),
					// Qt does not restore the size of a dock that is shown again after it was
					// hidden, therefore the script says how big each view must be
					ResizeDocksVertically(
						DockSize{"qsoTableDock", qsoTableDockHeight},
						DockSize{"rateDock", rateDockHeight},
						DockSize{"scoreTableDock", scoreTableDockHeight},
						DockSize{"scoreGraphDock", scoreGraphDockHeight},
					),
					ResizeDocksHorizontally(
						DockSize{"scoreTableDock", scoreTableWidth},
					),
					Wait(2 * time.Second),
					TriggerScreenshot("main_window_empty", WithMainWindow()),
					FillQSOList(0, -1),
					Wait(2 * time.Second),
					TriggerScreenshot("main_window_filled", WithMainWindow()),
				},
			},
			{
				enter: ForContest(cwtFixture),
				steps: []Step{
					TriggerScreenshot("qsos_table_filled", WithCaptureWidget("qsoTable")),
					TriggerScreenshot("score_graph_filled", WithCaptureWidget("scoreGraph")),
					TriggerScreenshot("score_table_filled", WithCaptureWidget("scoreTable")),
					TriggerScreenshot("qso_rate_filled", WithCaptureWidget("rate")),
				},
			},
			{
				steps: []Step{
					HideDocks(),
					ResizeMainWindow(basicWindowWidth, basicWindowHeight),
					Wait(1 * time.Second),
					TriggerScreenshot("main_window_basic", WithMainWindow()),
				},
			},
			{
				steps: []Step{
					EnterCallsign("AA3B"),
					Wait(1 * time.Second),
					TriggerScreenshot("main_window_entry",
						WithCaptureWidget("vfo1VFOArea"),
						WithMarkerAtWidget(1, "vfo1predictedBestMatch", 65, 10),
						WithMarkerAtWidget(2, "vfo1predictedExchange", 50, 10),
						WithMarkerAtWidget(3, "vfo1predictedValue", -30, 10),
						WithMarkerAtWidget(4, "vfo1dxccLabel", 20, 10),
					),
					TriggerScreenshot("main_window_supercheck", WithCaptureWidget("vfo1supercheckLabel")),
					TriggerScreenshot("main_window_vfo", WithCaptureWidget("vfo1VFOContainer")),
					TriggerScreenshot("main_window_status_bar", WithCaptureWidget("statusBar")),
				},
			},
			{
				steps: []Step{
					TriggerScreenshot("main_window_macros", WithCaptureUnion("", "keyerButtons", "esmWorkmode"),
						WithMarkerAtWidget(1, "workmode", -20, 20),
						WithMarkerAtWidget(2, "esm", 110, 20),
						WithMarkerAtWidget(3, "keyerButtons", 300, 20),
						WithMarkerAtWidget(4, "keyerSettingsButton", 20, 10),
						WithMarkerAtWidget(5, "keyerSpeed", -80, 10),
					),
					OpenKeyerSettings(),
					Wait(2 * time.Second),
					TriggerScreenshot("macros_dialog", WithCaptureWidget("keyerSettingsDialog")),
					HideWidget("keyerSettingsDialog"),
					Wait(1 * time.Second),
				},
			},
			{
				steps: []Step{
					Quit(),
				},
			},
		},
	}
}

func OpenNewContestDialog() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(r.App.New)
		r.UI(func() {
			r.App.NewContestController.SelectContestIdentifier(r.Fixture.identifier)
			r.App.NewContestController.EnterContestName(r.Fixture.name)
			r.App.NewContestController.RefreshView()
		})
		return 0
	}
}

// ForContest lets a section run only in the run of one contest, for screenshots that make
// no sense in the other one
func ForContest(fixture *contestFixture) Condition {
	return func(_ context.Context, r *Runtime) (bool, time.Duration) {
		return r.Fixture == fixture, 0
	}
}

func SetGenerateReport(generate bool) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() {
			r.App.Settings.EnterContestGenerateReport(generate)
			r.App.Settings.RefreshView()
		})
		return 0
	}
}

func OpenSettings() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(r.App.OpenSettings)
		return 0
	}
}

func RejectDialog(objectName string) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot close the dialog %s", objectName)
			return 0
		}
		if err := r.Screenshotter.RejectDialog(objectName); err != nil {
			log.Printf("RejectDialog %s failed: %v", objectName, err)
		}
		return 0
	}
}

func SetVFOFrequency(vfo core.VFOID, frequency core.Frequency) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() { r.App.VFOs[vfo].SetFrequency(frequency) })
		return 0
	}
}

func SetVFOMode(vfo core.VFOID, mode core.Mode) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() { r.App.VFOs[vfo].SetMode(mode) })
		return 0
	}
}

func SetWorkmode(workmode core.Workmode) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() { r.App.Workmode.SetWorkmode(workmode) })
		return 0
	}
}

func EnterCallsign(callsign string) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() {
			r.App.Entry.Clear()
			r.App.Entry.Enter(callsign)
			r.App.Entry.RefreshView()
		})
		return 0
	}
}

func OpenKeyerSettings() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(func() { r.App.Keyer.OpenKeyerSettings() })
		return 0
	}
}

func Quit() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		r.UI(r.App.Quit)
		return 0
	}
}

func ShowWidget(objectName string) Step {
	return setWidgetVisible(objectName, true)
}

func HideWidget(objectName string) Step {
	return setWidgetVisible(objectName, false)
}

func setWidgetVisible(objectName string, visible bool) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot change the visibility of %s", objectName)
			return 0
		}
		if err := r.Screenshotter.SetWidgetVisible(objectName, visible); err != nil {
			log.Printf("SetWidgetVisible %s failed: %v", objectName, err)
		}
		return 0
	}
}

type DockSize struct {
	Name string
	Size int
}

func ResizeDocksVertically(sizes ...DockSize) Step {
	return resizeDocks(sizes, false)
}

func ResizeDocksHorizontally(sizes ...DockSize) Step {
	return resizeDocks(sizes, true)
}

func resizeDocks(sizes []DockSize, horizontal bool) Step {
	names := make([]string, len(sizes))
	values := make([]int, len(sizes))
	for i, size := range sizes {
		names[i] = size.Name
		values[i] = size.Size
	}
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot resize the docks")
			return 0
		}
		if err := r.Screenshotter.ResizeDocks(names, values, horizontal); err != nil {
			log.Printf("ResizeDocks failed: %v", err)
		}
		return 0
	}
}

func RestoreWindowState() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot restore the window state")
			return 0
		}
		if err := r.Screenshotter.RestoreWindowState(); err != nil {
			log.Printf("RestoreWindowState failed: %v", err)
		}
		return 0
	}
}

func ResizeMainWindow(width, height int) Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot resize the main window")
			return 0
		}
		if err := r.Screenshotter.ResizeMainWindow(width, height); err != nil {
			log.Printf("ResizeMainWindow failed: %v", err)
		}
		return 0
	}
}

func HideDocks() Step {
	return func(_ context.Context, r *Runtime) time.Duration {
		if r.Screenshotter == nil {
			log.Printf("No screenshotter configured; cannot hide the docks")
			return 0
		}
		if err := r.Screenshotter.HideDocks(); err != nil {
			log.Printf("HideDocks failed: %v", err)
		}
		return 0
	}
}

// LegacyScreenshotsScript still asks the operator for the screenshots of the wiki pages
// that have no screenshot comments yet.
// the legacy script keeps the file names that the wiki pages use today, therefore its
// fixture carries no name suffix
var legacyFixture = &contestFixture{qsoData: cwtQSOData}

var LegacyScreenshotsScript = &Script{
	windowState:        screenshotsWindowState,
	fixture:            legacyFixture,
	demoRadio:          true,
	demoRadioSingleVFO: true,
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
				Wait(3 * time.Second),
				TriggerScreenshot("menu_file_new", WithCapturePopup(), WithDelay(150*time.Millisecond)),
				HighlightMenuAction("fileMenu", "Open Configuration File..."),
				Wait(3 * time.Second),
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
				TriggerScreenshot("main_window_supercheck", WithCaptureWidget("vfo1supercheckLabel")),
				TriggerScreenshot("main_window_vfo", WithCaptureUnion("", "vfo1Label", "vfo1FrequencyLabel", "vfo1BandCombo", "vfo1ModeCombo", "vfo1RIT", "vfo1TX")),
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
		d, err := os.Open(screenshotsFolder())
		if err != nil {
			log.Printf("Cannot open screenshots folder: %v", err)
			return 0
		}
		defer d.Close()

		names, err := d.Readdirnames(-1)
		if err != nil {
			log.Printf("Cannot read filenames in %s: %v", screenshotsFolder(), err)
			return 0
		}
		for _, name := range names {
			filename := filepath.Join(screenshotsFolder(), name)
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
		filename := filepath.Join(screenshotsFolder(), name)
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
		name := name + r.nameSuffix()
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
	return filepath.Join(screenshotsFolder(), name+".png"), filepath.Join(screenshotsFolder(), name+".bak.png")

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
		qsos := parseQSOCSV(r.qsoData())
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
	offset    time.Duration
	frequency core.Frequency
	workmode  core.Workmode
	values    []string
}

func (r *Runtime) nameSuffix() string {
	if r.Fixture == nil || r.Fixture.nameSuffix == "" {
		return ""
	}
	return "_" + r.Fixture.nameSuffix
}

func (r *Runtime) qsoData() string {
	if r.Fixture == nil {
		return ""
	}
	return r.Fixture.qsoData
}

func parseQSOCSV(csv string) []qsoData {
	lines := strings.Split(csv, "\n")
	result := make([]qsoData, 0, len(lines))
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) < 4 {
			continue
		}

		qso := qsoData{}
		offsetMinutes, _ := strconv.Atoi(fields[0])
		qso.offset = time.Duration(offsetMinutes) * time.Minute
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
	r.Clock.Set(contestStart().Add(qso.offset))
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
