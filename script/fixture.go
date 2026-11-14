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

//go:embed screenshots_callhistory_cwt.txt
var cwtCallHistory string

//go:embed screenshots_callhistory_waedc.txt
var waedcCallHistory string

//go:embed screenshots_qsos_cwt.csv
var cwtQSOData string

//go:embed screenshots_qsos_waedc.csv
var waedcQSOData string

// a contestFixture describes the contest that one screenshots run works with. Two runs
// produce the screenshots of both contests, because a contest with a predictable exchange
// shows features that a contest with a serial number does not, and the other way round.
type contestFixture struct {
	identifier            string
	name                  string
	nameSuffix            string
	exchangeValues        []string
	generateReport        bool
	generateSerial        bool
	enableQTCs            bool
	callHistory           string
	callHistoryFieldNames []string
	qsoData               string
}

var cwtFixture = &contestFixture{
	identifier:            "CW-OPS",
	name:                  "CWT Screenshot Demo",
	nameSuffix:            "cwt",
	exchangeValues:        []string{"Walter", "DL"},
	callHistory:           cwtCallHistory,
	callHistoryFieldNames: []string{"Name", "Exch1"},
	qsoData:               cwtQSOData,
}

// the WAE DX Contest is the contest with QTCs; a DL station works DX only, and the
// application generates both the report and the serial number of the exchange
var waedcFixture = &contestFixture{
	identifier:            "DARC-WAEDC-CW",
	name:                  "WAEDC Screenshot Demo",
	nameSuffix:            "waedc",
	exchangeValues:        []string{"", ""},
	generateReport:        true,
	generateSerial:        true,
	enableQTCs:            true,
	callHistory:           waedcCallHistory,
	callHistoryFieldNames: []string{"", ""},
	qsoData:               waedcQSOData,
}

const (
	fixtureStationCallsign = "DL0ABC"
	fixtureStationOperator = "DL1ABC"
	fixtureStationLocator  = "AA00xx"

	// the contest starts here, the offsets of the demo QSOs count from this moment
	fixtureTimebase = "2023-06-28T19:00:00Z"
	// the run starts shortly after the last demo QSO, so that the rate view shows the time
	// since then and the score graph draws the bar of the current time frame
	fixtureNowAfterLastQSO = 30 * time.Second

	// without a goal the rate indicator collapses every axis onto its centre, and the
	// score graph scales its bins to zero
	fixtureQSOsGoal   = 40
	fixturePointsGoal = 40
	fixtureMultisGoal = 25
)

func contestStart() time.Time {
	return mustParseFixtureTime(fixtureTimebase)
}

// now is the moment the run starts at: shortly after the last demo QSO, so that the demo
// QSOs are in the past and the contest is still running while the screenshots are taken
func (f *contestFixture) now() time.Time {
	var last time.Duration
	for _, qso := range parseQSOCSV(f.qsoData) {
		if qso.offset > last {
			last = qso.offset
		}
	}
	return contestStart().Add(last + fixtureNowAfterLastQSO)
}

func fixtureDir() string {
	return filepath.Join(os.TempDir(), "hellocontest-screenshots")
}

// prepareScreenshotsFixture writes the log file and the call history that the screenshots
// run works on, so that the application starts on a contest that is ready to photograph.
func prepareScreenshotsFixture(fixture *contestFixture) (string, error) {
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
	if err := os.WriteFile(callHistoryFilename, []byte(fixture.callHistory), 0644); err != nil {
		return "", fmt.Errorf("cannot write the fixture call history: %w", err)
	}

	station, err := fixtureStation()
	if err != nil {
		return "", err
	}
	contest, err := fixtureContest(fixture, callHistoryFilename)
	if err != nil {
		return "", err
	}

	filename := filepath.Join(dir, "screenshots-"+fixture.nameSuffix+".log")
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

func fixtureContest(fixture *contestFixture, callHistoryFilename string) (core.Contest, error) {
	definition, err := conval.IncludedDefinition(fixture.identifier)
	if err != nil {
		return core.Contest{}, fmt.Errorf("cannot load the contest definition %s: %w", fixture.identifier, err)
	}

	// the exchange fields are derived from the definition when the log is loaded, they do
	// not belong into the file
	return core.Contest{
		Definition:             definition,
		Name:                   fixture.name,
		StartTime:              mustParseFixtureTime(fixtureTimebase),
		ExchangeValues:         fixture.exchangeValues,
		GenerateReport:         fixture.generateReport,
		GenerateSerialExchange: fixture.generateSerial,
		// the settings enable QTCs only when a contest is selected in the dialog, not
		// when a log file is loaded
		EnableQTCs:            fixture.enableQTCs,
		CallHistoryFilename:   callHistoryFilename,
		CallHistoryFieldNames: fixture.callHistoryFieldNames,
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
