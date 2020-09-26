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

// MainMenuController provides the functionality for the main menu.
type MainMenuController interface {
	New()
	Open()
	SaveAs()
	ExportCabrillo()
	ExportADIF()
	Quit()
	ShowCallinfo()
}

// KeyerController controls the keyer.
type KeyerController interface {
	SetView(KeyerView)
	SetPatterns([]string)
	GetPattern(index int) string

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
	SetDupChecker(DupChecker)
	
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

type WorkmodeController interface {
	SetView(WorkmodeView)
	SetKeyer(KeyerController)

	SetWorkmode(Workmode)
}

type WorkmodeView interface {
	SetWorkmodeController(WorkmodeController)

	SetWorkmode(Workmode)
}
