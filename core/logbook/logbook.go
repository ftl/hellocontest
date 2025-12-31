package logbook

import (
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/scp"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/dxcc"
)

type DXCCEntities interface {
	Find(string) (dxcc.Prefix, bool)
}

type Logbook struct {
	clock core.Clock

	listeners []any

	dataLock          *sync.RWMutex
	qsos              []core.QSO
	sentQTCs          map[core.QSONumber]core.QTC // TODO: better store the QTC series?
	receivedQTCs      []core.QTC                  // TODO: better store QTC series?
	sentQTCsPerSeries []int
	availableQTCs     []core.QTC
	qtcsByCall        map[callsign.Callsign]int

	scoreCounter *scoreCounter
	dupes        dupeIndex     // used to find duplicate QSOs for a given callsign, band and mode, according to the contest rules
	worked       dupeIndex     // used to find worked QSOs for a given callsign
	callsigns    *scp.Database // used to find worked callsigns similar to a given string, e.g. for supercheck
	bandRule     conval.BandRule
}

func NewLogbook(clock core.Clock, settings core.Settings, entities DXCCEntities) *Logbook {
	result := &Logbook{
		clock:        clock,
		dataLock:     new(sync.RWMutex),
		scoreCounter: newScoreCounter(settings, entities),
		dupes:        make(dupeIndex),
		worked:       make(dupeIndex),
		callsigns:    scp.NewDatabase(),
	}

	result.clear(1000, 0)

	return result
}

// Settings

func (l *Logbook) StationChanged(station core.Station) {
	score, updated := l.stationChangedLocked(station)
	if updated {
		l.emitScoreChanged(score)
	}
}

func (l *Logbook) stationChangedLocked(station core.Station) (core.Score, bool) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	l.scoreCounter.StationChanged(station)

	if !l.scoreCounter.Valid() {
		return core.Score{}, false
	}

	score := l.refreshDerivedData()
	return score, true
}

func (l *Logbook) ContestChanged(contest core.Contest) {
	score, updated := l.contestChangedLocked(contest)

	if updated {
		l.emitScoreChanged(score)
	}
}

func (l *Logbook) contestChangedLocked(contest core.Contest) (core.Score, bool) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	l.scoreCounter.ContestChanged(contest)

	if contest.Definition != nil {
		l.bandRule = contest.Definition.Scoring.QSOBandRule
	}

	if !l.scoreCounter.Valid() {
		return core.Score{}, false
	}

	score := l.refreshDerivedData()
	return score, true
}

// Loading

func (l *Logbook) Load(qsos []core.QSO, qtcs []core.QTC) error {
	loadedQSOs, loadedQTCs, score, err := l.loadLocked(qsos, qtcs)
	if err != nil {
		return err
	}

	l.emitNotificationsAfterRefresh(loadedQSOs, loadedQTCs, score)

	return nil
}

func (l *Logbook) loadLocked(qsos []core.QSO, qtcs []core.QTC) ([]core.QSO, []core.QTC, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	// clear
	l.clear(len(qsos)*2, len(qtcs)*2)

	// load the QSOs
	err := l.putQSOs(qsos)
	if err != nil {
		return nil, nil, core.Score{}, err
	}

	// load the QTCs
	err = l.putQTCs(qtcs)
	if err != nil {
		return nil, nil, core.Score{}, err
	}

	// update first, because the QSO score might change during the update
	score := l.refreshDerivedData()

	// copy QSOs and QTCs for further processing outside the dataLock
	loadedQSOs := l.allQSOs()
	loadedQTCs := l.allQTCs()

	return loadedQSOs, loadedQTCs, score, nil
}

func (l *Logbook) clear(capQSOs int, capQTCs int) {
	l.qsos = make([]core.QSO, 0, capQSOs)
	l.sentQTCs = make(map[core.QSONumber]core.QTC, capQTCs)
	l.receivedQTCs = make([]core.QTC, 0, capQTCs)
	l.sentQTCsPerSeries = make([]int, 0, capQTCs)
	l.availableQTCs = make([]core.QTC, 0, capQSOs)
	l.qtcsByCall = make(map[callsign.Callsign]int)
}

func (l *Logbook) putQSOs(qsos []core.QSO) error {
	for _, qso := range qsos {
		err := l.putQSO(qso)
		if err != nil {
			return err
		}
		l.addAvailableQTCForQSO(qso)
	}
	return nil
}

