package cfg

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ftl/hamradio/cfg"
	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/pb"
)

const Filename = "hellocontest.json"
const DefaultSpotLifetime = 10 * time.Minute

var Default = &Data{
	LogDirectory: "$HOME/",
	HamDXMapPort: 17300,
	Station: &pb.Station{
		Callsign: "DL0ABC",
		Operator: "DL1ABC",
		Locator:  "AA00xx",
	},
	Contest: &pb.Contest{
		Name:       "Default",
		QsosGoal:   48,
		PointsGoal: 60,
		MultisGoal: 12,
	},
	Radios: []core.Radio{
		{
			Name:    "Hamlib Radio",
			Type:    "hamlib",
			Address: "localhost:4532",
			Keyer:   "local cwdaemon",
		},
		{
			Name:    "TCI Radio",
			Type:    "tci",
			Address: "localhost:40001",
			Keyer:   "radio",
		},
	},
	Keyers: []core.Keyer{
		{
			Name:    "Local CWDaemon",
			Type:    "cwdaemon",
			Address: "localhost:6789",
		},
	},
	KeyerSettings: core.KeyerSettings{
		WPM:                   25,
		Preset:                "Default",
		ParrotIntervalSeconds: 10,
	},
	KeyerPresets: []core.KeyerPreset{
		{
			Name: "Default",
			SPMacros: []string{
				"{{.MyCall}}",
				"rr {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
				"tu gl",
				"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
			},
			SPLabels: []string{"MyCall", "R Exch", "TU GL", "Exch AGN"},
			RunMacros: []string{
				"cq {{.MyCall}} test",
				"{{.TheirCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
				"tu {{.MyCall}} test",
				"{{.TheirCall}} tu",
			},
			RunLabels: []string{"CQ", "Exch", "TU QRZ?", "TU"},
		},
		{
			Name: "Gen. Serial",
			SPMacros: []string{
				"{{.MyCall}}",
				"rr {{.MyReport}} {{.MyNumber}}",
				"tu gl",
				"nr {{.MyNumber}} {{.MyNumber}}",
			},
			SPLabels: []string{"MyCall", "R Exch", "TU GL", "Exch AGN"},
			RunMacros: []string{
				"cq {{.MyCall}} test",
				"{{.TheirCall}} {{.MyReport}} {{.MyNumber}}",
				"tu  {{.MyCall}} test",
				"{{.TheirCall}} tu",
			},
			RunLabels: []string{"CQ", "Exch", "TU QRZ?", "TU"},
		},
		{
			Name: "Gen. Text",
			SPMacros: []string{
				"{{.MyCall}}",
				"rr {{.MyReport}} {{.MyXchange}}",
				"tu gl",
				"nr {{.MyXchange}} {{.MyXchange}}",
			},
			SPLabels: []string{"MyCall", "R Exch", "TU GL", "Exch AGN"},
			RunMacros: []string{
				"cq {{.MyCall}} test",
				"{{.TheirCall}} {{.MyReport}} {{.MyXchange}}",
				"tu  {{.MyCall}} test",
				"{{.TheirCall}} tu",
			},
			RunLabels: []string{"CQ", "Exch", "TU QRZ?", "TU"},
		},
		{
			Name: "CQ-WW-CW",
			SPMacros: []string{
				"{{.MyCall}}",
				"rr {{.MyReport}} {{index .MyExchanges 1 | cut}}",
				"tu gl",
				"nr {{index .MyExchanges 1 | cut}} {{index .MyExchanges 1 | cut}}",
			},
			SPLabels: []string{"MyCall", "R Exch", "TU GL", "Exch AGN"},
			RunMacros: []string{
				"cq {{.MyCall}} test",
				"{{.TheirCall}} {{.MyReport}} {{index .MyExchanges 1 | cut}}",
				"tu  {{.MyCall}} test",
				"{{.TheirCall}} tu",
			},
			RunLabels: []string{"CQ", "Exch", "TU QRZ?", "TU"},
		},
	},
	SpotLifetime: "10m",
	SpotSources: []core.SpotSource{
		{
			Name:        "Skimmer",
			Username:    "dl0abc",
			HostAddress: "localhost:7373",
			Type:        core.SkimmerSpot,
		},
		{
			Name:        "W3LPL",
			Username:    "dl0abc",
			HostAddress: "w3lpl.net:7373",
			Type:        core.RBNSpot,
			Filter:      core.OwnContinentSpotsOnly,
		},
	},
}

