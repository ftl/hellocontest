package pb

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/locator"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/parse"
)

func ToQSO(pbQSO *QSO) (core.QSO, error) {
	var qso core.QSO
	var err error
	qso.Callsign, err = callsign.Parse(pbQSO.Callsign)
	if err != nil {
		return core.QSO{}, err
	}
	qso.Time = time.Unix(pbQSO.Timestamp, 0)
	qso.Frequency = core.Frequency(pbQSO.Frequency)
	qso.Band, err = parse.Band(pbQSO.Band)
	if err != nil {
		return core.QSO{}, err
	}
	qso.Mode, err = parse.Mode(pbQSO.Mode)
	if err != nil {
		return core.QSO{}, err
	}
	if pbQSO.MyReport != "" {
		qso.MyReport, err = parse.RST(pbQSO.MyReport)
		if err != nil {
			qso.MyReport = ""
		}
	}
	qso.MyNumber = core.QSONumber(pbQSO.MyNumber)
	qso.MyExchange = pbQSO.MyExchange
	if pbQSO.TheirReport != "" {
		qso.TheirReport, err = parse.RST(pbQSO.TheirReport)
		if err != nil {
			return core.QSO{}, err
		}
	}
	qso.TheirNumber = core.QSONumber(pbQSO.TheirNumber)
	qso.TheirExchange = pbQSO.TheirExchange
	qso.LogTimestamp = time.Unix(pbQSO.LogTimestamp, 0)
	qso.Workmode = core.Workmode(pbQSO.Workmode)
	return qso, nil
}

func QSOToPB(qso core.QSO) *QSO {
	return &QSO{
		Callsign:      qso.Callsign.String(),
		Timestamp:     qso.Time.Unix(),
		Frequency:     float64(qso.Frequency),
		Band:          qso.Band.String(),
		Mode:          qso.Mode.String(),
		MyReport:      qso.MyReport.String(),
		MyNumber:      int32(qso.MyNumber),
		MyExchange:    qso.MyExchange,
		TheirReport:   qso.TheirReport.String(),
		TheirNumber:   int32(qso.TheirNumber),
		TheirExchange: qso.TheirExchange,
		LogTimestamp:  qso.LogTimestamp.Unix(),
		Workmode:      Workmode(qso.Workmode),
	}
}

func ToStation(pbStation *Station) (core.Station, error) {
	var station core.Station
	var err error
	station.Callsign, err = callsign.Parse(pbStation.Callsign)
	if err != nil {
		log.Printf("Cannot parse station callsign: %v", err)
		station.Callsign = callsign.Callsign{}
	}
	station.Operator, err = callsign.Parse(pbStation.Operator)
	if err != nil {
		log.Printf("Cannot parse station operator: %v", err)
		station.Operator = callsign.Callsign{}
	}
	station.Locator, err = locator.Parse(pbStation.Locator)
	if err != nil {
		log.Printf("Cannot parse station locator: %v", err)
		station.Locator = locator.Locator{}
	}
	return station, nil
}

func StationToPB(station core.Station) *Station {
	return &Station{
		Callsign: station.Callsign.String(),
		Operator: station.Operator.String(),
		Locator:  station.Locator.String(),
	}
}

func ToContest(pbContest *Contest) (core.Contest, error) {
	var contest core.Contest

	contest.Name = pbContest.Name
	contest.StartTime = pbContest.StartTime.AsTime()
	contest.OperationModeSprint = pbContest.SprintOperation
	contest.CallHistoryFilename = pbContest.CallHistoryFilename
	contest.CallHistoryFieldNames = pbContest.CallHistoryFieldNames

	contest.ExchangeValues = pbContest.ExchangeValues
	contest.GenerateSerialExchange = pbContest.GenerateSerialExchange
	contest.GenerateReport = pbContest.GenerateReport
	contest.QSOsGoal = int(pbContest.QsosGoal)
	contest.PointsGoal = int(pbContest.PointsGoal)
	contest.MultisGoal = int(pbContest.MultisGoal)

	if pbContest.DefinitionYaml == "" {
		return contest, nil
	}

	buffer := bytes.NewBufferString(pbContest.DefinitionYaml)
	definition, err := conval.LoadDefinitionYAML(buffer)
	if err != nil {
		log.Print(err)
		return contest, nil
	}
	contest.Definition = definition

	return contest, nil
}

