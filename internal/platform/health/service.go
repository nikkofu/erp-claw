package health

import "sync"

type Service struct {
	mu        sync.RWMutex
	liveness  bool
	readiness bool
}

func NewService() *Service {
	return &Service{
		liveness:  true,
		readiness: true,
	}
}

func (s *Service) Liveness() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.liveness
}

func (s *Service) Readiness() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.readiness
}

func (s *Service) SetLiveness(l bool) {
	s.mu.Lock()
	s.liveness = l
	s.mu.Unlock()
}

func (s *Service) SetReadiness(r bool) {
	s.mu.Lock()
	s.readiness = r
	s.mu.Unlock()
}