func (l *Logbook) putQSO(qso core.QSO) error {
	if qso.MyNumber <= 0 {
		return fmt.Errorf("Logbook.putQSO: invalid QSO number %d", qso.MyNumber)
	}

	lastNumber := l.lastQSONumber()
	if (lastNumber <= 0) || (qso.MyNumber > lastNumber) {
		l.qsos = append(l.qsos, qso)
		return nil
	}

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	}
	l.qsos[index] = qso
	return nil
}

func (l *Logbook) putQTCs(qtcs []core.QTC) error {
	for _, qtc := range qtcs {
		err := l.addQTC(qtc)
		if err != nil {
			return err
		}
	}
	return nil
}

// QSOs

func (l *Logbook) AddQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	qso, score, err := l.addQSOLocked(qso)
	if err != nil {
		// this must never happen
		panic(err)
	}
	l.emitQSOAdded(qso)
	l.emitScoreChanged(score)
}

func (l *Logbook) addQSOLocked(qso core.QSO) (core.QSO, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return core.QSO{}, core.Score{}, fmt.Errorf("Logbook.addQSO: invalid QSO number %d", qso.MyNumber)
	}

	lastNumber := l.lastQSONumber()
	if (lastNumber <= 0) || (qso.MyNumber > lastNumber) {
		qso = l.scoreCounter.AddQSO(qso)
		l.qsos = append(l.qsos, qso)
		l.addDerivedDataForQSO(qso)
		score := l.scoreCounter.Score()
		l.addAvailableQTCForQSO(qso)

		return qso, score, nil
	}

	// find the right index and insert the QSO
	index, found := l.findQSOIndex(qso.MyNumber)
	if found {
		return core.QSO{}, core.Score{}, fmt.Errorf("Logbook.addQSO: a QSO with number %d already exists, cannot be added again, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	qso = l.scoreCounter.AddQSO(qso)
	l.qsos = append(l.qsos[:index+1], l.qsos[index:]...)
	l.qsos[index] = qso
	l.addDerivedDataForQSO(qso)
	score := l.scoreCounter.Score()
	l.addAvailableQTCForQSO(qso)

	return qso, score, nil
}

func (l *Logbook) lastQSONumber() core.QSONumber {
	lastIndex := len(l.qsos) - 1
	if lastIndex < 0 {
		return 0
	}
	return l.qsos[lastIndex].MyNumber
}

func (l *Logbook) findQSOIndex(number core.QSONumber) (int, bool) {
	low := 0
	high := len(l.qsos) - 1

	for low <= high {
		median := (low + high) / 2

		if l.qsos[median].MyNumber < number {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(l.qsos) || l.qsos[low].MyNumber != number {
		return low, false
	}

	return low, true
}

func (l *Logbook) UpdateQSO(qso core.QSO) {
	qso.LogTimestamp = l.clock.Now()
	updatedQSOs, updatedQTCs, score, err := l.updateQSOLocked(qso)
	if err != nil {
		// this must never happen
		panic(err)
	}

	l.emitNotificationsAfterRefresh(updatedQSOs, updatedQTCs, score)
}

func (l *Logbook) updateQSOLocked(qso core.QSO) ([]core.QSO, []core.QTC, core.Score, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()
	if qso.MyNumber <= 0 {
		return nil, nil, core.Score{}, fmt.Errorf("Logbook.updateQSO: invalid QSO number %d", qso.MyNumber)
	}

	index, found := l.findQSOIndex(qso.MyNumber)
	if !found {
		return nil, nil, core.Score{}, fmt.Errorf("Logbook.updateQSO: the QSO with number %d cannot be found for update, use Logbook.UpdateQSO instead", qso.MyNumber)
	}
	l.qsos[index] = qso

	// update derived data first, because the QSO score might change during the update
	score := l.refreshDerivedData()

	// copy QSOs and QTCs for further processing outside the dataLock
	updatedQSOs := l.allQSOs()
	updatedQTCs := l.allQTCs()

	return updatedQSOs, updatedQTCs, score, nil
}

func (l *Logbook) AllQSOs() []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.allQSOs()
}

func (l *Logbook) allQSOs() []core.QSO {
	result := make([]core.QSO, len(l.qsos))
	copy(result, l.qsos)
	return result
}

func (l *Logbook) NextQSONumber() core.QSONumber {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.lastQSONumber() + 1
}

func (l *Logbook) lastQSOLocked() core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	if len(l.qsos) == 0 {
		return core.QSO{}
	}
	return l.qsos[len(l.qsos)-1]
}

func (l *Logbook) LastCallsign() callsign.Callsign {
	return l.lastQSOLocked().Callsign
}

func (l *Logbook) LastBand() core.Band {
	return l.lastQSOLocked().Band
}

func (l *Logbook) LastMode() core.Mode {
	return l.lastQSOLocked().Mode
}

func (l *Logbook) LastExchange() []string {
	return l.lastQSOLocked().MyExchange
}

// QTCs

func (l *Logbook) AddQTCSeries(series core.QTCSeries) {
	addedQTCs, err := l.addQTCSeriesLocked(series)
	if err != nil {
		// this must never happen, otherwise we have made a mistake
		panic(err)
	}

	for _, qtc := range addedQTCs {
		l.emitQTCAdded(qtc)
	}
}

func (l *Logbook) addQTCSeriesLocked(series core.QTCSeries) ([]core.QTC, error) {
	l.dataLock.Lock()
	defer l.dataLock.Unlock()

	addedQTCs := make([]core.QTC, 0, len(series.QTCs))
	for _, qtc := range series.QTCs {
		qtc.TheirCallsign = series.TheirCallsign
		qtc.Header = series.Header
		err := l.addQTC(qtc)
		if err != nil {
			return nil, err
		}

		// TODO: update score

		addedQTCs = append(addedQTCs, qtc)
	}

	return addedQTCs, nil
}

func (l *Logbook) addQTC(qtc core.QTC) error {
	if err := qtc.VerifyComplete(); err != nil {
		return fmt.Errorf("Logbook.addQTC: cannot log QTC %v: %w", qtc, err)
	}

	if qtc.Kind == core.SentQTC {
		if existing, ok := l.sentQTCs[qtc.QSONumber]; ok {
			return fmt.Errorf("Logbook.addQTC: QTC for QSO #%d already exists, cannot log another QTC for the same QSO: %v", qtc.QSONumber, existing)
		}
		l.sentQTCs[qtc.QSONumber] = qtc
		err := l.registerSentQTCSeries(qtc)
		if err != nil {
			return err
		}

		l.removeAvailableQTC(qtc)
	} else {
		l.receivedQTCs = append(l.receivedQTCs, qtc)
	}

	// TODO: mabye we need to count sent and received qtcs separateley if the contest allows both directions
	count := l.qtcsByCall[qtc.TheirCallsign]
	count++
	l.qtcsByCall[qtc.TheirCallsign] = count

	log.Printf("QTC added: %v", qtc)
	return nil
}

func (l *Logbook) registerSentQTCSeries(qtc core.QTC) error {
	if qtc.Kind != core.SentQTC {
		return nil
	}

	index := qtc.Header.SeriesNumber - 1
	switch {
	case index < 0: // this must never happen
		return fmt.Errorf("Logbook.registerQTCSeries: invalid QTC series number %d, must be greater than 0", qtc.Header.SeriesNumber)
	case len(l.sentQTCsPerSeries) == index: // the first of a new series
		l.sentQTCsPerSeries = append(l.sentQTCsPerSeries, 1)
	case len(l.sentQTCsPerSeries) > index: // the next of an existing series
		l.sentQTCsPerSeries[index]++
		// TODO: check if the series contains more than Header.QTCCount
	default: // this must never happen, the calculation of the next series number is broken
		return fmt.Errorf("Logbook.registerQTCSeries: unknown QTC series number %d, should not be greater than %d", qtc.Header.SeriesNumber, len(l.sentQTCsPerSeries))
	}
	return nil
}

func (l *Logbook) addAvailableQTCForQSO(qso core.QSO) {
	qtc := sentQTCFromQSO(qso)
	for i, availableQTC := range l.availableQTCs {
		if availableQTC.QSONumber == qso.MyNumber {
			l.availableQTCs[i] = qtc
			return
		}
	}
	l.availableQTCs = append(l.availableQTCs, qtc)
}

func (l *Logbook) removeAvailableQTC(qtc core.QTC) {
	if qtc.Kind != core.SentQTC {
		return
	}
	index := -1
	for i := range l.availableQTCs {
		if qtc.QSONumber == l.availableQTCs[i].QSONumber {
			index = i
			break
		}
	}
	switch {
	case index < 0:
	case index < len(l.availableQTCs)-1:
		l.availableQTCs = append(l.availableQTCs[:index], l.availableQTCs[index+1:]...)
	case index == len(l.availableQTCs)-1:
		l.availableQTCs = l.availableQTCs[:index]
	}
}

func sentQTCFromQSO(qso core.QSO) core.QTC {
	return core.QTC{
		Kind:        core.SentQTC,
		QSONumber:   qso.MyNumber,
		QTCTime:     core.QTCTimeFromTimestamp(qso.Time),
		QTCCallsign: qso.Callsign,
		QTCNumber:   qso.TheirNumber,
	}
}

func (l *Logbook) AllQTCs() []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.allQTCs()
}

