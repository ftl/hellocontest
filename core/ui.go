package core

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
	SetValues(KeyerValueProvider)
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
