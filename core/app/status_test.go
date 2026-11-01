package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core"
)

type serviceStatusListenerFunc func(core.Service, bool)

func (f serviceStatusListenerFunc) StatusChanged(service core.Service, available bool) {
	f(service, available)
}

func TestStatusChangedUpdatesTheStatusOnTheMainThread(t *testing.T) {
	var queued []func()
	status := newServiceStatus(func(f func()) {
		queued = append(queued, f)
	})
	notified := false
	status.Notify(serviceStatusListenerFunc(func(core.Service, bool) {
		notified = true
	}))

	status.StatusChanged(core.DXCCService, true)

	assert.Empty(t, status.status, "the status must not be modified by the calling goroutine")
	assert.False(t, notified, "the listener must not be called by the calling goroutine")

	for _, f := range queued {
		f()
	}

	assert.Equal(t, map[core.Service]bool{core.DXCCService: true}, status.status)
	assert.True(t, notified)
}

func TestNotifyReportsTheCurrentStatus(t *testing.T) {
	status := newServiceStatus(nil)
	status.StatusChanged(core.DXCCService, true)

	var reported []core.Service
	status.Notify(serviceStatusListenerFunc(func(service core.Service, available bool) {
		reported = append(reported, service)
	}))

	assert.Equal(t, []core.Service{core.DXCCService}, reported)
}
