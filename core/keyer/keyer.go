package keyer

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

const PatternCount = 4

// View represents the visual parts of the keyer.
type View interface {
	ShowMessage(...interface{})
	SetPattern(int, string)
	SetSpeed(int)
	SetPresetNames([]string)
	SetPreset(string)
}

// CWClient defines the interface used by the Keyer to output the CW.
type CWClient interface {
	Speed(int)
	Send(text string)
	Abort()
}

// KeyerValueProvider provides the variable values for the Keyer templates on demand.
type KeyerValueProvider func() core.KeyerValues

type Writer interface {
	WriteKeyer(core.KeyerSettings) error
}

// New returns a new Keyer that has no patterns or templates defined yet.
func New(settings core.Settings, client CWClient, keyerSettings core.KeyerSettings, workmode core.Workmode, presets []core.KeyerPreset) *Keyer {
	result := &Keyer{
		writer:          new(nullWriter),
		stationCallsign: settings.Station().Callsign,
		workmode:        workmode,
		spPatterns:      make(map[int]string),
		spTemplates:     make(map[int]*template.Template),
		runPatterns:     make(map[int]string),
		runTemplates:    make(map[int]*template.Template),
		presets:         presets,
		client:          client,
		values:          noValues,
	}
	result.setWorkmode(workmode)
	result.SetSettings(keyerSettings)
	result.presetNames = presetNames(presets)
	if result.client == nil {
		result.client = new(nullClient)
	}
	return result
}

func presetNames(presets []core.KeyerPreset) []string {
	result := make([]string, len(presets))
	for i, preset := range presets {
		result[i] = preset.Name
	}
	return result
}

type Keyer struct {
	writer         Writer
	view           View
	client         CWClient
	presets        []core.KeyerPreset
	presetNames    []string
	values         KeyerValueProvider
	savedSettings  core.KeyerSettings
	selectedPreset *core.KeyerPreset

	listeners []any

	stationCallsign callsign.Callsign
	workmode        core.Workmode
	wpm             int
	spPatterns      map[int]string
	spTemplates     map[int]*template.Template
	runPatterns     map[int]string
	runTemplates    map[int]*template.Template
	patterns        *map[int]string
	templates       *map[int]*template.Template
}

func (k *Keyer) setWorkmode(workmode core.Workmode) {
	k.workmode = workmode
	switch workmode {
	case core.SearchPounce:
		k.patterns = &k.spPatterns
		k.templates = &k.spTemplates
	case core.Run:
		k.patterns = &k.runPatterns
		k.templates = &k.runTemplates
	}
}

func (k *Keyer) SetWriter(writer Writer) {
	if writer == nil {
		k.writer = new(nullWriter)
		return
	}
	k.writer = writer
}

func (k *Keyer) SetSettings(settings core.KeyerSettings) {
	k.savedSettings = settings

	spMacros := settings.SPMacros
	runMacros := settings.RunMacros

	preset, ok := k.presetByName(settings.Preset)
	if ok {
		spMacros = applyPreset(settings.SPMacros, preset.SPMacros)
		runMacros = applyPreset(settings.RunMacros, preset.RunMacros)
	}

	k.wpm = settings.WPM
	for i, pattern := range spMacros {
		k.spPatterns[i] = pattern
		k.spTemplates[i], _ = template.New("").Parse(pattern)
	}
	for i, pattern := range runMacros {
		k.runPatterns[i] = pattern
		k.runTemplates[i], _ = template.New("").Parse(pattern)
	}

	k.showPatterns()
	if k.view != nil {
		k.view.SetSpeed(k.wpm)
	}
}

func (k *Keyer) presetByName(name string) (core.KeyerPreset, bool) {
	normalizeName := func(name string) string {
		return strings.TrimSpace(strings.ToLower(name))
	}

	name = normalizeName(name)
	if name == "" {
		return core.KeyerPreset{}, false
	}

	for _, preset := range k.presets {
		if normalizeName(preset.Name) == name {
			return preset, true
		}
	}

	return core.KeyerPreset{}, false
}

func applyPreset(settingsPatterns []string, presetPatterns []string) []string {
	if len(presetPatterns) > PatternCount {
		presetPatterns = presetPatterns[:PatternCount]
	}
	if len(strings.TrimSpace(strings.Join(settingsPatterns, ""))) != 0 {
		return settingsPatterns
	}

	result := make([]string, PatternCount)
	copy(result, presetPatterns)
	return result
}

func (k *Keyer) SetView(view View) {
	k.view = view
	k.showPatterns()
	k.view.SetPresetNames(k.presetNames)
	k.view.SetSpeed(k.wpm)
}

func (k *Keyer) showPatterns() {
	if k.view == nil {
		return
	}
	for i, pattern := range *k.patterns {
		k.view.SetPattern(i, pattern)
	}
}

func (k *Keyer) WorkmodeChanged(workmode core.Workmode) {
	k.setWorkmode(workmode)
	k.showPatterns()
}

func (k *Keyer) StationChanged(station core.Station) {
	k.stationCallsign = station.Callsign
}

func (k *Keyer) SetValues(values KeyerValueProvider) {
	k.values = values
}

func (k *Keyer) Save() {
	keyer, modified := k.getKeyerSettings()
	if !modified {
		return
	}
	k.savedSettings = keyer
	k.writer.WriteKeyer(keyer)
}

func (k *Keyer) KeyerSettings() core.KeyerSettings {
	keyer, _ := k.getKeyerSettings()
	return keyer
}

