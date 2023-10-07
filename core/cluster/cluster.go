package cluster

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/ftl/clusterix"
	"github.com/ftl/hamradio"
	"github.com/ftl/hamradio/bandplan"
	"github.com/ftl/hamradio/callsign"
	"github.com/ftl/hamradio/dxcc"

	"github.com/ftl/hellocontest/core"
	"github.com/ftl/hellocontest/core/network"
)

const traceClusterix = false

type Bandmap interface {
	Add(core.Spot)
}

type DXCCFinder interface {
	Find(string) (dxcc.Prefix, bool)
}

type View interface {
	AddSpotSourceEntry(name string)
	SetSpotSourceEnabled(name string, enabled bool)
}

type Clusters struct {
	clusters []*cluster
	bandmap  Bandmap
	bandplan bandplan.Bandplan
	entities DXCCFinder

	view          View
	ignoreUpdates bool

	valid       bool
	myCallsign  callsign.Callsign
	myCountry   string
	myContinent string
}

func NewClusters(sources []core.SpotSource, bandmap Bandmap, bandplan bandplan.Bandplan, entities DXCCFinder) *Clusters {
	result := &Clusters{
		clusters: make([]*cluster, 0, len(sources)),
		bandmap:  bandmap,
		bandplan: bandplan,
		entities: entities,
	}

	for _, spotSource := range sources {
		var cluster *cluster
		if isDemoSource(spotSource.Name) {
			cluster = newDemoCluster(result, spotSource, bandmap, bandplan)
		} else {
			cluster = newCluster(result, spotSource, bandmap, bandplan)
		}
		result.clusters = append(result.clusters, cluster)
	}

	return result
}

func isDemoSource(name string) bool {
	return strings.ToLower(strings.TrimSpace(name)) == "demo"
}

func (c *Clusters) cluster(name string) *cluster {
	for _, cluster := range c.clusters {
		if cluster.source.Name == name {
			return cluster
		}
	}
	return nil
}

func (c *Clusters) Valid() bool {
	return c.valid
}

func (c *Clusters) StationChanged(station core.Station) {
	c.myCallsign = station.Callsign
	entity, ok := c.entities.Find(c.myCallsign.String())
	if !ok {
		c.myCountry = ""
		c.myContinent = ""
		c.valid = false
		return
	}
	c.myCountry = entity.PrimaryPrefix
	c.myContinent = entity.Continent
	c.valid = true
}

func (c *Clusters) SetView(view View) {
	c.view = view
	c.doIgnoreUpdates(func() {
		for _, cluster := range c.clusters {
			view.AddSpotSourceEntry(cluster.source.Name)
		}
	})
}

func (c *Clusters) doIgnoreUpdates(f func()) {
	c.ignoreUpdates = true
	defer func() {
		c.ignoreUpdates = false
	}()
	f()
}

func (c *Clusters) SetSpotSourceEnabled(name string, enabled bool) {
	if c.ignoreUpdates {
		return
	}

	cluster := c.cluster(name)
	if cluster == nil {
		log.Printf("Cluster %s not found", name)
		return
	}

	var err error
	if enabled {
		err = cluster.Enable()
		if err != nil {
			log.Printf("Cannot enable cluster %s: %v", name, err)
		}
	} else {
		err = cluster.Disable()
		if err != nil {
			log.Printf("Cannot disable cluster %s: %v", name, err)
		}
	}

	if err != nil && c.view != nil {
		c.doIgnoreUpdates(func() {
			c.view.SetSpotSourceEnabled(name, cluster.Active())
		})
	}
}

func (c *Clusters) clusterConnected(name string, connected bool) {
	var status string
	if connected {
		status = "connected"
	} else {
		status = "disconnected"
	}
	log.Printf("Cluster %s %s", name, status)

	if !connected && c.view != nil {
		c.doIgnoreUpdates(func() {
			c.view.SetSpotSourceEnabled(name, false)
		})
	}
}

type clusterClient interface {
	Connected() bool
	Notify(any)
	Disconnect()
}

type openClusterFunc func(host *net.TCPAddr, username string, password string, trace bool) (clusterClient, error)

func openClusterix(host *net.TCPAddr, username string, password string, trace bool) (clusterClient, error) {
	return clusterix.Open(host, username, password, trace)
}

type cluster struct {
	parent   *Clusters
	source   core.SpotSource
	bandmap  Bandmap
	bandplan bandplan.Bandplan

	client      clusterClient
	openCluster openClusterFunc
}

func newCluster(parent *Clusters, source core.SpotSource, bandmap Bandmap, bandplan bandplan.Bandplan) *cluster {
	return &cluster{
		parent:   parent,
		source:   source,
		bandmap:  bandmap,
		bandplan: bandplan,

		openCluster: openClusterix,
	}
}

func (c *cluster) Active() bool {
	return c.client != nil && c.client.Connected()
}

func (c *cluster) Enable() error {
	if c.client != nil {
		return nil
	}

	hostAddress, err := network.ParseTCPAddr(c.source.HostAddress)
	if err != nil {
		return err
	}

	c.client, err = c.openCluster(hostAddress, c.source.Username, c.source.Password, traceClusterix)
	if err != nil {
		c.client = nil
		return err
	}

	c.client.Notify(c)

	return nil
}

func (c *cluster) Disable() error {
	if c.client == nil {
		return nil
	}
	c.client.Disconnect()
	c.client = nil

	return nil
}

func (c *cluster) Connected(connected bool) {
	if !connected {
		c.client = nil
	}
	c.parent.clusterConnected(c.source.Name, connected)
}