func (l *Logbook) allQTCs() []core.QTC {
	result := make([]core.QTC, 0, len(l.sentQTCs)+len(l.receivedQTCs))
	for _, qtc := range l.sentQTCs {
		result = append(result, qtc)
	}
	result = append(result, l.receivedQTCs...)
	slices.SortStableFunc(result, core.QTCByTimestamp)
	return result
}

func (l *Logbook) AvailableFor(theirCall callsign.Callsign) int {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	theirCallStr := theirCall.String()
	theirQTCCount := l.qtcsByCall[theirCall]
	theirQSOCount := 0
	for _, qtc := range l.availableQTCs {
		if qtc.QTCCallsign.String() == theirCallStr {
			theirQSOCount++
		}
	}
	return min(core.MaxQTCsPerCall-theirQTCCount, len(l.availableQTCs)-theirQSOCount)
}

func (l *Logbook) PrepareFor(theirCall callsign.Callsign, count int) []core.QTC {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	theirCallStr := theirCall.String()
	theirQTCCount := l.qtcsByCall[theirCall]
	maxLen := max(0, min(core.MaxQTCsPerCall-theirQTCCount, count))

	result := make([]core.QTC, 0, maxLen)
	for _, qtc := range l.availableQTCs {
		if len(result) >= maxLen {
			break
		}
		if qtc.QTCCallsign.String() != theirCallStr {
			qtc.TheirCallsign = theirCall
			result = append(result, qtc)
		}
	}

	return result
}

