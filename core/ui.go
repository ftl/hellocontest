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

	GotoNextField() EntryField
	GetActiveField() EntryField
	SetActiveField(EntryField)

	BandSelected(string)
	ModeSelected(string)
	EnterCallsign(string)

	Log()
	Reset()
	CurrentValues() KeyerValues
}

// EntryView represents the visual part of the QSO data entry.
type EntryView interface {
	SetEntryController(EntryController)

	GetCallsign() string
	SetCallsign(string)
	GetTheirReport() string
	SetTheirReport(string)
	GetTheirNumber() string
	SetTheirNumber(string)
	GetTheirXchange() string
	SetTheirXchange(string)
	GetBand() string
	SetBand(text string)
	GetMode() string
	SetMode(text string)
	GetMyReport() string
	SetMyReport(string)
	GetMyNumber() string
	SetMyNumber(string)
	GetMyXchange() string
	SetMyXchange(string)

	EnableExchangeFields(bool, bool)
	SetActiveField(EntryField)
	SetDuplicateMarker(bool)
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
}

// AppView represents the visual parts of the main application.
type AppView interface {
	SetMainMenuController(MainMenuController)

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
