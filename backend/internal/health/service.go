package health

import (
	"sync"
	"time"
)

type Status struct {
	Healthy       bool     `json:"healthy"`
	UnhealthyFor  int64    `json:"unhealthyForSeconds"`
	LastCheckUTC  string   `json:"lastCheckUtc"`
	FailureReason []string `json:"failureReason"`
}

type Service struct {
	mu             sync.Mutex
	status         Status
	unhealthySince *time.Time
}

func NewService() *Service {
	now := time.Now().UTC().Format(time.RFC3339)
	return &Service{status: Status{Healthy: true, LastCheckUTC: now}}
}

func (s *Service) SetStatus(st Status) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st.Healthy {
		s.unhealthySince = nil
		st.UnhealthyFor = 0
	} else if s.unhealthySince == nil {
		now := time.Now().UTC()
		s.unhealthySince = &now
	}
	st.LastCheckUTC = time.Now().UTC().Format(time.RFC3339)
	if s.unhealthySince != nil {
		st.UnhealthyFor = int64(time.Since(*s.unhealthySince).Seconds())
	}
	s.status = st
}

func (s *Service) MarkHealthy() {
	s.SetStatus(Status{Healthy: true, FailureReason: nil})
}

func (s *Service) MarkUnhealthy(reasons ...string) {
	s.SetStatus(Status{Healthy: false, FailureReason: reasons})
}

func (s *Service) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.status
	if s.unhealthySince != nil {
		st.UnhealthyFor = int64(time.Since(*s.unhealthySince).Seconds())
	}
	st.LastCheckUTC = time.Now().UTC().Format(time.RFC3339)
	return st
}
