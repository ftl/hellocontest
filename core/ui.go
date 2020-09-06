package core

// LogView represents the visual part of the log.
type LogbookView interface {
	SetLogbook(Logbook)

	UpdateAllRows([]QSO)
	RowAdded(QSO)
}

// EntryController controls the entry of QSO data.
type EntryController interface {
	DupChecker

	SetView(EntryView)
	SetKeyer(KeyerController)
	SetCallinfo(CallinfoController)

	GotoNextField() EntryField
	GetActiveField() EntryField
	SetActiveField(EntryField)

	BandSelected(string)
	ModeSelected(string)
	EnterCallsign(string)
	SendQuestion()
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

// AppController controls the main functions of the application.
type AppController interface {
	Startup()
	Shutdown()

	SetView(AppView)
	SetLogbookView(LogbookView)
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
	SendQuestion(q string)
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
	SetDuplicateMarker(bool)
	SetDXCC(string, string, int, int, bool)
	SetSupercheck(callsigns []AnnotatedCallsign)
}
