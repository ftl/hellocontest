package cfg

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ftl/hamradio/cfg"
	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/pb"
)

const Filename = "hellocontest.json"

var Default = Data{
	LogDirectory: "$HOME/",
	Station: pb.Station{
		Callsign: "DL0ABC",
		Operator: "DL1ABC",
		Locator:  "AA00zz",
	},
	Contest: pb.Contest{
		Name:       "Default",
		QsosGoal:   48,
		PointsGoal: 60,
		MultisGoal: 12,
	},
	Keyer: pb.Keyer{
		Wpm: 25,
		SpMacros: []string{
			"{{.MyCall}}",
			"rr {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
			"tu gl",
			"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
		},
		RunMacros: []string{
			"cq {{.MyCall}} test",
			"{{.TheirCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
			"tu {{.MyCall}} test",
			"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
		},
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
			RunMacros: []string{
				"cq {{.MyCall}} test",
				"{{.TheirCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}}",
				"tu {{.MyCall}} test",
				"nr {{.MyNumber}} {{.MyXchange}} {{.MyNumber}} {{.MyXchange}}",
			},
		},
	},
	KeyerType:     "tci",
	KeyerHost:     "localhost",
	KeyerPort:     6789,
	HamlibAddress: "localhost:4532",
	TCIAddress:    "localhost:40001",
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
	data.KeyerType = core.KeyerType(strings.ToLower(strings.TrimSpace(string(data.KeyerType))))
	return &LoadedConfiguration{
		data: data,
	}, nil
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
	Station       pb.Station         `json:"station"`
	Contest       pb.Contest         `json:"contest"`
	Keyer         pb.Keyer           `json:"keyer"`
	KeyerPresets  []core.KeyerPreset `json:"keyer_presets"`
	KeyerType     core.KeyerType     `json:"keyer_type"`
	KeyerHost     string             `json:"keyer_host"`
	KeyerPort     int                `json:"keyer_port"`
	HamlibAddress string             `json:"hamlib_address"`
	TCIAddress    string             `json:"tci_address"`
	SpotSources   []core.SpotSource  `json:"spot_sources"`
}

type LoadedConfiguration struct {
	data Data
}

func (c *LoadedConfiguration) LogDirectory() string {
	return os.ExpandEnv(c.data.LogDirectory)
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

func (c *LoadedConfiguration) Keyer() core.Keyer {
	result, _ := pb.ToKeyer(c.data.Keyer)
	return result
}

func (c *LoadedConfiguration) KeyerPresets() []core.KeyerPreset {
	return c.data.KeyerPresets
}

func (c *LoadedConfiguration) KeyerType() core.KeyerType {
	return c.data.KeyerType
}

func (c *LoadedConfiguration) KeyerHost() string {
	return c.data.KeyerHost
}

func (c *LoadedConfiguration) KeyerPort() int {
	return c.data.KeyerPort
}

func (c *LoadedConfiguration) HamlibAddress() string {
	return c.data.HamlibAddress
}

func (c *LoadedConfiguration) TCIAddress() string {
	return c.data.TCIAddress
}

func (c *LoadedConfiguration) SpotSources() []core.SpotSource {
	return c.data.SpotSources
}