func (l *Logbook) NextSeriesNumber() int {
	return len(l.sentQTCsPerSeries) + 1
}

// Derived Data

func (l *Logbook) Score() core.Score {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.scoreCounter.Score()
}

func (l *Logbook) refreshDerivedData() core.Score {
	l.scoreCounter.Clear()
	l.dupes = make(dupeIndex)
	l.worked = make(dupeIndex)
	l.callsigns = scp.NewDatabase()

	for i, qso := range l.qsos {
		qso = l.scoreCounter.AddQSO(qso)
		l.qsos[i] = qso
		l.addDerivedDataForQSO(qso)
	}
	score := l.scoreCounter.Score()

	return score
}

func (l *Logbook) addDerivedDataForQSO(qso core.QSO) {
	dupeBand, dupeMode := l.dupeBandAndMode(qso.Band, qso.Mode)
	l.dupes.Add(qso.Callsign, dupeBand, dupeMode, qso.MyNumber)
	l.worked.Add(qso.Callsign, core.NoBand, core.NoMode, qso.MyNumber)
	l.callsigns.Add(qso.Callsign.String())
}

func (l *Logbook) dupeBandAndMode(band core.Band, mode core.Mode) (core.Band, core.Mode) {
	switch l.bandRule {
	case conval.Once:
		return core.NoBand, core.NoMode
	case conval.OncePerBand:
		return band, core.NoMode
	case conval.OncePerBandAndMode:
		return band, mode
	default:
		return core.NoBand, core.NoMode
	}
}

func (l *Logbook) FindDuplicateQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	band, mode = l.dupeBandAndMode(band, mode)
	numbers := l.dupes.Get(callsign, band, mode)

	return l.getQSOs(numbers)
}

func (l *Logbook) FindWorkedQSOs(callsign callsign.Callsign, band core.Band, mode core.Mode) ([]core.QSO, bool) {
	qsos := l.findWorkedQSOsLocked(callsign)
	if len(qsos) == 0 {
		return qsos, false
	}

	duplicate := false
	for _, qso := range qsos {
		switch l.bandRule {
		case conval.Once:
			duplicate = true
		case conval.OncePerBand:
			duplicate = (qso.Band == band)
		case conval.OncePerBandAndMode:
			duplicate = (qso.Band == band) && (qso.Mode == mode)
		default:
			duplicate = false
		}
		if duplicate {
			break
		}
	}
	return qsos, duplicate
}

func (l *Logbook) findWorkedQSOsLocked(callsign callsign.Callsign) []core.QSO {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	numbers := l.worked.Get(callsign, core.NoBand, core.NoMode)
	return l.getQSOs(numbers)
}

func (l *Logbook) getQSOs(numbers []core.QSONumber) []core.QSO {
	result := make([]core.QSO, 0, len(numbers))
	for _, n := range numbers {
		listIndex, found := l.findQSOIndex(n)
		if !found {
			log.Printf("QSO number %d not found", n)
			continue
		}
		qso := l.qsos[listIndex]
		if len(result) > 0 && n > result[len(result)-1].MyNumber {
			result = append(result, qso)
		} else {
			resultIndex, found := findIndex(result, n)
			if !found {
				result = append(result[:resultIndex+1], result[resultIndex:]...)
			}
			result[resultIndex] = qso
		}
	}
	return result
}

