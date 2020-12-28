package app

import (
	"github.com/ftl/hellocontest/core"
)

func newServiceStatus() *ServiceStatus {
	return &ServiceStatus{
		status: make(map[core.Service]bool),
	}
}

type ServiceStatus struct {
	status    map[core.Service]bool
	listeners []interface{}
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
			serviceStatusListener.StatusChanged(service, available)
		}
	}
}
