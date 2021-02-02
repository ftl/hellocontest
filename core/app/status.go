package app

import (
	"github.com/ftl/hellocontest/core"
)

func newServiceStatus(asyncRunner core.AsyncRunner) *ServiceStatus {
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
	s.status[service] = available
	for _, listener := range s.listeners {
		if serviceStatusListener, ok := listener.(core.ServiceStatusListener); ok {
			s.asyncRunner(func() {
				serviceStatusListener.StatusChanged(service, available)
			})
		}
	}
}
