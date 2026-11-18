package app

import (
	"github.com/ftl/hellocontest/core/model"
)

func newServiceStatus(asyncRunner model.AsyncRunner) *ServiceStatus {
	if asyncRunner == nil {
		asyncRunner = func(f func()) { f() }
	}
	return &ServiceStatus{
		asyncRunner: asyncRunner,
		status:      make(map[model.Service]bool),
	}
}

type ServiceStatus struct {
	asyncRunner model.AsyncRunner
	status      map[model.Service]bool
	listeners   []interface{}
}

func (s *ServiceStatus) Notify(listener interface{}) {
	s.listeners = append(s.listeners, listener)
	if serviceStatusListener, ok := listener.(model.ServiceStatusListener); ok {
		for service, available := range s.status {
			serviceStatusListener.StatusChanged(service, available)
		}
	}
}

func (s *ServiceStatus) StatusChanged(service model.Service, available bool) {
	// the services report their status from their own goroutines, but the status and the
	// listeners belong to the main thread, see docs/concurrency_guide.md
	s.asyncRunner(func() {
		s.status[service] = available
		for _, listener := range s.listeners {
			if serviceStatusListener, ok := listener.(model.ServiceStatusListener); ok {
				serviceStatusListener.StatusChanged(service, available)
			}
		}
	})
}
