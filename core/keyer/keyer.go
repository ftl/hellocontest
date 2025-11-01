package keyer

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/ftl/hamradio/callsign"

	"github.com/ftl/hellocontest/core"
)

const PatternCount = 4

// ButtonView represents the visual parts of trigger the transmission of the keyer macros.
type ButtonView interface {
	ShowMessage(...any)
	SetLabel(int, string)
	SetPattern(int, string)
	SetSpeed(int)
}

// SettingsView represents the visual parts to enter keyer macros.
type SettingsView interface {
	Show()
	ShowMessage(...any)
	ClearMessage()
	SetLabel(core.Workmode, int, string)
	SetMacro(core.Workmode, int, string)
	SetPresetNames([]string)
	SetPreset(string)
	SetParrotIntervalSeconds(int)
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

// KeyerStoppedListener is notified when the keyer was actively stopped by the user.
type KeyerStoppedListener interface {
	KeyerStopped()
}

type Parrot interface {
	KeyerStoppedListener
	SetInterval(time.Duration)
}

// New returns a new Keyer that has no patterns or templates defined yet.
func New(settings core.Settings, client CWClient, keyerSettings core.KeyerSettings, workmode core.Workmode, presets []core.KeyerPreset) *Keyer {
	result := &Keyer{
		writer:          new(nullWriter),
		parrot:          new(nullParrot),
		stationCallsign: settings.Station().Callsign,
		workmode:        workmode,
		spLabels:        make(map[int]string),
		spPatterns:      make(map[int]string),
		spTemplates:     make(map[int]*template.Template),
		runLabels:       make(map[int]string),
		runPatterns:     make(map[int]string),
		runTemplates:    make(map[int]*template.Template),
		presets:         presets,
		client:          client,
		values:          noValues,
	}
	result.setWorkmode(workmode)
	result.SetSettings(keyerSettings, "")
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
	parrot         Parrot
	buttonView     ButtonView
	settingsView   SettingsView
	client         CWClient
	presets        []core.KeyerPreset
	presetNames    []string
	values         KeyerValueProvider
	savedSettings  core.KeyerSettings
	selectedPreset *core.KeyerPreset

	listeners []any

	stationCallsign       callsign.Callsign
	workmode              core.Workmode
	wpm                   int
	parrotIntervalSeconds int
	spLabels              map[int]string
	spPatterns            map[int]string
	spTemplates           map[int]*template.Template
	runLabels             map[int]string
	runPatterns           map[int]string
	runTemplates          map[int]*template.Template
	labels                *map[int]string
	patterns              *map[int]string
	templates             *map[int]*template.Template
	lastTransmission      string
}

func (k *Keyer) setWorkmode(workmode core.Workmode) {
	k.workmode = workmode
	switch workmode {
	case core.SearchPounce:
		k.labels = &k.spLabels
		k.patterns = &k.spPatterns
		k.templates = &k.spTemplates
	case core.Run:
		k.labels = &k.runLabels
		k.patterns = &k.runPatterns
		k.templates = &k.runTemplates
	}
}

func (k *Keyer) getTemplatesByWorkmode(workmode core.Workmode) map[int]*template.Template {
	switch workmode {
	case core.SearchPounce:
		return k.spTemplates
	case core.Run:
		return k.runTemplates
	default:
		return nil
	}
}

func (k *Keyer) SetWriter(writer Writer) {
	if writer == nil {
		k.writer = new(nullWriter)
		return
	}
	k.writer = writer
}

func (k *Keyer) SetParrot(parrot Parrot) {
	if parrot == nil {
		k.parrot = new(nullParrot)
		return
	}
	k.parrot = parrot
	k.parrot.SetInterval(time.Duration(k.parrotIntervalSeconds) * time.Second)
}

func (k *Keyer) SetSettings(settings core.KeyerSettings, presetName string) {
	k.savedSettings = settings

	spLabels := settings.SPLabels
	spMacros := settings.SPMacros
	runLabels := settings.RunLabels
	runMacros := settings.RunMacros

	preset, ok := k.presetByName(presetName)
	if !ok {
		preset, ok = k.presetByName(settings.Preset)
	}
	if ok {
		spLabels = applyPreset(settings.SPLabels, preset.SPLabels)
		spMacros = applyPreset(settings.SPMacros, preset.SPMacros)
		runLabels = applyPreset(settings.RunLabels, preset.RunLabels)
		runMacros = applyPreset(settings.RunMacros, preset.RunMacros)
		k.selectedPreset = &preset
	} else {
		k.selectedPreset = nil
	}

	k.wpm = settings.WPM
	for i, label := range spLabels {
		k.spLabels[i] = label
	}
	for i, pattern := range spMacros {
		k.spPatterns[i] = pattern
		k.spTemplates[i], _ = parseTemplate(pattern)
	}
	for i, label := range runLabels {
		k.runLabels[i] = label
	}
	for i, pattern := range runMacros {
		k.runPatterns[i] = pattern
		k.runTemplates[i], _ = parseTemplate(pattern)
	}

	k.parrotIntervalSeconds = settings.ParrotIntervalSeconds
	k.parrot.SetInterval(time.Duration(k.parrotIntervalSeconds) * time.Second)

	k.showPatterns()
	if k.buttonView != nil {
		k.buttonView.SetSpeed(k.wpm)
	}
	k.showKeyerSettings()
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

func (k *Keyer) SetView(view ButtonView) {
	k.buttonView = view
	k.showPatterns()
	k.buttonView.SetSpeed(k.wpm)
}

func (k *Keyer) showPatterns() {
	if k.buttonView == nil {
		return
	}
	for i, label := range *k.labels {
		k.buttonView.SetLabel(i, label)
	}
	for i, pattern := range *k.patterns {
		k.buttonView.SetPattern(i, pattern)
	}
}

func (k *Keyer) SetSettingsView(view SettingsView) {
	k.settingsView = view
	k.showKeyerSettings()
	k.settingsView.SetPresetNames(k.presetNames)
	if k.selectedPreset != nil {
		k.settingsView.SetPreset(k.selectedPreset.Name)
	} else {
		k.settingsView.SetPreset("")
	}
}

func (k *Keyer) OpenKeyerSettings() {
	if k.settingsView == nil {
		return
	}

	k.settingsView.Show()
	k.settingsView.SetPresetNames(k.presetNames)
	if k.selectedPreset != nil {
		k.settingsView.SetPreset(k.selectedPreset.Name)
	} else {
		k.settingsView.SetPreset("")
	}
	k.showKeyerSettings()
}

func (k *Keyer) showKeyerSettings() {
	if k.settingsView == nil {
		return
	}
	if k.selectedPreset == nil {
		k.settingsView.SetPreset("")
	} else {
		k.settingsView.SetPreset(k.selectedPreset.Name)
	}
	for i, label := range k.spLabels {
		k.settingsView.SetLabel(core.SearchPounce, i, label)
	}
	for i, pattern := range k.spPatterns {
		k.settingsView.SetMacro(core.SearchPounce, i, pattern)
	}
	for i, label := range k.runLabels {
		k.settingsView.SetLabel(core.Run, i, label)
	}
	for i, pattern := range k.runPatterns {
		k.settingsView.SetMacro(core.Run, i, pattern)
	}
	k.settingsView.SetParrotIntervalSeconds(k.parrotIntervalSeconds)
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
	keyer.ParrotIntervalSeconds = k.parrotIntervalSeconds
	keyer.SPLabels = make([]string, len(k.spLabels))
	for i := range keyer.SPLabels {
		label, ok := k.spLabels[i]
		if !ok {
			continue
		}
		keyer.SPLabels[i] = label
	}
	keyer.SPMacros = make([]string, len(k.spPatterns))
	for i := range keyer.SPMacros {
		pattern, ok := k.spPatterns[i]
		if !ok {
			continue
		}
		keyer.SPMacros[i] = pattern
	}
	keyer.RunLabels = make([]string, len(k.runLabels))
	for i := range keyer.RunLabels {
		label, ok := k.runLabels[i]
		if !ok {
			continue
		}
		keyer.RunLabels[i] = label
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
		k.settingsView.SetPreset("")
		return
	}
	preset := *k.selectedPreset
	k.settingsView.SetPreset(preset.Name)

	settings := core.KeyerSettings{
		WPM:                   k.savedSettings.WPM,
		ParrotIntervalSeconds: k.savedSettings.ParrotIntervalSeconds,
		Preset:                name,
		SPLabels:              make([]string, len(preset.SPLabels)),
		SPMacros:              make([]string, len(preset.SPMacros)),
		RunLabels:             make([]string, len(preset.RunLabels)),
		RunMacros:             make([]string, len(preset.RunMacros)),
	}
	copy(settings.SPLabels, preset.SPLabels)
	copy(settings.SPMacros, preset.SPMacros)
	copy(settings.RunLabels, preset.RunLabels)
	copy(settings.RunMacros, preset.RunMacros)
	k.SetSettings(settings, "")
	k.Save()
}

func (k *Keyer) EnterSpeed(speed int) {
	log.Printf("speed entered: %d", speed)
	k.wpm = speed
	k.client.Speed(k.wpm)
}

func (k *Keyer) EnterParrotIntervalSeconds(interval int) {
	log.Printf("parrot interval entered: %d", interval)
	k.parrotIntervalSeconds = interval
	k.parrot.SetInterval(time.Duration(k.parrotIntervalSeconds) * time.Second)
}

func (k *Keyer) EnterLabel(workmode core.Workmode, index int, text string) {
	switch workmode {
	case core.SearchPounce:
		k.spLabels[index] = text
	case core.Run:
		k.runLabels[index] = text
	}

	if workmode == k.workmode {
		k.buttonView.SetLabel(index, text)
	}
}

func (k *Keyer) EnterMacro(workmode core.Workmode, index int, pattern string) {
	t, err := parseTemplate(pattern)
	if err != nil {
		k.settingsView.ShowMessage(err)
	} else {
		k.settingsView.ClearMessage()
	}

	switch workmode {
	case core.SearchPounce:
		k.spPatterns[index] = pattern
		k.spTemplates[index] = t
	case core.Run:
		k.runPatterns[index] = pattern
		k.runTemplates[index] = t
	}

	if workmode == k.workmode {
		k.buttonView.SetPattern(index, pattern)
	}

	if k.selectedPreset == nil {
		return
	}

	presetPattern := ""
	switch workmode {
	case core.SearchPounce:
		presetPattern = k.selectedPreset.SPMacros[index]
	case core.Run:
		presetPattern = k.selectedPreset.RunMacros[index]
	}

	if presetPattern != pattern {
		k.selectedPreset = nil
		k.settingsView.SetPreset("")
	}
}

func (k *Keyer) EnterPattern(index int, pattern string) {
	(*k.patterns)[index] = pattern
	var err error
	(*k.templates)[index], err = parseTemplate(pattern)
	if err != nil {
		k.buttonView.ShowMessage(err)
	} else {
		k.buttonView.ShowMessage()
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
		k.settingsView.SetPreset("")
	}
}

func (k *Keyer) GetPattern(index int) string {
	return (*k.patterns)[index]
}

func (k *Keyer) GetText(workmode core.Workmode, index int) (string, error) {
	buffer := bytes.NewBufferString("")
	templates := k.getTemplatesByWorkmode(workmode)
	template, ok := templates[index]
	if !ok {
		return "", nil
	}
	err := template.Execute(buffer, k.fillins())
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (k *Keyer) fillins() map[string]any {
	values := k.values()
	result := map[string]any{
		"MyCall":      k.stationCallsign.String(),
		"MyReport":    cutDefault(values.MyReport.String()),
		"MyNumber":    cutDefault(pad(3, values.MyNumber.String())),
		"MyXchange":   values.MyXchange,
		"MyExchange":  values.MyExchange,
		"MyExchanges": values.MyExchanges,
		"TheirCall":   values.TheirCall,
	}
	return result
}

func (k *Keyer) Send(index int) {
	message, err := k.GetText(k.workmode, index)
	if err != nil {
		k.buttonView.ShowMessage(err)
		return
	}
	k.send(message)
}

func (k *Keyer) SendQuestion(q string) {
	s := strings.TrimSpace(q) + "?"
	k.send(s)
}

func (k *Keyer) SendText(text string, args ...any) {
	k.send(fmt.Sprintf(text, args...))
}

func (k *Keyer) send(s string) {
	log.Printf("sending %q", s)
	k.lastTransmission = s
	k.client.Send(s)
}

func (k *Keyer) Repeat() {
	k.send(k.lastTransmission)
}

func (k *Keyer) Stop() {
	log.Println("abort sending")
	k.client.Abort()
	k.parrot.KeyerStopped()
	k.emitKeyerStopped()
}

func (k *Keyer) Notify(listener any) {
	k.listeners = append(k.listeners, listener)
}

func (k *Keyer) emitKeyerStopped() {
	for _, listener := range k.listeners {
		if keyerStoppedListener, ok := listener.(KeyerStoppedListener); ok {
			keyerStoppedListener.KeyerStopped()
		}
	}
}

func parseTemplate(pattern string) (*template.Template, error) {
	funcs := map[string]any{
		"cut":     cutDefault,
		"cutOnly": cutOnly,
		"pad":     pad,
	}
	return template.New("").Funcs(funcs).Parse(pattern)
}

func cutDefault(value string) string {
	return cut(value)
}

func cutOnly(args ...any) string {
	if len(args) == 0 {
		return ""
	}
	value, valueOK := args[len(args)-1].(string)
	if !valueOK {
		return ""
	}

	numbers := make([]int, 0, len(args)-2)
	for i := range args[:len(args)-1] {
		number, ok := args[i].(int)
		if !ok {
			continue
		}
		numbers = append(numbers, number)
	}

	return cut(value, numbers...)
}

func cut(value string, numbers ...int) string {
	cuts := cutsFor(numbers...)
	result := value
	for digit, cut := range cuts {
		result = strings.Replace(result, digit, cut, -1)
	}
	return result
}

func cutsFor(numbers ...int) map[string]string {
	if len(numbers) == 0 {
		return map[string]string{
			"0": "t",
			"9": "n",
		}
	}

	allCuts := map[string]string{
		"0": "t",
		"1": "a",
		"2": "u",
		"3": "v",
		"5": "e",
		"7": "g",
		"8": "d",
		"9": "n",
	}
	cuts := make(map[string]string)
	for _, number := range numbers {
		i := strconv.Itoa(number)
		cut, ok := allCuts[i]
		if ok {
			cuts[i] = cut
		}
	}
	return cuts
}

func pad(length int, value string) string {
	result := value
	for len(result) < length {
		result = "0" + result
	}
	return result
}

func noValues() core.KeyerValues {
	return core.KeyerValues{}
}

type nullWriter struct{}

func (*nullWriter) WriteKeyer(core.KeyerSettings) error { return nil }

type nullClient struct{}

func (*nullClient) Connect() error    { return nil }
func (*nullClient) IsConnected() bool { return true }
func (*nullClient) Speed(int)         {}
func (*nullClient) Send(text string)  {}
func (*nullClient) Abort()            {}

type nullParrot struct{}

func (*nullParrot) KeyerStopped()             {}
func (*nullParrot) SetInterval(time.Duration) {}
