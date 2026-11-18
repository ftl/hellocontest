package model

type Keyer struct {
	Name    string    `json:"name"`
	Type    KeyerType `json:"type"`
	Address string    `json:"address"`
}

// RadioKeyer is the name used to reference the radio's internal keyer.
const RadioKeyer = "radio"

// CQMacroIndex is the index of the keyer macro that is used to call CQ when running
const CQMacroIndex = 0

type KeyerType string

const (
	KeyerTypeCWDaemon KeyerType = "cwdaemon"
)

type KeyerSettings struct {
	WPM                   int      `json:"wpm"`
	Preset                string   `json:"preset"`
	SPMacros              []string `json:"sp_macros"`
	RunMacros             []string `json:"run_macros"`
	SPLabels              []string `json:"sp_labels"`
	RunLabels             []string `json:"run_labels"`
	ParrotIntervalSeconds int      `json:"parrot_interval_seconds"`
}

type KeyerPreset struct {
	Name      string   `json:"name"`
	SPMacros  []string `json:"sp_macros"`
	RunMacros []string `json:"run_macros"`
	SPLabels  []string `json:"sp_labels"`
	RunLabels []string `json:"run_labels"`
}

// KeyerValues contains the values that can be used as variables in the keyer templates.
type KeyerValues struct {
	TheirCall   string
	MyNumber    QSONumber
	MyReport    RST
	MyXchange   string
	MyExchange  string
	MyExchanges []string
	LastNumber  QSONumber
}
