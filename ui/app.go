package ui

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	qtlib "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/mainthread"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/app"
	"github.com/ftl/hellocontest/core/cfg"
	"github.com/ftl/hellocontest/core/clock"
	"github.com/ftl/hellocontest/core/settings"
)

const (
	settingsOrg        = "com.thecodingflow"
	settingsApp        = "hellocontest"
	settingGeometry    = "ui/mainWindowGeometry"
	settingWindowState = "ui/mainWindowState"

	defaultMainWindowWidth  = 400
	defaultMainWindowHeight = 600
)

// Script matches the ui.Script interface used in main.go
type Script interface {
	app.Script
	Now() time.Time
	SetScreenshotter(s Screenshotter)
	WindowState() string
}

// Run starts the Qt6 application
func Run(version string, sponsors string, startupScript Script, args []string) {
	// enforce the use of XWayland to make docking work properly
	if runtime.GOOS == "linux" && strings.Contains(strings.ToLower(os.Getenv("XDG_SESSION_TYPE")), "wayland") {
		os.Setenv("QT_QPA_PLATFORM", "xcb")
	}

	qApp := qtlib.NewQApplication(os.Args)
	style := NewStyle()
	style.Apply(qApp)

	a := &application{
		version:       version,
		sponsors:      sponsors,
		startupScript: startupScript,
		style:         style,
	}

	a.window = qtlib.NewQMainWindow2()
	a.window.SetWindowTitle("Hello Contest")
	a.window.SetDockOptions(
		qtlib.QMainWindow__AllowNestedDocks |
			qtlib.QMainWindow__AllowTabbedDocks |
			qtlib.QMainWindow__AnimatedDocks |
			qtlib.QMainWindow__GroupedDragging,
	)

	a.stopKeyHandler = newStopKeyHandler(a.window.QWidget)

	configuration, err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	var timebase core.Clock
	if startupScript != nil {
		timebase = startupScript
	} else {
		timebase = clock.New()
	}

	a.controller = app.NewController(version, timebase, a, a.runAsync, configuration, sponsors)
	a.controller.Startup()
	a.controller.SetView(a)
	a.stopKeyHandler.SetStopKeyController(a.controller)

	a.actions = newActions(a.window.QWidget, a.controller, configuration.Keybindings())
	a.controller.Settings.Notify(a.actions)
	a.controller.Entry.Notify(a.actions)
	a.controller.Workmode.Notify(a.actions)
	a.controller.QTCList.Notify(a.actions)
	a.controller.VFOs[core.VFO1].Notify(a.actions)

	a.createMenu()
	a.createCentralWidget()
	a.createStatusBar()
	a.createViews(timebase)
	a.createDialogs()

	qApp.QGuiApplication.OnPaletteChanged(func(_ *qtlib.QPalette) {
		a.centralArea.RepaintForThemeChange()
		a.qsoTableView.RepaintForThemeChange()
		a.qtcTableView.RepaintForThemeChange()
		a.scoreGraphView.RepaintForThemeChange()
		a.scoreTableView.RepaintForThemeChange()
		a.rateView.RepaintForThemeChange()
		a.spotsView.RepaintForThemeChange()
		a.clockView.RepaintForThemeChange()
	})

	if startupScript == nil {
		a.restoreWindowState()
		a.window.OnCloseEvent(func(super func(event *qtlib.QCloseEvent), event *qtlib.QCloseEvent) {
			a.storeWindowState()
			event.Accept()
		})
	}

	a.controller.Settings.Notify(settings.ContestListenerFunc(func(c core.Contest) {
		a.SetExchangeFields(c.MyExchangeFields, c.TheirExchangeFields, c.GenerateSerialExchange)
	}))
	contest := a.controller.Settings.Contest()
	a.SetExchangeFields(contest.MyExchangeFields, contest.TheirExchangeFields, contest.GenerateSerialExchange)

	a.window.Show()
	a.controller.Refresh()

	if startupScript != nil {
		a.restoreWindowStateFromString(startupScript.WindowState())
		startupScript.SetScreenshotter(newScreenshotter(a))
		scriptCtx, cancelScript := context.WithCancel(context.Background())
		qApp.OnAboutToQuit(cancelScript)
		a.controller.RunScript(scriptCtx, startupScript)
	}

	qtlib.QApplication_Exec()
}

