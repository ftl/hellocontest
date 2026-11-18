package model

import "strconv"

type Radio struct {
	Name    string            `json:"name"`
	Type    RadioType         `json:"type"`
	Address string            `json:"address"`
	Keyer   string            `json:"keyer"`
	Options map[string]string `json:"options"`
}

type RadioType string

const (
	RadioTypeHamlib RadioType = "hamlib"
	RadioTypeTCI    RadioType = "tci"
	RadioTypeDemo   RadioType = "demo"
)

const (
	DemoRadioName       = "Demo"
	DemoSingleVFOOption = "single_vfo"
)

// DemoRadio is a radio that only exists inside the application. A script selects it to put
// the user interface into a defined state, without a transceiver and without a keyer.
func DemoRadio(singleVFO bool) Radio {
	return Radio{
		Name:    DemoRadioName,
		Type:    RadioTypeDemo,
		Keyer:   RadioKeyer,
		Options: map[string]string{DemoSingleVFOOption: strconv.FormatBool(singleVFO)},
	}
}

type VFOID int

const (
	VFO1 VFOID = iota
	VFO2

	VFOCount
)

type VFO interface {
	Name() string
	Notify(any)
	Refresh()
	SetFrequency(Frequency)
	ShiftFrequency(Frequency)
	SetBand(Band)
	SetMode(Mode)
	IncrementalTuningActive(kind IncrementalTuningKind) bool
	SetIncrementalTuningActive(kind IncrementalTuningKind, active bool)
	ShiftIncrementalTuning(IncrementalTuningKind, Frequency)
	ToggleAvailableIncrementalTuning()
	ShiftAvailableIncrementalTuning(Frequency)
}

type CurrentVFOListener interface {
	CurrentVFOChanged(VFOID)
}

type CurrentVFOListenerFunc func(VFOID)

func (f CurrentVFOListenerFunc) CurrentVFOChanged(vfo VFOID) {
	f(vfo)
}

type TXVFOListener interface {
	TXVFOChanged(VFOID)
}

type TXVFOListenerFunc func(VFOID)

func (f TXVFOListenerFunc) TXVFOChanged(vfo VFOID) {
	f(vfo)
}

type FocusedVFOListener interface {
	FocusedVFOChanged(VFOID)
}

type FocusedVFOListenerFunc func(VFOID)

func (f FocusedVFOListenerFunc) FocusedVFOChanged(vfo VFOID) {
	f(vfo)
}

type VFOFrequencyListener interface {
	VFOFrequencyChanged(VFOID, Frequency)
}

type VFOTunedListener interface {
	VFOTuned(VFOID, Frequency)
}

type VFOBandListener interface {
	VFOBandChanged(VFOID, Band)
}

type VFOModeListener interface {
	VFOModeChanged(VFOID, Mode)
}

type IncrementalTuningKind int

const (
	RIT IncrementalTuningKind = iota
	XIT
)

func (k IncrementalTuningKind) String() string {
	switch k {
	case RIT:
		return "RIT"
	case XIT:
		return "XIT"
	default:
		return "unknown"
	}
}

func (k IncrementalTuningKind) Workmode() Workmode {
	if k == RIT {
		return Run
	}
	return SearchPounce
}

type VFOIncrementalTuningListener interface {
	VFOIncrementalTuningChanged(vfo VFOID, kind IncrementalTuningKind, active bool, offset Frequency)
}

type VFOIncrementalTuningActiveListener interface {
	VFOIncrementalTuningActiveChanged(vfo VFOID, kind IncrementalTuningKind, active bool)
}

type VFOIncrementalTuningVisibilityListener interface {
	VFOIncrementalTuningVisibilityChanged(vfo VFOID, kind IncrementalTuningKind, visible bool)
}

type VFOPTTListener interface {
	VFOPTTChanged(VFOID, bool)
}

type RadioChangedListener interface {
	RadioChanged(name string, singleVFO bool)
}

type RadioChangedListenerFunc func(name string, singleVFO bool)

func (f RadioChangedListenerFunc) RadioChanged(name string, singleVFO bool) {
	f(name, singleVFO)
}

type ConnectionChangedListener interface {
	ConnectionChanged(bool)
}

type ConnectionChangedFunc func(bool)

func (f ConnectionChangedFunc) ConnectionChanged(connected bool) {
	f(connected)
}