func findIndex(list []core.QSO, number core.QSONumber) (int, bool) {
	low := 0
	high := len(list) - 1

	for low <= high {
		median := (low + high) / 2

		if list[median].MyNumber < number {
			low = median + 1
		} else {
			high = median - 1
		}
	}

	if low == len(list) || list[low].MyNumber != number {
		return low, false
	}

	return low, true
}

func (l *Logbook) Find(s string) ([]core.AnnotatedCallsign, error) {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	if l.callsigns == nil {
		return nil, nil
	}
	matches, err := l.callsigns.Find(s)
	if err != nil {
		return nil, err
	}

	return toAnnotatedCallsigns(matches), nil
}

func toAnnotatedCallsigns(matches []scp.Match) []core.AnnotatedCallsign {
	result := make([]core.AnnotatedCallsign, 0, len(matches))

	for _, match := range matches {
		annotatedCallsign, err := toAnnotatedCallsign(match)
		if err != nil {
			log.Print(err)
			continue
		}
		result = append(result, annotatedCallsign)
	}

	return result
}

func toAnnotatedCallsign(match scp.Match) (core.AnnotatedCallsign, error) {
	cs, err := callsign.Parse(match.Key())
	if err != nil {
		return core.AnnotatedCallsign{}, nil
	}
	return core.AnnotatedCallsign{
		Callsign:   cs,
		Assembly:   toMatchingAssembly(match),
		Comparable: match,
		Compare: func(a any, b any) bool {
			aMatch, aOk := a.(scp.Match)
			bMatch, bOk := b.(scp.Match)
			if !aOk || !bOk {
				return false
			}
			return aMatch.LessThan(bMatch)
		},
	}, nil
}

func toMatchingAssembly(match scp.Match) core.MatchingAssembly {
	result := make(core.MatchingAssembly, len(match.Assembly))

	for i, part := range match.Assembly {
		result[i] = core.MatchingPart{OP: core.MatchingOperation(part.OP), Value: part.Value}
	}

	return result
}

func (l *Logbook) FillSummary(summary *core.Summary) {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	l.scoreCounter.FillSummary(summary)
}

func (l *Logbook) Value(callsign callsign.Callsign, entity dxcc.Prefix, band core.Band, mode core.Mode, exchange []string) (points, multis int, multiValues map[conval.Property]string) {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	return l.scoreCounter.Value(callsign, entity, band, mode, exchange)
}

// Save as

func (l *Logbook) WriteAll(writer Writer) error {
	l.dataLock.RLock()
	defer l.dataLock.RUnlock()

	err := l.writeAllQSOs(writer)
	if err != nil {
		return err
	}
	return l.writeAllQTCs(writer)
}

func (l *Logbook) writeAllQSOs(writer Writer) error {
	for _, qso := range l.qsos {
		err := writer.WriteQSO(qso)
		if err != nil {
			return fmt.Errorf("cannot write QSO %v: %w", qso, err)
		}
	}
	return nil
}

func (l *Logbook) writeAllQTCs(writer Writer) error {
	for _, qtc := range l.allQTCs() {
		err := writer.WriteQTC(qtc)
		if err != nil {
			return fmt.Errorf("cannot write QTC %v: %w", qtc, err)
		}
	}
	return nil
}

// Notifications

func (l *Logbook) Notify(listener any) {
	l.listeners = append(l.listeners, listener)
}

func (l *Logbook) emitNotificationsAfterRefresh(qsos []core.QSO, qtcs []core.QTC, score core.Score) {
	l.emitLogbookCleared()
	for _, qso := range qsos {
		l.emitQSOAdded(qso)
	}
	for _, qtc := range qtcs {
		l.emitQTCAdded(qtc)
	}
	l.emitScoreChanged(score)
}

func (l *Logbook) emitLogbookCleared() {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(LogbookClearedListener); ok {
			listener.LogbookCleared()
		}
	}
}

func (l *Logbook) emitQSOAdded(qso core.QSO) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(QSOAddedListener); ok {
			listener.QSOAdded(qso)
		}
	}
}

func (l *Logbook) emitQTCAdded(qtc core.QTC) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(QTCAddedListener); ok {
			listener.QTCAdded(qtc)
		}
	}
}

func (l *Logbook) emitScoreChanged(score core.Score) {
	for _, rawListener := range l.listeners {
		if listener, ok := rawListener.(ScoreChangedListener); ok {
			listener.ScoreChanged(score)
		}
	}
}