type application struct {
	version       string
	sponsors      string
	startupScript Script

	style *Style

	window     *qtlib.QMainWindow
	controller *app.Controller

	stopKeyHandler *stopKeyHandler

	actions *actions

	centralArea *centralArea

	entryView    *entryView
	callinfoView *callinfoView
	workmodeView *workmodeView
	statusView   *statusView
	keyerView    *keyerView
	esmView      *esmView

	mainMenu       *mainMenu
	radioMenu      *radioMenu
	spotSourceMenu *spotSourceMenu

	qsoTableView   *qsoTableView
	qtcTableView   *qtcTableView
	scoreGraphView *scoreGraphView
	scoreTableView *scoreTableView
	spotsView      *spotsView
	rateView       *rateView
	clockView      *clockView

	settingsDialog       *settingsDialog
	keyerSettingsDialog  *keyerSettingsDialog
	newContestDialog     *newContestDialog
	summaryDialog        *summaryDialog
	exportCabrilloDialog *exportCabrilloDialog
	qtcDialog            *qtcDialog
}

func (a *application) findMenu(name string) *qtlib.QMenu {
	if a.mainMenu == nil {
		return nil
	}
	switch name {
	case "fileMenu":
		return a.mainMenu.fileMenu
	case "editMenu":
		return a.mainMenu.editMenu
	}
	return nil
}

func (a *application) createMenu() {
	a.radioMenu = newRadioMenu(a.actions)
	a.spotSourceMenu = newSpotSourceMenu(a.actions)
	a.mainMenu = newMainMenu(a.window, a.actions, a.radioMenu, a.spotSourceMenu)

	// TEMPORAL COUPLING: first create all the widgets, then call the controller
	a.controller.Clusters.SetView(a.spotSourceMenu)
	a.controller.Radio.SetView(a.radioMenu)

	a.window.SetMenuBar(a.mainMenu.menuBar)
}

func (a *application) createCentralWidget() {
	// setup entry view
	a.entryView = newEntryView()
	a.entryView.SetEntryController(a.controller.Entry)
	a.controller.Entry.SetView(a.entryView)
	a.controller.Entry.Notify(a.entryView)

	// setup callinfo view
	a.callinfoView = newCallinfoView()
	a.callinfoView.SetQTCsEnabled(a.controller.QTCList.QTCsEnabled())
	a.controller.Callinfo.SetView(a.callinfoView)
	a.controller.QTCList.Notify(a.callinfoView)

	// setup remaining center parts
	a.esmView = newESMView()
	a.esmView.SetESMController(a.controller.Entry)
	a.controller.Entry.SetESMView(a.esmView)

	a.workmodeView = newWorkmodeView()
	a.workmodeView.SetWorkmodeController(a.controller.Workmode)
	a.controller.Workmode.SetView(a.workmodeView)

	a.keyerView = newKeyerView()
	a.keyerView.SetKeyerController(a.controller.Keyer)
	// TODO: replace with explicit type
	keyerButtons := &keyerButtonAdapter{keyerView: a.keyerView, messenger: a.entryView}
	a.controller.Keyer.SetView(keyerButtons)
	a.controller.Parrot.SetView(a.keyerView)

	a.centralArea = newCentralArea(a.entryView, a.callinfoView, a.esmView, a.workmodeView, a.keyerView)

	a.window.SetCentralWidget(a.centralArea.root)
}

func (a *application) createStatusBar() {
	a.statusView = newStatusView()
	a.controller.ServiceStatus.Notify(a.statusView)
	a.window.SetStatusBar(a.statusView.statusBar)
}

