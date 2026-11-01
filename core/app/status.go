package app

import (
	"github.com/ftl/hellocontest/core"
)

func newServiceStatus(asyncRunner core.AsyncRunner) *ServiceStatus {
	if asyncRunner == nil {
		asyncRunner = func(f func()) { f() }
	}
	return &ServiceStatus{
		asyncRunner: asyncRunner,
		status:      make(map[core.Service]bool),
	}
}

type ServiceStatus struct {
	asyncRunner core.AsyncRunner
	status      map[core.Service]bool
	listeners   []interface{}
}

func (s *ServiceStatus) Notify(listener interface{}) {
	s.listeners = append(s.listeners, listener)
	if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
		for service, available := range s.status {
			serviceStatusListener.StatusChanged(service, available)
		}
	}
}

func (s *ServiceStatus) StatusChanged(service core.Service, available bool) {
	// the services report their status from their own goroutines, but the status and the
	// listeners belong to the main thread, see docs/concurrency_guide.md
	s.asyncRunner(func() {
		s.status[service] = available
		for _, listener := range s.listeners {
			if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
				serviceStatusListener.StatusChanged(service, available)
			}
		}
	})
}