// Load loads the configuration from the default location (see github.com/ftl/cfg/LoadJSON()).
func Load() (*LoadedConfiguration, error) {
	if !cfg.Exists("", Filename) {
		cfg.PrepareDirectory("")
		err := cfg.SaveJSON("", Filename, Default)
		if err != nil {
			return nil, err
		}
	}

	var data Data
	err := cfg.LoadJSON("", Filename, &data)
	if err != nil {
		return nil, err
	}

	for i, radio := range data.Radios {
		radio.Type = core.RadioType(normalizeName(string(radio.Type)))
		data.Radios[i] = radio
	}
	for i, keyer := range data.Keyers {
		keyer.Type = core.KeyerType(normalizeName(string(keyer.Type)))
		data.Keyers[i] = keyer
	}

	return &LoadedConfiguration{
		data: &data,
	}, nil
}

func normalizeName(name string) string {
	return strings.TrimSpace(strings.ToLower(name))
}

// Directory returns the configuration directory. It panics if the directory could not be determined.
func Directory() string {
	dir, err := cfg.Directory("")
	if err != nil {
		panic(errors.Wrap(err, "cannot determine configuration directory"))
	}
	return dir
}

func AbsoluteFilename() string {
	return filepath.Join(Directory(), Filename)
}

type Data struct {
	LogDirectory  string             `json:"log_directory"`
	HamDXMapPort  int                `json:"ham_dx_map_port"`
	Station       *pb.Station        `json:"station"`
	Contest       *pb.Contest        `json:"contest"`
	Radios        []core.Radio       `json:"radios"`
	Keyers        []core.Keyer       `json:"keyers"`
	KeyerSettings core.KeyerSettings `json:"keyer_settings"`
	KeyerPresets  []core.KeyerPreset `json:"keyer_presets"`
	SpotLifetime  string             `json:"spot_lifetime"`
	SpotSources   []core.SpotSource  `json:"spot_sources"`
}

type LoadedConfiguration struct {
	data *Data
}

func (c *LoadedConfiguration) LogDirectory() string {
	return os.ExpandEnv(c.data.LogDirectory)
}

func (c *LoadedConfiguration) HamDXMapPort() int {
	return c.data.HamDXMapPort
}

func (c *LoadedConfiguration) Station() core.Station {
	result, err := pb.ToStation(c.data.Station)
	if err != nil {
		log.Printf("Cannot parse default station settings: %v", err)
		return core.Station{}
	}
	return result
}

func (c *LoadedConfiguration) Contest() core.Contest {
	result, _ := pb.ToContest(c.data.Contest)
	return result
}

func (c *LoadedConfiguration) Radios() []core.Radio {
	return c.data.Radios
}

func (c *LoadedConfiguration) Keyers() []core.Keyer {
	return c.data.Keyers
}

func (c *LoadedConfiguration) KeyerSettings() core.KeyerSettings {
	return c.data.KeyerSettings
}

func (c *LoadedConfiguration) KeyerPresets() []core.KeyerPreset {
	return c.data.KeyerPresets
}

func (c *LoadedConfiguration) SpotLifetime() time.Duration {
	result, err := time.ParseDuration(c.data.SpotLifetime)
	if err != nil {
		log.Printf("cannot parse spot lifetime: %v", err)
		return DefaultSpotLifetime
	}
	return result
}

func (c *LoadedConfiguration) SpotSources() []core.SpotSource {
	return c.data.SpotSources
}