func (a *application) createViews(timebase core.Clock) {
	// setup QSO list dock
	a.qsoTableView = newQSOTableView(a.window.QWidget, a.controller.QSOList)
	a.controller.QSOList.SetView(a.qsoTableView)
	a.controller.QSOList.Notify(a.qsoTableView)

	// setup QTC list dock
	a.qtcTableView = newQTCTableView(a.window.QWidget)
	a.qtcTableView.SetQTCsEnabled(a.controller.QTCList.QTCsEnabled())
	a.controller.QTCList.SetView(a.qtcTableView)
	a.controller.QTCList.Notify(a.qtcTableView)

	// setup rate dock (created before score dock so it appears above on the left)
	a.rateView = newRateView(a.window.QWidget)
	a.controller.Rate.SetView(a.rateView)

	// setup score graph dock (created after rate dock so it appears below on the left)
	a.scoreGraphView = newScoreGraphView(a.window.QWidget, timebase)
	a.controller.Rate.Notify(a.scoreGraphView)

	// setup score table dock
	a.scoreTableView = newScoreTableView(a.window.QWidget)
	// TODO: decouple scoreGraph and scoreTable
	scoreComposite := &scoreCompositeAdapter{
		graph: a.scoreGraphView,
		table: a.scoreTableView,
	}
	a.controller.ScoreController.SetView(scoreComposite)

	// setup clock dock
	a.clockView = newClockView(a.window.QWidget)
	a.controller.ClockView.SetView(a.clockView)

	// setup spots dock
	a.spotsView = newSpotsView(a.window.QWidget, a.controller.Bandmap, a.style)
	a.controller.Bandmap.SetView(a.spotsView)

	// setup the default positions for the dockable view components
	a.window.SetCorner(qtlib.TopLeftCorner, qtlib.LeftDockWidgetArea)
	a.window.SetCorner(qtlib.BottomLeftCorner, qtlib.LeftDockWidgetArea)
	a.window.SetCorner(qtlib.TopRightCorner, qtlib.RightDockWidgetArea)
	a.window.SetCorner(qtlib.BottomRightCorner, qtlib.RightDockWidgetArea)
	a.window.AddDockWidget(qtlib.TopDockWidgetArea, a.qsoTableView.Dock())
	a.window.AddDockWidget(qtlib.TopDockWidgetArea, a.qtcTableView.Dock())
	a.window.AddDockWidget(qtlib.LeftDockWidgetArea, a.rateView.Dock())
	a.window.AddDockWidget(qtlib.LeftDockWidgetArea, a.scoreGraphView.Dock())
	a.window.AddDockWidget(qtlib.LeftDockWidgetArea, a.scoreTableView.Dock())
	a.window.AddDockWidget(qtlib.RightDockWidgetArea, a.clockView.Dock())
	a.window.AddDockWidget(qtlib.RightDockWidgetArea, a.spotsView.Dock())
}

func (a *application) createDialogs() {
	a.settingsDialog = newSettingsDialog(a.window, a.controller.Settings)
	a.controller.Settings.SetView(a.settingsDialog)

	a.keyerSettingsDialog = newKeyerSettingsDialog(a.window, a.controller.Keyer)
	a.controller.Keyer.SetSettingsView(a.keyerSettingsDialog)

	a.newContestDialog = newNewContestDialog(a.window, a.controller.NewContestController)
	a.controller.NewContestController.SetView(a.newContestDialog)

	a.summaryDialog = newSummaryDialog(a.window, a.controller.SummaryController)
	a.controller.SummaryController.SetView(a.summaryDialog)

	a.exportCabrilloDialog = newExportCabrilloDialog(a.window, a.controller.ExportCabrilloController)
	a.controller.ExportCabrilloController.SetView(a.exportCabrilloDialog)

	a.qtcDialog = newQTCDialog(a.window, a.controller.QTCController)
	a.controller.QTCController.SetView(a.qtcDialog)
}

