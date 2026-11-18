package pb

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/locator"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ftl/hellocontest/core/model"
)

func ToQSO(pbQSO *QSO) (model.QSO, error) {
	var qso model.QSO
	var err error
	qso.Callsign, err = model.ParseCallsign(pbQSO.Callsign)
	if err != nil {
		return model.QSO{}, err
	}
	qso.Time = time.Unix(pbQSO.Timestamp, 0)
	qso.Frequency = model.Frequency(pbQSO.Frequency)
	qso.Band, err = model.ParseBand(pbQSO.Band)
	if err != nil {
		return model.QSO{}, err
	}
	qso.Mode, err = model.ParseMode(pbQSO.Mode)
	if err != nil {
		return model.QSO{}, err
	}
	if pbQSO.MyReport != "" {
		qso.MyReport, err = model.ParseRST(pbQSO.MyReport)
		if err != nil {
			qso.MyReport = ""
		}
	}
	qso.MyNumber = model.QSONumber(pbQSO.MyNumber)
	qso.MyExchange = pbQSO.MyExchange
	if pbQSO.TheirReport != "" {
		qso.TheirReport, err = model.ParseRST(pbQSO.TheirReport)
		if err != nil {
			return model.QSO{}, err
		}
	}
	qso.TheirNumber = model.QSONumber(pbQSO.TheirNumber)
	qso.TheirExchange = pbQSO.TheirExchange
	qso.LogTimestamp = time.Unix(pbQSO.LogTimestamp, 0)
	qso.Workmode = model.Workmode(pbQSO.Workmode)
	return qso, nil
}

func QSOToPB(qso model.QSO) *QSO {
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

func ToStation(pbStation *Station) (model.Station, error) {
	var station model.Station
	var err error
	station.Callsign, err = model.ParseCallsign(pbStation.Callsign)
	if err != nil {
		log.Printf("Cannot parse station callsign: %v", err)
		station.Callsign = model.Callsign{}
	}
	station.Operator, err = model.ParseCallsign(pbStation.Operator)
	if err != nil {
		log.Printf("Cannot parse station operator: %v", err)
		station.Operator = model.Callsign{}
	}
	station.Locator, err = locator.Parse(pbStation.Locator)
	if err != nil {
		log.Printf("Cannot parse station locator: %v", err)
		station.Locator = locator.Locator{}
	}
	station.Bandplan = model.BandplanID(pbStation.Bandplan)
	if station.Bandplan == "" {
		station.Bandplan = model.DefaultBandplanID
	}
	return station, nil
}

func StationToPB(station model.Station) *Station {
	return &Station{
		Callsign: station.Callsign.String(),
		Operator: station.Operator.String(),
		Locator:  station.Locator.String(),
		Bandplan: station.Bandplan.String(),
	}
}

func ToContest(pbContest *Contest) (model.Contest, error) {
	var contest model.Contest

	contest.Name = pbContest.Name
	contest.StartTime = pbContest.StartTime.AsTime()
	contest.OperationModeSprint = pbContest.SprintOperation
	contest.CallHistoryFilename = pbContest.CallHistoryFilename
	contest.CallHistoryFieldNames = pbContest.CallHistoryFieldNames

	contest.ExchangeValues = pbContest.ExchangeValues
	contest.GenerateSerialExchange = pbContest.GenerateSerialExchange
	contest.GenerateReport = pbContest.GenerateReport
	contest.EnableQTCs = pbContest.EnableQtcs
	contest.DetailedMultis = pbContest.DetailedMultis
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

func ContestToPB(contest model.Contest) *Contest {
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
		EnableQtcs:             contest.EnableQTCs,
		DetailedMultis:         contest.DetailedMultis,

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

func ToKeyerSettings(pbSettings *Keyer) (model.KeyerSettings, error) {
	var result model.KeyerSettings
	result.WPM = int(pbSettings.Wpm)
	result.SPLabels = pbSettings.SpLabels
	result.SPMacros = pbSettings.SpMacros
	result.RunLabels = pbSettings.RunLabels
	result.RunMacros = pbSettings.RunMacros
	result.ParrotIntervalSeconds = int(pbSettings.ParrotIntervalSeconds)
	return result, nil
}

func KeyerSettingsToPB(settings model.KeyerSettings) *Keyer {
	return &Keyer{
		Wpm:                   int32(settings.WPM),
		SpLabels:              settings.SPLabels,
		SpMacros:              settings.SPMacros,
		RunLabels:             settings.RunLabels,
		RunMacros:             settings.RunMacros,
		ParrotIntervalSeconds: int32(settings.ParrotIntervalSeconds),
	}
}

func ToQTC(pbQTC *QTC) (model.QTC, error) {
	var qtc model.QTC
	var err error

	qtc.Timestamp = time.Unix(pbQTC.Timestamp, 0)
	if qtc.Timestamp.IsZero() {
		return model.QTC{}, fmt.Errorf("QTC has no timestamp: %v", pbQTC)
	}
	qtc.Frequency = model.Frequency(pbQTC.Frequency)
	qtc.Band, err = model.ParseBand(pbQTC.Band)
	if err != nil {
		return model.QTC{}, err
	}
	qtc.Mode, err = model.ParseMode(pbQTC.Mode)
	if err != nil {
		return model.QTC{}, err
	}

	qtc.Kind = model.QTCKind(pbQTC.Kind)
	qtc.QSONumber = model.QSONumber(pbQTC.QsoNumber)
	qtc.TheirCallsign, err = model.ParseCallsign(pbQTC.TheirCallsign)
	if err != nil {
		return model.QTC{}, err
	}

	if pbQTC.Header != "" {
		qtc.Header, err = model.ParseQTCHeader(pbQTC.Header)
		if err != nil {
			return model.QTC{}, err
		}
	}

	if pbQTC.QtcTime != "" {
		qtc.QTCTime, err = model.ParseQTCTime(pbQTC.QtcTime, model.ZeroQTCTime)
		if err != nil {
			return model.QTC{}, err
		}
	}

	qtc.QTCCallsign, err = model.ParseCallsign(pbQTC.QtcCallsign)
	if err != nil {
		return model.QTC{}, err
	}

	qtc.QTCNumber = model.QSONumber(pbQTC.QtcNumber)

	return qtc, nil
}

func QTCToPB(qtc model.QTC) *QTC {
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