func ContestToPB(contest core.Contest) *Contest {
	definitionYaml := ""
	if contest.Definition != nil {
		buffer := bytes.NewBuffer([]byte{})
		err := conval.SaveDefinitionYAML(buffer, contest.Definition, false)
		if err != nil {
			log.Printf("Cannot store the contest definition: %v", err)
		} else {
			definitionYaml = buffer.String()
		}
	}

	return &Contest{
		DefinitionYaml:         definitionYaml,
		ExchangeValues:         contest.ExchangeValues,
		GenerateSerialExchange: contest.GenerateSerialExchange,
		GenerateReport:         contest.GenerateReport,

		Name:                  contest.Name,
		StartTime:             timestamppb.New(contest.StartTime),
		SprintOperation:       contest.OperationModeSprint,
		CallHistoryFilename:   contest.CallHistoryFilename,
		CallHistoryFieldNames: contest.CallHistoryFieldNames,
		QsosGoal:              int32(contest.QSOsGoal),
		PointsGoal:            int32(contest.PointsGoal),
		MultisGoal:            int32(contest.MultisGoal),
	}
}

func ToKeyerSettings(pbSettings *Keyer) (core.KeyerSettings, error) {
	var result core.KeyerSettings
	result.WPM = int(pbSettings.Wpm)
	result.SPLabels = pbSettings.SpLabels
	result.SPMacros = pbSettings.SpMacros
	result.RunLabels = pbSettings.RunLabels
	result.RunMacros = pbSettings.RunMacros
	result.ParrotIntervalSeconds = int(pbSettings.ParrotIntervalSeconds)
	return result, nil
}

func KeyerSettingsToPB(settings core.KeyerSettings) *Keyer {
	return &Keyer{
		Wpm:                   int32(settings.WPM),
		SpLabels:              settings.SPLabels,
		SpMacros:              settings.SPMacros,
		RunLabels:             settings.RunLabels,
		RunMacros:             settings.RunMacros,
		ParrotIntervalSeconds: int32(settings.ParrotIntervalSeconds),
	}
}

func ToQTC(pbQTC *QTC) (core.QTC, error) {
	var qtc core.QTC
	var err error

	qtc.Timestamp = time.Unix(pbQTC.Timestamp, 0)
	if qtc.Timestamp.IsZero() {
		return core.QTC{}, fmt.Errorf("QTC has no timestamp: %v", pbQTC)
	}
	qtc.Frequency = core.Frequency(pbQTC.Frequency)
	qtc.Band, err = parse.Band(pbQTC.Band)
	if err != nil {
		return core.QTC{}, err
	}
	qtc.Mode, err = parse.Mode(pbQTC.Mode)
	if err != nil {
		return core.QTC{}, err
	}

	qtc.Kind = core.QTCKind(pbQTC.Kind)
	qtc.QSONumber = core.QSONumber(pbQTC.QsoNumber)
	qtc.TheirCallsign, err = callsign.Parse(pbQTC.TheirCallsign)
	if err != nil {
		return core.QTC{}, err
	}

	if pbQTC.Header != "" {
		qtc.Header, err = core.ParseQTCHeader(pbQTC.Header)
		if err != nil {
			return core.QTC{}, err
		}
	}

	if pbQTC.QtcTime != "" {
		qtc.QTCTime, err = core.ParseQTCTime(pbQTC.QtcTime, core.ZeroQTCTime)
		if err != nil {
			return core.QTC{}, err
		}
	}

	qtc.QTCCallsign, err = callsign.Parse(pbQTC.QtcCallsign)
	if err != nil {
		return core.QTC{}, err
	}

	qtc.QTCNumber = core.QSONumber(pbQTC.QtcNumber)

	return qtc, nil
}

func QTCToPB(qtc core.QTC) *QTC {
	return &QTC{
		Timestamp:     qtc.Timestamp.Unix(),
		Frequency:     float64(qtc.Frequency),
		Band:          qtc.Band.String(),
		Mode:          qtc.Mode.String(),
		Kind:          int32(qtc.Kind),
		QsoNumber:     int32(qtc.QSONumber),
		TheirCallsign: qtc.TheirCallsign.String(),
		Header:        qtc.Header.String(),
		QtcTime:       qtc.QTCTime.String(),
		QtcCallsign:   qtc.QTCCallsign.String(),
		QtcNumber:     int32(qtc.QTCNumber),
	}
}
