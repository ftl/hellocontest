package core

// LogView represents the visual part of the log.
type LogView interface {
	SetLog(Log)

	UpdateAllRows([]QSO)
	RowAdded(QSO)
}

// EntryController controls the entry of QSO data.
type EntryController interface {
	SetView(EntryView)
	SetCallinfo(CallinfoController)

	GotoNextField() EntryField
	GetActiveField() EntryField
	SetActiveField(EntryField)

	BandSelected(string)
	ModeSelected(string)
	EnterCallsign(string)
	QSOSelected(QSO)

	Log()
	Reset()
	CurrentValues() KeyerValues
}

// EntryView represents the visual part of the QSO data entry.
type EntryView interface {
	SetEntryController(EntryController)

	Callsign() string
	SetCallsign(string)
	TheirReport() string
	SetTheirReport(string)
	TheirNumber() string
	SetTheirNumber(string)
	TheirXchange() string
	SetTheirXchange(string)
	Band() string
	SetBand(text string)
	Mode() string
	SetMode(text string)
	MyReport() string
	SetMyReport(string)
	MyNumber() string
	SetMyNumber(string)
	MyXchange() string
	SetMyXchange(string)

	EnableExchangeFields(bool, bool)
	SetActiveField(EntryField)
	SetDuplicateMarker(bool)
	SetEditingMarker(bool)
	ShowMessage(...interface{})
	ClearMessage()
}

// EntryField represents an entry field in the visual part.
type EntryField int

// The entry fields.
const (
	CallsignField EntryField = iota
	TheirReportField
	TheirNumberField
	TheirXchangeField
	MyReportField
	MyNumberField
	MyXchangeField
	OtherField
)

// AppController controls the main functions of the application.
type AppController interface {
	Startup()
	Shutdown()

	SetView(AppView)
	SetLogView(LogView)
	SetEntryView(EntryView)
	SetKeyerView(KeyerView)
	SetCallinfoView(CallinfoView)
}

// AppView represents the visual parts of the main application.
type AppView interface {
	SetMainMenuController(MainMenuController)
	BringToFront()

	ShowFilename(string)
	SelectOpenFile(string, ...string) (string, bool, error)
	SelectSaveFile(string, ...string) (string, bool, error)

	ShowInfoDialog(string, ...interface{})
	ShowErrorDialog(string, ...interface{})
}

// MainMenuController provides the functionality for the main menu.
type MainMenuController interface {
	New()
	Open()
	SaveAs()
	ExportCabrillo()
	ExportADIF()
	Quit()
	Callinfo()
}

// KeyerController controls the keyer.
type KeyerController interface {
	SetView(KeyerView)
	SetPatterns([]string)

	Send(int)
	Stop()
	EnterPattern(int, string)
	EnterSpeed(int)
}

// KeyerView represents the visual parts of the keyer.
type KeyerView interface {
	SetKeyerController(KeyerController)

	ShowMessage(...interface{})

	Pattern(int) string
	SetPattern(int, string)
	Speed() int
	SetSpeed(int)
}

type CallinfoController interface {
	SetView(CallinfoView)

	Show()
	Hide()

	ShowCallsign(string)
}

type CallinfoView interface {
	SetCallinfoController(CallinfoController)
	Show()
	Hide()
	Visible() bool

	SetCallsign(string)
	SetDXCC(string, string, int, int, bool)
}