func (a *application) restoreWindowStateFromString(windowState string) {
	if windowState == "" {
		return
	}
	tmp, err := os.CreateTemp("", "hellocontest-windowstate-*.ini")
	if err != nil {
		log.Printf("restoreWindowStateFromString: cannot create temp file: %v", err)
		return
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.WriteString(windowState); err != nil {
		tmp.Close()
		log.Printf("restoreWindowStateFromString: cannot write temp file: %v", err)
		return
	}
	if err := tmp.Close(); err != nil {
		log.Printf("restoreWindowStateFromString: cannot close temp file: %v", err)
		return
	}
	settings := qtlib.NewQSettings4(tmpName, qtlib.QSettings__IniFormat)
	if v := settings.Value(*qtlib.NewQAnyStringView3(settingGeometry), qtlib.NewQVariant()); v.IsValid() {
		a.window.RestoreGeometry(v.ToByteArray())
	}
	if v := settings.Value(*qtlib.NewQAnyStringView3(settingWindowState), qtlib.NewQVariant()); v.IsValid() {
		a.window.RestoreState(v.ToByteArray())
	}
}

func (a *application) restoreWindowState() {
	settings := qtlib.NewQSettings7(settingsOrg, settingsApp)
	if v := settings.Value(*qtlib.NewQAnyStringView3(settingGeometry), qtlib.NewQVariant()); v.IsValid() {
		a.window.RestoreGeometry(v.ToByteArray())
	} else {
		a.window.Resize(defaultMainWindowWidth, defaultMainWindowHeight)
	}
	if v := settings.Value(*qtlib.NewQAnyStringView3(settingWindowState), qtlib.NewQVariant()); v.IsValid() {
		a.window.RestoreState(v.ToByteArray())
	}
}

func (a *application) storeWindowState() {
	settings := qtlib.NewQSettings7(settingsOrg, settingsApp)
	settings.SetValue(
		*qtlib.NewQAnyStringView3(settingGeometry),
		qtlib.NewQVariant12(a.window.SaveGeometry()),
	)
	settings.SetValue(
		*qtlib.NewQAnyStringView3(settingWindowState),
		qtlib.NewQVariant12(a.window.SaveState()),
	)
	settings.Sync()
}

func (a *application) SetExchangeFields(myExchangeFields, theirExchangeFields []core.ExchangeField, generateSerialExchange bool) {
	a.centralArea.SetExchangeFields(myExchangeFields, theirExchangeFields, generateSerialExchange)
}

// runAsync posts a function to run on the Qt main thread (implements core.AsyncRunner)
func (a *application) runAsync(f func()) {
	mainthread.Start(f)
}

// Quit implements app.Quitter
func (a *application) Quit() {
	qtlib.QCoreApplication_Quit()
}

// BringToFront implements app.View
func (a *application) BringToFront() {
	a.window.Raise()
	a.window.ActivateWindow()
}

// ShowFilename implements app.View
func (a *application) ShowFilename(filename string) {
	a.window.SetWindowTitle(fmt.Sprintf("Hello Contest - %s", filename))
}

// SelectOpenFile implements app.View
func (a *application) SelectOpenFile(title string, dir string, patterns ...string) (string, bool, error) {
	filter := buildFileFilter(patterns)
	dlg := qtlib.NewQFileDialog6(nil, title, dir, filter)
	dlg.SetAcceptMode(qtlib.QFileDialog__AcceptOpen)
	dlg.SetFileMode(qtlib.QFileDialog__ExistingFile)
	// Wayland: force a proper top-level window so the compositor lets the user
	// drag the dialog anywhere, including onto other monitors.
	dlg.SetWindowFlags(
		qtlib.Window |
			qtlib.CustomizeWindowHint |
			qtlib.WindowTitleHint |
			qtlib.WindowCloseButtonHint,
	)
	if dlg.Exec() != int(qtlib.QDialog__Accepted) {
		return "", false, nil
	}
	files := dlg.SelectedFiles()
	if len(files) == 0 || files[0] == "" {
		return "", false, nil
	}
	return files[0], true, nil
}

// SelectSaveFile implements app.View
func (a *application) SelectSaveFile(title string, dir string, filename string, patterns ...string) (string, bool, error) {
	filter := buildFileFilter(patterns)
	defaultPath := dir
	if filename != "" {
		defaultPath = dir + "/" + filename
	}
	result := qtlib.QFileDialog_GetSaveFileName4(nil, title, defaultPath, filter)
	if result == "" {
		return "", false, nil
	}
	return result, true, nil
}

// ShowInfoDialog implements app.View
func (a *application) ShowInfoDialog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	qtlib.QMessageBox_Information(a.window.QWidget, "Information", msg)
}

// ShowQuestionDialog implements app.View
func (a *application) ShowQuestionDialog(format string, args ...any) bool {
	msg := fmt.Sprintf(format, args...)
	result := qtlib.QMessageBox_Question(a.window.QWidget, "Question", msg)
	return result == qtlib.QMessageBox__Yes
}

// ShowErrorDialog implements app.View
func (a *application) ShowErrorDialog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	qtlib.QMessageBox_Critical(a.window.QWidget, "Error", msg)
}

type scoreCompositeAdapter struct {
	graph *scoreGraphView
	table *scoreTableView
}

func (s *scoreCompositeAdapter) ShowScore(score core.Score) {
	s.graph.ShowScore(score)
	s.table.ShowScore(score)
}

func (s *scoreCompositeAdapter) ShowGraph() {
	s.graph.Show()
}

func (s *scoreCompositeAdapter) ShowTable() {
	s.table.Show()
}

func (s *scoreCompositeAdapter) SetGoals(points, multis int) {
	s.graph.SetGoals(points, multis)
	s.table.SetGoals(points, multis)
}

func (s *scoreCompositeAdapter) SetQTCsEnabled(enabled bool) {
	s.graph.SetQTCsEnabled(enabled)
	s.table.SetQTCsEnabled(enabled)
}

type keyerButtonAdapter struct {
	*keyerView
	messenger *entryView
}

func (a *keyerButtonAdapter) ShowMessage(args ...any) {
	a.messenger.ShowMessage(core.VFO1, args...)
}

// buildFileFilter converts glob patterns like "*.log" to Qt filter format
func buildFileFilter(patterns []string) string {
	if len(patterns) == 0 {
		return "All Files (*)"
	}
	result := ""
	for i, p := range patterns {
		if i > 0 {
			result += ";;"
		}
		result += fmt.Sprintf("Files (%s)", p)
	}
	result += ";;All Files (*)"
	return result
}