func (k *Keyer) getKeyerSettings() (core.KeyerSettings, bool) {
	var keyer core.KeyerSettings
	keyer.WPM = k.wpm
	keyer.SPMacros = make([]string, len(k.spPatterns))
	for i := range keyer.SPMacros {
		pattern, ok := k.spPatterns[i]
		if !ok {
			continue
		}
		keyer.SPMacros[i] = pattern
	}
	keyer.RunMacros = make([]string, len(k.runPatterns))
	for i := range keyer.RunMacros {
		pattern, ok := k.runPatterns[i]
		if !ok {
			continue
		}
		keyer.RunMacros[i] = pattern
	}

	modified := (fmt.Sprintf("%v", keyer) != fmt.Sprintf("%v", k.savedSettings))
	return keyer, modified
}

func (k *Keyer) SelectPreset(name string) {
	k.selectedPreset = nil
	for _, preset := range k.presets {
		if preset.Name == name {
			k.selectedPreset = &preset
			break
		}
	}
	if k.selectedPreset == nil {
		k.view.SetPreset("")
		return
	}
	preset := *k.selectedPreset
	k.view.SetPreset(preset.Name)

	settings := core.KeyerSettings{
		WPM:       k.savedSettings.WPM,
		Preset:    name,
		SPMacros:  make([]string, len(preset.SPMacros)),
		RunMacros: make([]string, len(preset.RunMacros)),
	}
	copy(settings.SPMacros, preset.SPMacros)
	copy(settings.RunMacros, preset.RunMacros)
	k.SetSettings(settings)
	k.Save()
}

func (k *Keyer) EnterSpeed(speed int) {
	log.Printf("speed entered: %d", speed)
	k.wpm = speed
	k.client.Speed(k.wpm)
}

func (k *Keyer) EnterPattern(index int, pattern string) {
	(*k.patterns)[index] = pattern
	var err error
	(*k.templates)[index], err = template.New("").Parse(pattern)
	if err != nil {
		k.view.ShowMessage(err)
	} else {
		k.view.ShowMessage()
	}

	if k.selectedPreset == nil {
		return
	}

	presetPattern := ""
	switch k.workmode {
	case core.SearchPounce:
		presetPattern = k.selectedPreset.SPMacros[index]
	case core.Run:
		presetPattern = k.selectedPreset.RunMacros[index]
	}

	if presetPattern != pattern {
		k.selectedPreset = nil
		k.view.SetPreset("")
	}
}

func (k *Keyer) GetPattern(index int) string {
	return (*k.patterns)[index]
}

func (k *Keyer) GetText(index int) (string, error) {
	buffer := bytes.NewBufferString("")
	template, ok := (*k.templates)[index]
	if !ok {
		return "", nil
	}
	err := template.Execute(buffer, k.fillins())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (k *Keyer) fillins() map[string]string {
	values := k.values()
	result := map[string]string{
		"MyCall":     k.stationCallsign.String(),
		"MyReport":   softcut(values.MyReport.String()),
		"MyNumber":   softcut(values.MyNumber.String()),
		"MyXchange":  values.MyXchange,
		"MyExchange": values.MyExchange,
		"TheirCall":  values.TheirCall,
	}
	for i, exchange := range values.MyExchanges {
		key := fmt.Sprintf("MyExchange%d", i+1)
		result[key] = exchange
		_, err := strconv.Atoi(exchange)
		if err == nil {
			intKey := key + "Number"
			result[intKey] = softcut(exchange)
		}
	}
	return result
}

func (k *Keyer) Send(index int) {
	message, err := k.GetText(index)
	if err != nil {
		k.view.ShowMessage(err)
		return
	}
	k.send(message)
	log.Printf("Sending %s", message)
}

func (k *Keyer) SendQuestion(q string) {
	s := strings.TrimSpace(q) + "?"
	k.send(s)
}

func (k *Keyer) send(s string) {
	log.Printf("sending %s", s)
	k.client.Send(s)
}

func (k *Keyer) Stop() {
	log.Println("abort sending")
	k.client.Abort()
}

func (k *Keyer) Notify(listener any) {
	k.listeners = append(k.listeners, listener)
}

// softcut replaces 0 and 9 with their "cut" counterparts t and n.
func softcut(s string) string {
	cuts := map[string]string{
		"0": "t",
		"9": "n",
	}
	result := s
	for digit, cut := range cuts {
		result = strings.Replace(result, digit, cut, -1)
	}
	return result
}

// cut replaces digits with the "cut" counterparts. (see http://wiki.bavarian-contest-club.de/wiki/Contest-FAQ#Was_sind_.22Cut_Numbers.22.3F)
func cut(s string) string {
	cuts := map[string]string{
		"0": "t",
		"1": "a",
		"2": "u",
		"3": "v",
		"5": "e",
		"7": "g",
		"8": "d",
		"9": "n",
	}
	result := s
	for digit, cut := range cuts {
		result = strings.Replace(result, digit, cut, -1)
	}
	return result
}

func noValues() core.KeyerValues {
	return core.KeyerValues{}
}

type nullWriter struct{}

func (w *nullWriter) WriteKeyer(core.KeyerSettings) error { return nil }

type nullClient struct{}

func (*nullClient) Connect() error    { return nil }
func (*nullClient) IsConnected() bool { return true }
func (*nullClient) Speed(int)         {}
func (*nullClient) Send(text string)  {}
func (*nullClient) Abort()            {}
