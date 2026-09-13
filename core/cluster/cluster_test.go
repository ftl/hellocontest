package cluster

import (
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/ftl/hamradio/bandplan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ftl/hellocontest/core"
)

func TestParseSNR(t *testing.T) {
	tt := []struct {
		desc string
		text string
		want float64
	}{
		{desc: "a skimmer comment", text: "CW 25 dB 24 WPM CQ", want: 25},
		{desc: "without a space", text: "CW 25dB 24 WPM CQ", want: 25},
		{desc: "a negative value", text: "CW -3 dB 24 WPM CQ", want: -3},
		{desc: "a lowercase unit", text: "CW 25 db 24 WPM CQ", want: 25},
		{desc: "without a value", text: "CQ DX", want: 0},
		{desc: "an empty comment", text: "", want: 0},
		{desc: "the WPM is no SNR", text: "24 WPM", want: 0},
	}
	for _, tc := range tt {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.want, parseSNR(tc.text))
		})
	}
}

func TestCluster_Enable_WithoutARetryInterval(t *testing.T) {
	opener := &clusterOpenerStub{failures: 1}
	cluster := setupTestCluster(opener, 0)

	require.NoError(t, cluster.Enable())

	assert.Eventually(t, func() bool { return opener.Attempts() == 1 }, time.Second, time.Millisecond)
	assert.Never(t, func() bool { return opener.Attempts() > 1 }, 50*time.Millisecond, 10*time.Millisecond,
		"without a retry interval the source connects one time")
	assert.False(t, cluster.Active())
}

func TestCluster_Enable_RetriesUntilItConnects(t *testing.T) {
	opener := &clusterOpenerStub{failures: 2}
	cluster := setupTestCluster(opener, time.Millisecond)

	require.NoError(t, cluster.Enable())

	assert.Eventually(t, cluster.Active, time.Second, time.Millisecond, "the third attempt connects")
	assert.Equal(t, 3, opener.Attempts())
}

func TestCluster_Enable_StopsTheRetryOnDisable(t *testing.T) {
	opener := &clusterOpenerStub{failures: 1000}
	cluster := setupTestCluster(opener, time.Millisecond)
	require.NoError(t, cluster.Enable())
	assert.Eventually(t, func() bool { return opener.Attempts() > 0 }, time.Second, time.Millisecond)

	require.NoError(t, cluster.Disable())

	attempts := opener.Attempts()
	assert.Eventually(t, func() bool { return opener.Attempts() <= attempts+1 }, time.Second, 10*time.Millisecond,
		"the disabled source stops after the running attempt")
}

func TestCluster_ConnectionLost_Reconnects(t *testing.T) {
	opener := &clusterOpenerStub{}
	cluster := setupTestCluster(opener, time.Millisecond)
	require.NoError(t, cluster.Enable())
	require.Eventually(t, cluster.Active, time.Second, time.Millisecond)

	cluster.Connected(false)

	assert.Eventually(t, func() bool { return opener.Attempts() == 2 }, time.Second, time.Millisecond,
		"the lost connection needs no manual enable")
	assert.Eventually(t, cluster.Active, time.Second, time.Millisecond)
}

func TestCluster_ConnectionLost_WithoutARetryInterval(t *testing.T) {
	opener := &clusterOpenerStub{}
	cluster := setupTestCluster(opener, 0)
	require.NoError(t, cluster.Enable())
	require.Eventually(t, cluster.Active, time.Second, time.Millisecond)

	cluster.Connected(false)

	assert.Eventually(t, func() bool { return !cluster.Active() }, time.Second, time.Millisecond)
	assert.Never(t, func() bool { return opener.Attempts() > 1 }, 50*time.Millisecond, 10*time.Millisecond,
		"the source waits for the user")
}

func setupTestCluster(opener *clusterOpenerStub, retryInterval time.Duration) *cluster {
	parent := new(Clusters)
	source := core.SpotSource{Name: "test", HostAddress: "localhost:7373"}
	result := newCluster(parent, source, new(bandmapSpy), bandplan.IARURegion1, staticTestClock{}, opener.open)
	result.retryInterval = retryInterval
	return result
}

type clusterOpenerStub struct {
	lock     sync.Mutex
	attempts int
	failures int
}

func (o *clusterOpenerStub) open(host *net.TCPAddr, username string, password string, trace bool) (clusterClient, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	o.attempts++
	if o.attempts <= o.failures {
		return nil, fmt.Errorf("no connection")
	}
	return new(clusterClientStub), nil
}

func (o *clusterOpenerStub) Attempts() int {
	o.lock.Lock()
	defer o.lock.Unlock()
	return o.attempts
}

type clusterClientStub struct {
	disconnected bool
}

func (c *clusterClientStub) Connected() bool { return !c.disconnected }
func (c *clusterClientStub) Notify(any)      {}
func (c *clusterClientStub) Disconnect()     { c.disconnected = true }

type staticTestClock struct{}

func (staticTestClock) Now() time.Time { return testSpotTime }

func TestClusters_UpdatesTheViewOnTheUIThread(t *testing.T) {
	var pending []func()
	view := new(spotSourceViewSpy)
	clusters := NewClusters(
		[]core.SpotSource{{Name: "test", HostAddress: "localhost:7373"}},
		new(bandmapSpy), bandplan.IARURegion1, nil, staticTestClock{}, time.Minute,
		func(f func()) { pending = append(pending, f) },
	)
	clusters.SetView(view)
	require.Empty(t, view.entries, "the view belongs to the thread of the user interface")
	runPending(&pending)
	require.Equal(t, []string{"test"}, view.entries)

	// a cluster reports its connection from its own goroutine
	clusters.clusterConnected("test", true)

	assert.Empty(t, view.enabled, "the view waits for its thread")
	runPending(&pending)
	assert.Equal(t, map[string]bool{"test": true}, view.enabled)
}

func runPending(pending *[]func()) {
	functions := *pending
	*pending = nil
	for _, f := range functions {
		f()
	}
}

type spotSourceViewSpy struct {
	entries []string
	enabled map[string]bool
}

func (v *spotSourceViewSpy) AddSpotSourceEntry(name string) {
	v.entries = append(v.entries, name)
}

func (v *spotSourceViewSpy) SetSpotSourceEnabled(name string, enabled bool) {
	if v.enabled == nil {
		v.enabled = make(map[string]bool)
	}
	v.enabled[name] = enabled
}
