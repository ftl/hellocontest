package cfg

import (
	"log"
	"path/filepath"

	"github.com/ftl/hamradio/cfg"
	"github.com/pkg/errors"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/pb"
)

const Filename = "hellocontest.json"

var Default = Data{
	Station: pb.Station{
		Callsign: "DL0ABC",
		Operator: "DL1ABC",
		Locator:  "AA00zz",
	},
	Contest: pb.Contest{
		Name:                    "Default",
		EnterTheirNumber:        true,
		EnterTheirXchange:       true,
		RequireTheirXchange:     true,
		AllowMultiBand:          true,
		AllowMultiMode:          true,
		SameCountryPoints:       1,
		SameContinentPoints:     3,
		SpecificCountryPoints:   10,
		SpecificCountryPrefixes: []string{"DL"},
		OtherPoints:             5,
		Multis: &pb.Multis{
			Dxcc:    true,
			Wpx:     true,
			Xchange: true,
		},
		XchangeMultiPattern: "\\d+",
		CountPerBand:        true,
		CabrilloQsoTemplate: "{{.QRG}} {{.Mode}} {{.Date}} {{.Time}} {{.MyCall}} {{.MyReport}} {{.MyNumber}} {{.MyXchange}} {{.TheirCall}} {{.TheirReport}} {{.TheirNumber}} {{.TheirXchange}}",
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
	HamlibAddress: "localhost:4532",
	KeyerHost:     "localhost",
	KeyerPort:     6789,
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
	Station       pb.Station
	Contest       pb.Contest
	Keyer         pb.Keyer
	KeyerHost     string `json:"keyer_host"`
	KeyerPort     int    `json:"keyer_port"`
	HamlibAddress string `json:"hamlib_address"`
}

type LoadedConfiguration struct {
	data Data
}

func (c *LoadedConfiguration) Station() core.Station {
	result, err := pb.ToStation(c.data.Station)
	if err != nil {
		log.Printf("Cannot parse default station settings: %v", err)
		return core.Station{}
	}
	return result
}

func (c *LoadedConfiguration) Keyer() core.Keyer {
	result, _ := pb.ToKeyer(c.data.Keyer)
	return result
}

func (c *LoadedConfiguration) Contest() core.Contest {
	result, _ := pb.ToContest(c.data.Contest)
	return result
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
