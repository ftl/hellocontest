package model

import (
	"fmt"
	"time"

	"github.com/ftl/hamradio/locator"
)

// Clock represents a source of the current time.
type Clock interface {
	Now() time.Time
}

// FilterPlaceholder can be used as placeholder for a missed character in the callsign.
const FilterPlaceholder = "."

type Settings interface {
	Station() Station
	Contest() Contest
}

type Station struct {
	Callsign Callsign
	Operator Callsign
	Locator  locator.Locator
	Bandplan BandplanID
}

type AsyncRunner func(func())

type RemoteServerSettings struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type Service int

const (
	NoService Service = iota
	RadioService
	KeyerService
	DXCCService
	SCPService
	CallHistoryService
	MapService
	RemoteService
)

type ServiceStatusListener interface {
	StatusChanged(service Service, avialable bool)
}

type ServiceStatusListenerFunc func(Service, bool)

func (f ServiceStatusListenerFunc) StatusChanged(service Service, available bool) {
	f(service, available)
}

type CallsignEnteredListener interface {
	CallsignEntered(vfo VFOID, callsign string)
}

type CallsignLoggedListener interface {
	CallsignLogged(callsign string, frequency Frequency)
}

type TransmissionStartedListener interface {
	TransmissionStarted(vfo VFOID)
}

func FormatTimestamp(ts time.Time) string {
	return ts.UTC().Format("2006-01-02 15:04Z")
}

func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%02dh%02d", int(d.Hours()), int(d.Minutes())-int(d.Hours())*60)
}

func Emit[L any](listeners []any, notify func(listener L)) {
	for i := range listeners {
		listener, ok := listeners[i].(L)
		if ok {
			notify(listener)
		}
	}
}
