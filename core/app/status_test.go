package app

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ftl/hellocontest/core/model"
)

type serviceStatusListenerFunc func(model.Service, bool)

func (f serviceStatusListenerFunc) StatusChanged(service model.Service, available bool) {
	f(service, available)
}

func TestStatusChangedUpdatesTheStatusOnTheMainThread(t *testing.T) {
	var queued []func()
	status := newServiceStatus(func(f func()) {
		queued = append(queued, f)
	})
	notified := false
	status.Notify(serviceStatusListenerFunc(func(model.Service, bool) {
		notified = true
	}))

	status.StatusChanged(model.DXCCService, true)

	assert.Empty(t, status.status, "the status must not be modified by the calling goroutine")
	assert.False(t, notified, "the listener must not be called by the calling goroutine")

	for _, f := range queued {
		f()
	}

	assert.Equal(t, map[model.Service]bool{model.DXCCService: true}, status.status)
	assert.True(t, notified)
}

func TestNotifyReportsTheCurrentStatus(t *testing.T) {
	status := newServiceStatus(nil)
	status.StatusChanged(model.DXCCService, true)

	var reported []model.Service
	status.Notify(serviceStatusListenerFunc(func(service model.Service, available bool) {
		reported = append(reported, service)
	}))

	assert.Equal(t, []model.Service{model.DXCCService}, reported)
}
