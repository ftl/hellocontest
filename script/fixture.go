package script

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ftl/conval"
	"github.com/ftl/hamradio/locator"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/store"
)

//go:embed screenshots_callhistory.txt
var callHistoryFixture string

const (
	fixtureContestIdentifier = "cw-ops"
	fixtureContestName       = "CWT Screenshot Demo"
	fixtureStationCallsign   = "DL0ABC"
	fixtureStationOperator   = "DL1ABC"
	fixtureStationLocator    = "AA00xx"
	fixtureOperatorName      = "Walter"
	fixtureOperatorNumber    = "DL"

	// the fixture QSOs and the script clock share this base, otherwise the rate view and
	// the score graph look back at an hour without any contact
	fixtureTimebase = "2023-06-28T19:00:00Z"
	// the clock keeps running while the screenshots are taken, therefore the demo QSOs end
	// well before the contest does: at the end of the contest the score graph draws no bar
	// for the current time frame, and the rate counter stops counting the time since the
	// last QSO
	fixtureQSOShift = 19 * time.Minute
	// seconds after the last demo QSO, so that the rate view shows the time since then
	fixtureNow = "2023-06-28T19:40:30Z"

	fixtureReport = "599"

	// without a goal the rate indicator collapses every axis onto its centre, and the
	// score graph scales its bins to zero
	fixtureQSOsGoal   = 40
	fixturePointsGoal = 40
	fixtureMultisGoal = 25
)

func fixtureDir() string {
	return filepath.Join(os.TempDir(), "hellocontest-screenshots")
}

// prepareScreenshotsFixture writes the log file and the call history that the screenshots
// run works on, so that the application starts on a contest that is ready to photograph.
func prepareScreenshotsFixture() (string, error) {
	// a missing target folder would only show up as a failed save of the first screenshot
	folder := screenshotsFolder()
	if info, err := os.Stat(folder); err != nil || !info.IsDir() {
		return "", fmt.Errorf("the screenshots folder %s does not exist, set %s", folder, screenshotsFolderEnv)
	}

	dir := fixtureDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create the fixture directory %s: %w", dir, err)
	}

	callHistoryFilename := filepath.Join(dir, "callhistory.txt")
	if err := os.WriteFile(callHistoryFilename, []byte(callHistoryFixture), 0644); err != nil {
		return "", fmt.Errorf("cannot write the fixture call history: %w", err)
	}

	station, err := fixtureStation()
	if err != nil {
		return "", err
	}
	contest, err := fixtureContest(callHistoryFilename)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, "screenshots.log")
	fileStore := store.NewFileStore(filename)
	if err := fileStore.Clear(); err != nil {
		return "", fmt.Errorf("cannot create the fixture log %s: %w", filename, err)
	}
	if err := fileStore.WriteStation(station); err != nil {
		return "", fmt.Errorf("cannot write the fixture station: %w", err)
	}
	if err := fileStore.WriteContest(contest); err != nil {
		return "", fmt.Errorf("cannot write the fixture contest: %w", err)
	}
	for _, qso := range fixtureQSOs() {
		if err := fileStore.WriteQSO(qso); err != nil {
			return "", fmt.Errorf("cannot write the fixture QSO %s: %w", qso.Callsign, err)
		}
	}

	return filename, nil
}

func fixtureStation() (core.Station, error) {
	loc, err := locator.Parse(fixtureStationLocator)
	if err != nil {
		return core.Station{}, fmt.Errorf("cannot parse the fixture locator: %w", err)
	}
	return core.Station{
		Callsign: core.MustParseCallsign(fixtureStationCallsign),
		Operator: core.MustParseCallsign(fixtureStationOperator),
		Locator:  loc,
	}, nil
}

func fixtureContest(callHistoryFilename string) (core.Contest, error) {
	definition, err := conval.IncludedDefinition(fixtureContestIdentifier)
	if err != nil {
		return core.Contest{}, fmt.Errorf("cannot load the contest definition %s: %w", fixtureContestIdentifier, err)
	}

	// the exchange fields are derived from the definition when the log is loaded, they do
	// not belong into the file
	return core.Contest{
		Definition:            definition,
		Name:                  fixtureContestName,
		StartTime:             mustParseFixtureTime(fixtureTimebase),
		ExchangeValues:        []string{fixtureOperatorName, fixtureOperatorNumber},
		CallHistoryFilename:   callHistoryFilename,
		CallHistoryFieldNames: []string{"Name", "Exch1"},
		QSOsGoal:              fixtureQSOsGoal,
		PointsGoal:            fixturePointsGoal,
		MultisGoal:            fixtureMultisGoal,
	}, nil
}

func mustParseFixtureTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func fixtureQSOs() []core.QSO {
	base := mustParseFixtureTime(fixtureTimebase)
	data := parseQSOCSV()
	result := make([]core.QSO, 0, len(data))
	for i, qso := range data {
		timestamp := base.Add(time.Duration(qso.minute)*time.Minute - fixtureQSOShift)
		if len(qso.values) < 3 {
			continue
		}
		result = append(result, core.QSO{
			Callsign:      core.MustParseCallsign(qso.values[0]),
			Time:          timestamp,
			LogTimestamp:  timestamp,
			Frequency:     qso.frequency,
			Band:          bandOfFrequency(qso.frequency),
			Mode:          core.ModeCW,
			MyReport:      core.RST(fixtureReport),
			MyNumber:      core.QSONumber(i + 1),
			MyExchange:    []string{fixtureOperatorName, fixtureOperatorNumber},
			TheirReport:   core.RST(fixtureReport),
			TheirNumber:   core.QSONumber(i + 1),
			TheirExchange: []string{qso.values[1], qso.values[2]},
			Workmode:      qso.workmode,
		})
	}
	return result
}

// bandOfFrequency covers only the bands that the fixture QSOs use
func bandOfFrequency(frequency core.Frequency) core.Band {
	switch {
	case frequency >= 3500000 && frequency < 4000000:
		return core.Band80m
	case frequency >= 7000000 && frequency < 7300000:
		return core.Band40m
	default:
		return core.NoBand
	}
}