func (c *cluster) DX(msg clusterix.DXMessage) {
	if !c.isSpotterRelevant(msg.Spotter) {
		return
	}

	spot := core.Spot{
		Call:      msg.Call,
		Frequency: core.Frequency(msg.Frequency),
		Band:      toCoreBand(c.parent.bandplan.ByFrequency(hamradio.Frequency(msg.Frequency)).Name),
		Mode:      c.inferCoreMode(msg),
		Time:      msg.Time,
		Source:    c.source.Type,
	}
	c.bandmap.Add(spot)
}

func (c *cluster) isSpotterRelevant(spotter string) bool {
	spotterCountry, spotterContinent, ok := c.findSpotterRegion(spotter)
	if !ok {
		return true // better safe than sorry, for now
	}

	switch c.source.Filter {
	case core.AllSpots:
		return true
	case core.OwnContinentSpotsOnly:
		return spotterContinent == c.parent.myContinent
	case core.OwnCountrySpotsOnly:
		return spotterCountry == c.parent.myCountry
	default:
		return false
	}
}

func (c *cluster) findSpotterRegion(spotter string) (string, string, bool) {
	dashIndex := strings.Index(spotter, "-")
	var spotterCallString string
	if dashIndex == -1 {
		spotterCallString = spotter
	} else {
		spotterCallString = spotter[:dashIndex]
	}
	spotterCall, err := callsign.Parse(spotterCallString)
	if err != nil {
		return "", "", false
	}
	prefix, ok := c.parent.entities.Find(spotterCall.String())
	if !ok {
		return "", "", false
	}
	return prefix.PrimaryPrefix, prefix.Continent, true
}

func (c *cluster) inferCoreMode(msg clusterix.DXMessage) core.Mode {
	text := strings.ToLower(strings.TrimSpace(msg.Text))
	switch {
	case strings.Contains(text, "cw"):
		return core.ModeCW
	case strings.Contains(text, "rtty"):
		return core.ModeRTTY
	case strings.Contains(text, "psk"):
		return core.ModeDigital
	case strings.Contains(text, "ft8"):
		return core.ModeDigital
	case strings.Contains(text, "ft4"):
		return core.ModeDigital
	case strings.Contains(text, "jt9"):
		return core.ModeDigital
	case strings.Contains(text, "jt65"):
		return core.ModeDigital
	default:
		return c.modeFromFrequency(core.Frequency(msg.Frequency))
	}
}

func (c *cluster) modeFromFrequency(f core.Frequency) core.Mode {
	frequency := hamradio.Frequency(f)
	band := c.bandplan.ByFrequency(frequency)
	if band.Name == bandplan.BandUnknown {
		return core.NoMode
	}

	for _, portion := range band.Portions {
		if portion.Contains(frequency) {
			return core.Mode(portion.Mode)
		}
	}
	return core.NoMode
}

func toCoreBand(bandName bandplan.BandName) core.Band {
	if bandName == bandplan.BandUnknown {
		return core.NoBand
	}
	return core.Band(bandName)
}

func newDemoCluster(parent *Clusters, source core.SpotSource, bandmap Bandmap, bandplan bandplan.Bandplan) *cluster {
	return &cluster{
		parent:   parent,
		source:   source,
		bandmap:  bandmap,
		bandplan: bandplan,

		openCluster: openDemoCluster(bandmap),
	}
}

func openDemoCluster(bandmap Bandmap) openClusterFunc {
	return func(host *net.TCPAddr, username string, password string, trace bool) (clusterClient, error) {
		result := &demoCluster{
			connected: true,
			bandmap:   bandmap,
			spots: []core.Spot{
				{
					Call:      callsign.MustParse("W1WA"),
					Frequency: 7012000,
					Band:      core.Band40m,
					Mode:      core.ModeCW,
					Source:    core.ClusterSpot,
				},
				{
					Call:      callsign.MustParse("DA0BCC"),
					Frequency: 7013700,
					Band:      core.Band40m,
					Mode:      core.ModeCW,
					Source:    core.RBNSpot,
				},
				{
					Call:      callsign.MustParse("DA0HQ"),
					Frequency: 7012300,
					Band:      core.Band40m,
					Mode:      core.ModeCW,
					Source:    core.SkimmerSpot,
				},
				{
					Call:      callsign.MustParse("G0ABC"),
					Frequency: 7015000,
					Band:      core.Band40m,
					Mode:      core.ModeCW,
					Source:    core.ManualSpot,
				},
				{
					Call:      callsign.MustParse("PA0ABC"),
					Frequency: 7019000,
					Band:      core.Band40m,
					Mode:      core.ModeCW,
					Source:    core.WorkedSpot,
				},
			},
		}
		return result, nil
	}
}

type demoCluster struct {
	listeners []any
	connected bool
	bandmap   Bandmap
	spots     []core.Spot
}

func (c *demoCluster) emitConnected() {
	for _, l := range c.listeners {
		if listener, ok := l.(clusterix.ConnectionListener); ok {
			listener.Connected(c.connected)
		}
	}
}

func (c *demoCluster) emitDemoSpots() {
	now := time.Now()
	for i, spot := range c.spots {
		spot.Time = now.Add(-1 * time.Duration(i) * time.Minute)
		c.bandmap.Add(spot)
	}
}

func (c *demoCluster) Connected() bool {
	return c.connected
}

func (c *demoCluster) Notify(listener any) {
	c.listeners = append(c.listeners, listener)
	c.emitConnected()
	c.emitDemoSpots()
}

func (c *demoCluster) Disconnect() {
	c.connected = false
	c.emitConnected()
}
